package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Host string
	Port string

	Env string

	MaxUploadSize int64
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	cfg := &Config{
		Host: getEnv("HOST", "0.0.0.0"),
		Port: getEnv("PORT", "4000"),

		Env: getEnv("ENV", "development"),

		MaxUploadSize: getEnvAsInt64("MAX_UPLOAD_SIZE", 52428800),
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
