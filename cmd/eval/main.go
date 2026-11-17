// Command eval measures retrieval quality: given a question, does the right
// document come back, at what rank and with what score. It needs only Qdrant
// and the embedder; it never calls the chat LLM.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/eval"
	"github.com/whoAngeel/rago/internal/infrastructure/chunker"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
	"github.com/whoAngeel/rago/internal/infrastructure/logger"
	"github.com/whoAngeel/rago/internal/infrastructure/ocr"
	"github.com/whoAngeel/rago/internal/infrastructure/openrouter"
	"github.com/whoAngeel/rago/internal/infrastructure/parser"
	"github.com/whoAngeel/rago/internal/infrastructure/qdrant"
)

const (
	evalCollection = "rago_eval"
	evalUserID     = 987654321 // synthetic user, never a real one
	purgeUpToDocID = 64        // stale document ids cleared before a fresh ingest
)

// stringList is a repeatable string flag.
type stringList []string

func (l *stringList) String() string     { return strings.Join(*l, ",") }
func (l *stringList) Set(v string) error { *l = append(*l, v); return nil }

const dataRoot = "testdata"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "eval:", err)
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	topK := fs.Int("top-k", 0, "chunks retrieved per question (default: RETRIEVAL_TOP_K from config)")
	out := fs.String("out", "", "also write the report as JSON to this path")
	skipIngest := fs.Bool("skip-ingest", false, "reuse the data already in the "+evalCollection+" collection")
	allowRemote := fs.Bool("allow-remote", false, "allow a QDRANT_HOST that is not localhost/127.0.0.1/::1")
	var questionPaths, dataDirs stringList
	fs.Var(&questionPaths, "questions", "question file (repeatable); default: testdata/eval/*.json plus testdata/private/questions.json if present")
	fs.Var(&dataDirs, "data-dir", "directory with documents (repeatable, non-recursive); default: testdata, plus testdata/manuals and testdata/private if present")
	dump := fs.Bool("dump-chunks", false, "print the chunks ingestion would produce and exit; needs no Qdrant, embedder or API key")
	dumpSource := fs.String("source", "", "with --dump-chunks, only this document (base file name)")
	forceOCR := fs.Bool("force-ocr", false, "rasterise and OCR every PDF page instead of using its text layer")
	// The defaults below apply to --dump-chunks (it never loads config); an
	// ingest run replaces the ones not set explicitly with the server's config.
	def := chunker.DefaultSettings()
	chunkKind := fs.String("chunker", def.Kind, "chunking strategy: fixed or recursive (default: CHUNKER from config)")
	chunkSize := fs.Int("chunk-size", def.Size, "chunk size in characters (default: CHUNK_SIZE from config)")
	chunkOverlap := fs.Int("chunk-overlap", def.Overlap, "overlap between consecutive chunks in characters (default: CHUNK_OVERLAP from config)")
	ctxHeader := fs.Bool("context-header", def.Header, "prepend a [title — p. N] line to every chunk before embedding (default: CHUNK_CONTEXT_HEADER from config)")
	cacheDir := fs.String("parse-cache-dir", defaultParseCacheDir(), "directory caching parsed pages (empty disables)")
	fs.Parse(os.Args[1:])

	explicit := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { explicit[f.Name] = true })
	chunkCfg := eval.ChunkConfig{Kind: *chunkKind, Size: *chunkSize, Overlap: *chunkOverlap, Header: *ctxHeader}

	if len(dataDirs) == 0 {
		dataDirs = DefaultDirs()
	}
	sources, warnings, err := eval.CollectSources(dataDirs)
	if err != nil {
		return err
	}
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
	if *dump {
		if err := chunkCfg.Validate(); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, chunkCfg)
		return dumpChunks(sources, *dumpSource, *forceOCR, chunkCfg, *cacheDir)
	}
	if *dumpSource != "" {
		return fmt.Errorf("--source only applies with --dump-chunks")
	}
	if len(questionPaths) == 0 {
		paths, err := eval.DefaultQuestionPaths(dataRoot)
		if err != nil {
			return err
		}
		questionPaths = paths
	}

	// config.Load requires these even though the eval never uses them.
	for k, v := range map[string]string{"DATABASE_URL": "unused", "MINIO_ENDPOINT": "unused"} {
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if err := eval.CheckHost(cfg.QdrantHost, *allowRemote); err != nil {
		return err
	}
	if *topK <= 0 {
		*topK = cfg.RetrievalTopK
	}
	// Chunking follows the server's config; explicit flags override it.
	serverChunking, warn := chunker.SettingsFromConfig(*cfg)
	if warn != nil {
		fmt.Fprintln(os.Stderr, "warning: invalid chunking configuration, using defaults:", warn)
	}
	if !explicit["chunker"] {
		chunkCfg.Kind = serverChunking.Kind
	}
	if !explicit["chunk-size"] {
		chunkCfg.Size = serverChunking.Size
	}
	if !explicit["chunk-overlap"] {
		chunkCfg.Overlap = serverChunking.Overlap
	}
	if !explicit["context-header"] {
		chunkCfg.Header = serverChunking.Header
	}
	if err := chunkCfg.Validate(); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, chunkCfg)

	questions, err := eval.LoadQuestionFiles(questionPaths)
	if err != nil {
		return err
	}
	if err := eval.ValidateQuestions(questions, eval.SourceNames(sources)); err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log := logger.New(cfg.Env)
	store, err := qdrant.NewQdrantAdapter(cfg.QdrantHost, cfg.QdrantPort)
	if err != nil {
		return fmt.Errorf("qdrant: %w", err)
	}
	embedder, err := openrouter.NewEmbedderAdapter(cfg.OpenRouterKey, cfg.OpenRouterBaseUrl, cfg.EmbeddingModel)
	if err != nil {
		return fmt.Errorf("embedder: %w", err)
	}

	if !*skipIngest {
		if err := ingest(ctx, cfg, log, store, embedder, sources, *forceOCR, chunkCfg, *cacheDir); err != nil {
			return err
		}
	}

	results, err := eval.RunQuestions(ctx, embedder, store, evalCollection, evalUserID, questions, *topK)
	if err != nil {
		return err
	}
	report := eval.BuildReport(results, *topK, cfg.RetrievalScoreThreshold, eval.DefaultSweep)
	if err := report.WriteTable(os.Stdout); err != nil {
		return err
	}
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := report.WriteJSON(f); err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, "\nJSON report written to", *out)
	}
	return nil
}

// ingest rebuilds the eval collection content with the same parsers, chunker
// settings and IngestUsecase the server uses, against the dedicated collection.
func ingest(ctx context.Context, cfg *config.Config, log ports.Logger, store ports.VectorStore, embedder ports.Embedder, sources []eval.Source, forceOCR bool, chunkCfg eval.ChunkConfig, cacheDir string) error {
	evalCfg := *cfg
	evalCfg.QdrantCollection = evalCollection
	uc := application.NewIngestUsecase(store, embedder, log, evalCfg)

	// Same parsers and chunker settings as cmd/server/main.go, PDF and OCR included.
	engine := ocr.NewEngine(cfg)
	registry := eval.NewParserRegistry(engine, ocr.PDFOptions(cfg)...)
	engineName, model, mode := ocr.Effective(cfg)
	fmt.Fprintf(os.Stderr, "OCR: engine %s, PDF OCR mode %s\n", ocrLabel(cfg), mode)
	if forceOCR {
		ctx = context.WithValue(ctx, parser.ForceOCRKey, true)
		fmt.Fprintf(os.Stderr, "force-OCR on: every PDF page goes through OCR (engine: %s)\n", ocrLabel(cfg))
	}
	chunkr, err := chunkCfg.Chunker()
	if err != nil {
		return err
	}
	cache := &eval.ParseCache{Dir: cacheDir, Signature: eval.ParseSignature(engineName, model, mode, forceOCR)}

	if err := store.CreateCollection(ctx, evalCollection, cfg.EmbeddingDim); err != nil {
		return fmt.Errorf("creating collection %s: %w", evalCollection, err)
	}
	for id := 1; id <= purgeUpToDocID; id++ {
		if err := store.DeleteByDocumentID(ctx, evalCollection, id); err != nil {
			return fmt.Errorf("purging %s: %w", evalCollection, err)
		}
	}

	noSteps := func(string) func(error) { return func(error) {} }
	for i, src := range sources {
		name := src.Name
		ct, err := eval.ContentTypeFor(name)
		if err != nil {
			log.Warn("skipping unsupported file", "file", name)
			continue
		}
		content, err := eval.ReadSource(src)
		if err != nil {
			return err
		}
		docs, err := cache.Docs(ctx, registry, name, content)
		if err != nil {
			return err
		}
		chunks, err := chunker.ChunkPages(docs, name, chunkr, chunkCfg.Header)
		if err != nil {
			return err
		}
		doc := &domain.Document{ID: i + 1, UserID: evalUserID, Filename: name, ContentType: ct}
		if err := uc.Execute(ctx, doc, nil, chunks, noSteps); err != nil {
			return fmt.Errorf("ingesting %s: %w", name, err)
		}
		fmt.Fprintf(os.Stderr, "ingested %s (%d chunks)\n", name, len(chunks))
	}
	fmt.Fprintln(os.Stderr, cache.Summary())
	return nil
}

// DefaultDirs is the default document directory list.
func DefaultDirs() []string { return eval.DefaultDataDirs(dataRoot) }

// dumpChunks parses and chunks the sources as ingestion would and prints them.
func dumpChunks(sources []eval.Source, only string, forceOCR bool, chunkCfg eval.ChunkConfig, cacheDir string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	// --dump-chunks needs no config or API key, so it always uses Tesseract.
	registry := eval.NewParserRegistry(parser.NewTesseractOCR())
	chunkr, err := chunkCfg.Chunker()
	if err != nil {
		return err
	}
	cache := &eval.ParseCache{Dir: cacheDir, Signature: eval.ParseSignature("tesseract", "", "auto", forceOCR)}
	if forceOCR {
		ctx = context.WithValue(ctx, parser.ForceOCRKey, true)
		fmt.Fprintln(os.Stderr, "force-OCR on: every PDF page goes through OCR (engine: tesseract)")
	}
	var docs []eval.DocChunks
	for _, src := range sources {
		if only != "" && src.Name != only {
			continue
		}
		if _, err := eval.ContentTypeFor(src.Name); err != nil {
			fmt.Fprintln(os.Stderr, "warning: skipping unsupported file", src.Name)
			continue
		}
		content, err := eval.ReadSource(src)
		if err != nil {
			return err
		}
		parsed, err := cache.Docs(ctx, registry, src.Name, content)
		if err != nil {
			return err
		}
		chunks, err := chunker.ChunkPages(parsed, src.Name, chunkr, chunkCfg.Header)
		if err != nil {
			return err
		}
		docs = append(docs, eval.DocChunks{Name: src.Name, Chunks: chunks})
	}
	fmt.Fprintln(os.Stderr, cache.Summary())
	return eval.WriteChunkDump(os.Stdout, docs, only)
}

// ocrLabel names the active OCR engine for log output.
func ocrLabel(cfg *config.Config) string {
	if engine, model, _ := ocr.Effective(cfg); engine == "llm" {
		return "llm " + model
	}
	return "tesseract"
}

// defaultParseCacheDir is <user cache dir>/rago-eval/parse, or "" (cache
// disabled) when the platform has no user cache directory.
func defaultParseCacheDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "rago-eval", "parse")
}
