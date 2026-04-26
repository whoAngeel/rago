package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/whoAngeel/rago/internal/engine"
	"github.com/whoAngeel/rago/internal/provider"
	"github.com/whoAngeel/rago/internal/store"
)

func main() {
	ctx := context.Background()

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	// Cargar configuración
	cfg, err := engine.LoadConfig()
	if err != nil {
		log.Fatalf("Error config: %v", err)
	}

	// Inicializar Store
	vStore, err := store.NewQdrantStore(cfg.QdrantHost, cfg.QdrantPort)
	if err != nil {
		log.Fatalf("Error store: %v", err)
	}

	// Inicializar Embedder
	embedder, err := provider.NewEmbedder(cfg.OpenRouterKey, cfg.BaseUrl, cfg.EmbeddingModel)
	if err != nil {
		log.Fatalf("Error embedder: %v", err)
	}

	// Inicializar Engine
	rag, err := engine.NewRAGEngine(vStore, embedder, cfg)
	if err != nil {
		log.Fatalf("Error engine: %v", err)
	}

	collection := os.Getenv("QDRANT_COLLECTION")
	if collection == "" {
		collection = "documents"
	}

	switch command {
	case "debug":
		count, err := vStore.GetPointsCount(ctx, collection)
		if err != nil {
			log.Fatalf("Error debug: %v", err)
		}
		fmt.Printf("Colección: %s\nDocumentos: %d\n", collection, count)

	case "ingest":
		if len(os.Args) < 3 {
			fmt.Println("Uso: rago ingest <archivo>")
			return
		}
		path := os.Args[2]
		fmt.Printf("Ingestando %s...\n", path)
		err = rag.Ingest(ctx, collection, path)
		if err != nil {
			log.Fatalf("Error ingest: %v", err)
		}
		fmt.Println("¡Ingesta completada!")

	case "ask":
		if len(os.Args) < 3 {
			fmt.Println("Uso: rago ask \"pregunta\"")
			return
		}
		question := os.Args[2]
		resp, err := rag.Ask(ctx, collection, question)
		if err != nil {
			log.Fatalf("Error ask: %v", err)
		}
		fmt.Printf("\nRespuesta:\n%s\n", resp)

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("RAGO CLI - Comandos disponibles:")
	fmt.Println("  debug           - Ver estado de la colección")
	fmt.Println("  ingest <file>   - Ingestar archivo de texto")
	fmt.Println("  ask <pregunta>  - Hacer una pregunta al motor RAG")
}
