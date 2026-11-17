package config

import "testing"

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("OPEN_ROUTER_API", "key")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("MINIO_ENDPOINT", "minio:9000")
}

func TestLoad_RetrievalDefaults(t *testing.T) {
	setRequiredEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.RetrievalTopK != 8 {
		t.Errorf("RetrievalTopK = %d; want 8", cfg.RetrievalTopK)
	}
	if cfg.RetrievalScoreThreshold != 0.25 {
		t.Errorf("RetrievalScoreThreshold = %v; want 0.25", cfg.RetrievalScoreThreshold)
	}
}

func TestLoad_RetrievalEnvOverrides(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("RETRIEVAL_TOP_K", "5")
	t.Setenv("RETRIEVAL_SCORE_THRESHOLD", "0.55")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.RetrievalTopK != 5 {
		t.Errorf("RetrievalTopK = %d; want 5", cfg.RetrievalTopK)
	}
	if cfg.RetrievalScoreThreshold != 0.55 {
		t.Errorf("RetrievalScoreThreshold = %v; want 0.55", cfg.RetrievalScoreThreshold)
	}
}

func TestLoad_LLMSamplingDefaults(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("LLM_TEMPERATURE", "")
	t.Setenv("LLM_MAX_TOKENS", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LLMTemperature != 0.2 || cfg.LLMMaxTokens != 0 {
		t.Errorf("defaults = %v, %v; want 0.2, 0", cfg.LLMTemperature, cfg.LLMMaxTokens)
	}
}

func TestLoad_LLMSamplingOverrides(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("LLM_TEMPERATURE", "0.7")
	t.Setenv("LLM_MAX_TOKENS", "2048")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LLMTemperature != 0.7 || cfg.LLMMaxTokens != 2048 {
		t.Errorf("overrides = %v, %v; want 0.7, 2048", cfg.LLMTemperature, cfg.LLMMaxTokens)
	}
}
