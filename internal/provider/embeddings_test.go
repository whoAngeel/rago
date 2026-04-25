package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestComputeEmbeddings(t *testing.T) {
	godotenv.Load("../../.env")
	apiKey := os.Getenv("OPEN_ROUTER_API")
	baseURL := os.Getenv("OPEN_ROUTER_BASE_URL")

	fmt.Println(apiKey, baseURL)

	e, err := NewEmbedder(apiKey, baseURL, "text-embedding-3-small")

	if err != nil {
		t.Fatalf("Error al crear embedder: %v", err)
	}

	texts := []string{"Hola mundo", "Rago Go"}
	vectors, err := e.ComputeEmbeddings(context.Background(), texts)

	if err != nil {
		t.Errorf("Error computando: %v", err)
	}

	fmt.Printf("Vectores obtenidos: %v\n", len(vectors))
	fmt.Printf("Vectores 1: %v\n", len(vectors[0]))
	if len(vectors) != 2 {
		t.Errorf("Se esperaban 2 vectores, se obtuvieron %d", len(vectors))
	}
}
