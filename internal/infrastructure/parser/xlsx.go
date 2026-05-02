package parser

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tealeg/xlsx/v3"
	"github.com/tmc/langchaingo/schema"
)

type XLSXParser struct{}

func NewXLSXParser() *XLSXParser {
	return &XLSXParser{}
}

func (p *XLSXParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
	tmpDir, err := os.MkdirTemp("", "xlsx-parse-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "input.xlsx")
	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(f, reader); err != nil {
		f.Close()
		return nil, err
	}
	f.Close()

	xlFile, err := xlsx.OpenFile(filePath)
	if err != nil {
		return nil, err
	}

	var docs []schema.Document
	for _, sheet := range xlFile.Sheets {
		headers := extractHeaders(sheet)
		if len(headers) == 0 {
			continue
		}

		for rowIdx := 1; rowIdx < sheet.MaxRow; rowIdx++ {
			row, err := sheet.Row(rowIdx)
			if err != nil {
				continue
			}

			var parts []string
			for j := 0; j < row.Sheet.MaxCol; j++ {
				cell := row.GetCell(j)
				val := strings.TrimSpace(cell.Value)
				if j < len(headers) && headers[j] != "" && val != "" {
					parts = append(parts, fmt.Sprintf("%s: %s", headers[j], val))
				} else if val != "" {
					parts = append(parts, val)
				}
			}

			content := strings.TrimSpace(strings.Join(parts, ", "))
			if content == "" {
				continue
			}

			docs = append(docs, schema.Document{
				PageContent: content,
				Metadata: map[string]any{
					"sheet":      sheet.Name,
					"row":        rowIdx + 1,
					"headers":    strings.Join(headers, "|"),
					"chunk_type": "structured",
				},
			})
		}
	}
	return docs, nil
}

func extractHeaders(sheet *xlsx.Sheet) []string {
	if sheet.MaxRow == 0 {
		return nil
	}
	row, err := sheet.Row(0)
	if err != nil {
		return nil
	}
	var headers []string
	for j := 0; j < row.Sheet.MaxCol; j++ {
		headers = append(headers, strings.TrimSpace(row.GetCell(j).Value))
	}
	return headers
}
