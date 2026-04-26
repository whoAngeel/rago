package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/whoAngeel/rago/internal/provider"
	"github.com/whoAngeel/rago/internal/store"
	"os"
)

type RagEngine struct {
	store    store.VectorStore
	embedder *provider.Embedder
	llm      llms.Model
}

func NewRAGEngine(store store.VectorStore, embedder *provider.Embedder, cfg *Config) (*RagEngine, error) {
	llm, err := openai.New(
		openai.WithToken(cfg.OpenRouterKey),
		openai.WithBaseURL(cfg.BaseUrl),
		openai.WithModel(cfg.Model),
	)

	if err != nil {
		return nil, fmt.Errorf("error creating llm: %w", err)
	}
	return &RagEngine{
		store:    store,
		embedder: embedder,
		llm:      llm,
	}, nil
}

func (e *RagEngine) Ask(ctx context.Context, collection string, question string) (string, error) {
	// embeding de la pregunta
	questionVector, err := e.embedder.ComputeEmbeddings(ctx, []string{question})
	if err != nil {
		return "", fmt.Errorf("error on embedding: %w", err)
	}

	// search qdrant
	docs, err := e.store.Search(ctx, collection, questionVector[0], 5)
	if err != nil {
		return "", fmt.Errorf("error on searching: %w", err)
	}

	if len(docs) == 0 {
		return "Relevant info not found", nil
	}
	// format context
	contextStr := formatContext(docs)

	// 4 Create context prompt
	prompt := buildPrompt(question, contextStr)

	// llm call
	resp, err := llms.GenerateFromSinglePrompt(ctx, e.llm, prompt)

	if err != nil {
		return "", fmt.Errorf("error on llm: %w", err)
	}
	return resp, nil
}

func (e *RagEngine) Ingest(ctx context.Context, collection string, filePath string) error {
	// 1. Leer archivo
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error leyendo archivo: %w", err)
	}

	// 2. Splitter (1000 chars, 200 overlap)
	splitter := textsplitter.NewRecursiveCharacter()
	splitter.ChunkSize = 1000
	splitter.ChunkOverlap = 200

	chunks, err := splitter.SplitText(string(data))
	if err != nil {
		return fmt.Errorf("error splitting text: %w", err)
	}

	// 3. Preparar documentos y vectores
	docs := make([]schema.Document, len(chunks))
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		docs[i] = schema.Document{
			PageContent: chunk,
			Metadata: map[string]any{
				"source": filePath,
				"chunk":  i,
			},
		}
		texts[i] = chunk
	}

	// 4. Embeddings
	vectors, err := e.embedder.ComputeEmbeddings(ctx, texts)
	if err != nil {
		return fmt.Errorf("error en embeddings: %w", err)
	}

	// 5. Asegurar colección (ej: 1536 para OpenAI)
	// Usamos la dimensión del primer vector o de config
	dim := len(vectors[0])
	err = e.store.CreateCollection(ctx, collection, dim)
	if err != nil {
		return fmt.Errorf("error creando colección: %w", err)
	}

	// 6. Upsert
	err = e.store.UpsertDocuments(ctx, collection, docs, vectors)
	if err != nil {
		return fmt.Errorf("error upserting: %w", err)
	}

	return nil
}

func formatContext(docs []schema.Document) string {
	var b strings.Builder
	for i, doc := range docs {
		b.WriteString(fmt.Sprintf("[%d] %s\n", i+1, doc.PageContent))
	}
	return b.String()
}

func buildPrompt(question, context string) string {
	return fmt.Sprintf(`Usa el siguiente context para responder la pregunta.
	Context: 
	%s
	
	Pregunta: %s
	Respuesta: `, context, question)
}
