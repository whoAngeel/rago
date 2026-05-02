package parser

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/tmc/langchaingo/schema"
)

type CSVParser struct{}

func NewCSVParser() *CSVParser {
	return &CSVParser{}
}

func (p *CSVParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
	csvReader := csv.NewReader(reader)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading csv headers: %w", err)
	}

	var docs []schema.Document
	rowNum := 1
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			rowNum++
			continue
		}

		var parts []string
		for i, h := range headers {
			if i < len(record) {
				parts = append(parts, fmt.Sprintf("%s: %s", h, record[i]))
			}
		}

		docs = append(docs, schema.Document{
			PageContent: strings.Join(parts, ", "),
			Metadata: map[string]any{
				"row":        rowNum,
				"headers":    strings.Join(headers, "|"),
				"chunk_type": "structured",
			},
		})
		rowNum++
	}
	return docs, nil
}
