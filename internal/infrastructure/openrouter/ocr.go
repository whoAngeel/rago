package openrouter

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const ocrInstruction = "Transcribe all text in the image exactly as printed, in reading order. " +
	"Keep the original language, numbers, units and punctuation. " +
	"Render tables row by row with cells separated by \" | \". " +
	"Do not translate, summarise, describe images or add commentary. " +
	"Output only the transcription, or an empty response if the page has no text."

const maxErrBody = 300

// OCRAdapter transcribes page images with a multimodal model served through
// the OpenRouter chat completions API. It implements ports.OCR.
type OCRAdapter struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewOCRAdapter builds an OCR adapter for the given model.
func NewOCRAdapter(apiKey, baseURL, model string) *OCRAdapter {
	return &OCRAdapter{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

type ocrContentPart struct {
	Type     string       `json:"type"`
	Text     string       `json:"text,omitempty"`
	ImageURL *ocrImageURL `json:"image_url,omitempty"`
}

type ocrImageURL struct {
	URL string `json:"url"`
}

type ocrMessage struct {
	Role    string           `json:"role"`
	Content []ocrContentPart `json:"content"`
}

type ocrRequest struct {
	Model       string       `json:"model"`
	Temperature float64      `json:"temperature"`
	Messages    []ocrMessage `json:"messages"`
}

type ocrResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Recognize sends the PNG to the model and returns the transcribed text.
func (a *OCRAdapter) Recognize(ctx context.Context, png []byte) (string, error) {
	body, err := json.Marshal(ocrRequest{
		Model:       a.model,
		Temperature: 0,
		Messages: []ocrMessage{{
			Role: "user",
			Content: []ocrContentPart{
				{Type: "text", Text: ocrInstruction},
				{Type: "image_url", ImageURL: &ocrImageURL{URL: "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)}},
			},
		}},
	})
	if err != nil {
		return "", fmt.Errorf("ocr: encoding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ocr: building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ocr: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrBody))
		return "", fmt.Errorf("ocr: status %d: %s", resp.StatusCode, snippet)
	}

	var out ocrResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("ocr: decoding response: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("ocr: response has no choices")
	}
	return stripCodeFence(strings.TrimSpace(out.Choices[0].Message.Content)), nil
}

// stripCodeFence removes a Markdown code fence wrapping the whole text.
func stripCodeFence(s string) string {
	if !strings.HasPrefix(s, "```") || !strings.HasSuffix(s, "```") || len(s) < 6 {
		return s
	}
	inner := s[3 : len(s)-3]
	if i := strings.IndexByte(inner, '\n'); i >= 0 {
		inner = inner[i+1:] // drop the optional language tag line
	}
	return strings.TrimSpace(inner)
}
