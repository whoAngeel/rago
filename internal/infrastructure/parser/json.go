package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/tmc/langchaingo/schema"
)

type JSONParser struct{}

func NewJSONParser() *JSONParser {
	return &JSONParser{}
}

func (p *JSONParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
	var data any
	if err := json.NewDecoder(reader).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding json: %w", err)
	}

	var docs []schema.Document

	if arr, ok := data.([]any); ok {
		for i, item := range arr {
			docs = append(docs, schema.Document{
				PageContent: flattenJSON(item),
				Metadata: map[string]any{
					"index":      i,
					"chunk_type": "structured",
				},
			})
		}
	} else {
		docs = append(docs, schema.Document{
			PageContent: flattenJSON(data),
			Metadata: map[string]any{
				"chunk_type": "structured",
			},
		})
	}
	return docs, nil
}

func flattenJSON(v any) string {
	switch val := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var parts []string
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s: %v", k, flattenJSON(val[k])))
		}
		return strings.Join(parts, ", ")
	case []any:
		var parts []string
		for _, item := range val {
			parts = append(parts, flattenJSON(item))
		}
		return strings.Join(parts, " | ")
	default:
		return fmt.Sprintf("%v", v)
	}
}
