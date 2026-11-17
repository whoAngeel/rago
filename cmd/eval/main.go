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
	"syscall"

	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/eval"
	chunkerpkg "github.com/whoAngeel/rago/internal/infrastructure/chunker"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
	"github.com/whoAngeel/rago/internal/infrastructure/logger"
	"github.com/whoAngeel/rago/internal/infrastructure/openrouter"
	parserpkg "github.com/whoAngeel/rago/internal/infrastructure/parser"
	"github.com/whoAngeel/rago/internal/infrastructure/qdrant"
)

const (
	evalCollection = "rago_eval"
	evalUserID     = 987654321 // synthetic user, never a real one
	purgeUpToDocID = 64        // stale document ids cleared before a fresh ingest
)

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
	questionsPath := fs.String("questions", "testdata/eval/questions.json", "path to the questions dataset")
	dataDir := fs.String("data-dir", "testdata", "directory with the sample documents")
	fs.Parse(os.Args[1:])

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

	files, err := eval.ListSourceFiles(*dataDir)
	if err != nil {
		return err
	}
	questions, err := eval.LoadQuestions(*questionsPath)
	if err != nil {
		return err
	}
	if err := eval.ValidateQuestions(questions, files); err != nil {
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
		if err := ingest(ctx, cfg, log, store, embedder, *dataDir, files); err != nil {
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
func ingest(ctx context.Context, cfg *config.Config, log ports.Logger, store ports.VectorStore, embedder ports.Embedder, dataDir string, files []string) error {
	evalCfg := *cfg
	evalCfg.QdrantCollection = evalCollection
	uc := application.NewIngestUsecase(store, embedder, log, evalCfg)

	// Same registrations as cmd/server/main.go, limited to text-based formats.
	registry := parserpkg.NewRegistry()
	registry.Register("text/plain", parserpkg.NewPlainTextAdapter())
	registry.Register("text/markdown", parserpkg.NewPlainTextAdapter())
	registry.Register("text/csv", parserpkg.NewCSVParser())
	registry.Register("application/json", parserpkg.NewJSONParser())
	chunker := chunkerpkg.NewFixedChunker(1000, 200) // same size/overlap as cmd/server/main.go

	if err := store.CreateCollection(ctx, evalCollection, cfg.EmbeddingDim); err != nil {
		return fmt.Errorf("creating collection %s: %w", evalCollection, err)
	}
	for id := 1; id <= purgeUpToDocID; id++ {
		if err := store.DeleteByDocumentID(ctx, evalCollection, id); err != nil {
			return fmt.Errorf("purging %s: %w", evalCollection, err)
		}
	}

	noSteps := func(string) func(error) { return func(error) {} }
	for i, name := range files {
		ct, err := eval.ContentTypeFor(name)
		if err != nil {
			log.Warn("skipping unsupported file", "file", name)
			continue
		}
		content, err := os.ReadFile(filepath.Join(dataDir, name))
		if err != nil {
			return err
		}
		chunks, err := eval.BuildChunks(ctx, registry, chunker, name, content)
		if err != nil {
			return err
		}
		doc := &domain.Document{ID: i + 1, UserID: evalUserID, Filename: name, ContentType: ct}
		if err := uc.Execute(ctx, doc, nil, chunks, noSteps); err != nil {
			return fmt.Errorf("ingesting %s: %w", name, err)
		}
		fmt.Fprintf(os.Stderr, "ingested %s (%d chunks)\n", name, len(chunks))
	}
	return nil
}
