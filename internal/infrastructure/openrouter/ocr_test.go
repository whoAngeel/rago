package openrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOCRAdapter_Recognize(t *testing.T) {
	t.Run("returns trimmed content", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/chat/completions") {
				t.Errorf("request = %s %s", r.Method, r.URL.Path)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer k" {
				t.Errorf("Authorization = %q", got)
			}
			var body struct {
				Model    string `json:"model"`
				Messages []struct {
					Content []struct {
						Type     string `json:"type"`
						ImageURL struct {
							URL string `json:"url"`
						} `json:"image_url"`
					} `json:"content"`
				} `json:"messages"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Model != "m" {
				t.Errorf("model = %q", body.Model)
			}
			found := false
			for _, p := range body.Messages[0].Content {
				if p.Type == "image_url" && strings.HasPrefix(p.ImageURL.URL, "data:image/png;base64,") {
					found = true
				}
			}
			if !found {
				t.Error("no data-URL image_url part")
			}
			w.Write([]byte(`{"choices":[{"message":{"content":"  hello \n"}}]}`))
		}))
		defer srv.Close()

		got, err := NewOCRAdapter("k", srv.URL+"/", "m").Recognize(context.Background(), []byte("png"))
		if err != nil || got != "hello" {
			t.Errorf("got %q, %v; want %q, nil", got, err, "hello")
		}
	})

	t.Run("server error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "boom", http.StatusInternalServerError)
		}))
		defer srv.Close()

		_, err := NewOCRAdapter("k", srv.URL, "m").Recognize(context.Background(), []byte("png"))
		if err == nil || !strings.Contains(err.Error(), "500") {
			t.Errorf("err = %v; want status 500 error", err)
		}
	})
}
