package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Host string
	Port string

	QdrantHost       string
	QdrantPort       int
	QdrantCollection string

	OpenRouterKey     string
	OpenRouterBaseUrl string
	Model             string
	EmbeddingModel    string
	EmbeddingDim      int

	Env string

	MaxUploadSize int64
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	cfg := &Config{
		Host: getEnv("HOST", "0.0.0.0"),
		Port: getEnv("PORT", "4000"),

		QdrantHost:       getEnv("QDRANT_HOST", "localhost"),
		QdrantPort:       getEnvAsInt("QDRANT_PORT", 6334),
		QdrantCollection: getEnv("QDRANT_COLLECTION", "default"),

		OpenRouterKey:     getEnv("OPEN_ROUTER_API", ""),
		OpenRouterBaseUrl: getEnv("OPEN_ROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
		Model:             getEnv("LLM_MODEL", "google/gemini-2.5-flash"),
		EmbeddingModel:    getEnv("EMBEDDING_MODEL", "text-embedding-3-small"),
		EmbeddingDim:      getEnvAsInt("EMBEDDING_DIMENSION", 1536),

		Env: getEnv("ENV", "development"),

		MaxUploadSize: getEnvAsInt64("MAX_UPLOAD_SIZE", 52428800),
	}

	if cfg.OpenRouterKey == "" {
		return nil, fmt.Errorf("API KEY is required")
	}
	if cfg.QdrantHost == "" {
		return nil, fmt.Errorf("QDRANT HOST is required")
	}

	return cfg, nil
}

// Helpers para parsear env vars
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsInt64(key string, defaultVal int64) int64 {
	valStr := os.Getenv(key)
	if val, err := strconv.ParseInt(valStr, 10, 64); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valStr := os.Getenv(key)
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal string) time.Duration {
	valStr := getEnv(key, defaultVal)
	if duration, err := time.ParseDuration(valStr); err == nil {
		return duration
	}
	// Si falla el parse, retorna el default parseado
	duration, _ := time.ParseDuration(defaultVal)
	return duration
}
func getFloat64(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}
