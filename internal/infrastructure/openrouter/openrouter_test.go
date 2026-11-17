package openrouter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type capturedRequest struct {
	Messages []struct {
		Role    string `json:"role"`
		Content any    `json:"content"`
	} `json:"messages"`
	Temperature *float64 `json:"temperature"`
	// langchaingo's openai client serialises llms.WithMaxTokens as max_completion_tokens.
	MaxTokens *int `json:"max_completion_tokens"`
}

func newServer(t *testing.T, got *capturedRequest, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, got); err != nil {
			t.Errorf("bad request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

const okBody = `{"id":"1","object":"chat.completion","model":"m","choices":[{"index":0,"message":{"role":"assistant","content":"hola"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

func TestGenerateAnswer_SendsSystemAndUserMessagesWithTemperature(t *testing.T) {
	var got capturedRequest
	srv := newServer(t, &got, okBody)
	a, err := NewOpenRouterAdapter("k", srv.URL, "m", 0.2, 0)
	if err != nil {
		t.Fatal(err)
	}
	res, err := a.GenerateAnswer(context.Background(), "SYSTEM", "USER")
	if err != nil {
		t.Fatal(err)
	}
	if res.Answer != "hola" || res.Model != "m" {
		t.Errorf("result = %+v", res)
	}
	if len(got.Messages) != 2 || got.Messages[0].Role != "system" || got.Messages[1].Role != "user" {
		t.Fatalf("messages = %+v; want system then user", got.Messages)
	}
	if !strings.Contains(toJSON(got.Messages[0].Content), "SYSTEM") || !strings.Contains(toJSON(got.Messages[1].Content), "USER") {
		t.Errorf("message contents = %+v", got.Messages)
	}
	if got.Temperature == nil || *got.Temperature != 0.2 {
		t.Errorf("temperature = %v; want 0.2", got.Temperature)
	}
	if got.MaxTokens != nil {
		t.Errorf("max_completion_tokens = %v; want omitted when 0", *got.MaxTokens)
	}
}

func TestGenerateAnswer_SendsMaxTokensWhenPositive(t *testing.T) {
	var got capturedRequest
	srv := newServer(t, &got, okBody)
	a, _ := NewOpenRouterAdapter("k", srv.URL, "m", 0.2, 512)
	if _, err := a.GenerateAnswer(context.Background(), "S", "U"); err != nil {
		t.Fatal(err)
	}
	if got.MaxTokens == nil || *got.MaxTokens != 512 {
		t.Errorf("max_completion_tokens = %v; want 512", got.MaxTokens)
	}
}

func TestGenerateAnswer_NoChoicesIsAnError(t *testing.T) {
	var got capturedRequest
	srv := newServer(t, &got, `{"id":"1","object":"chat.completion","model":"m","choices":[]}`)
	a, _ := NewOpenRouterAdapter("k", srv.URL, "m", 0.2, 0)
	if _, err := a.GenerateAnswer(context.Background(), "S", "U"); err == nil {
		t.Fatal("want error for empty choices")
	}
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
