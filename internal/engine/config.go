package engine

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	QdrantHost    string
	OpenRouterKey string
	Model         string
	BaseUrl       string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()
	return &Config{
		QdrantHost:    os.Getenv("QDRANT_HOST"),
		OpenRouterKey: os.Getenv("OPEN_ROUTER_API"),
		Model:         os.Getenv("LLM_MODEL"),
		BaseUrl:       os.Getenv("OPEN_ROUTER_BASE_URL"),
	}, nil
}
