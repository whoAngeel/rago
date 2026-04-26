package engine

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	QdrantHost      string
	QdrantPort      int
	OpenRouterKey   string
	Model           string
	EmbeddingModel  string
	EmbeddingDim   int
	BaseUrl         string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()
	port := 6334
	if p := os.Getenv("QDRANT_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}
	dim := 1536
	if d := os.Getenv("EMBEDDING_DIMENSION"); d != "" {
		fmt.Sscanf(d, "%d", &dim)
	}
	return &Config{
		QdrantHost:      os.Getenv("QDRANT_HOST"),
		QdrantPort:     port,
		OpenRouterKey: os.Getenv("OPEN_ROUTER_API"),
		Model:         os.Getenv("LLM_MODEL"),
		EmbeddingModel: os.Getenv("EMBEDDING_MODEL"),
		EmbeddingDim:  dim,
		BaseUrl:       os.Getenv("OPEN_ROUTER_BASE_URL"),
	}, nil
}
