package parser

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"regexp"
	"strings"

	"github.com/tmc/langchaingo/schema"
)

var (
	docxParagraphRE = regexp.MustCompile(`(?s)<w:p[ >].*?</w:p>`)
	docxTextRE      = regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)
)

type DOCXParser struct{}

func NewDOCXParser() *DOCXParser {
	return &DOCXParser{}
}

func (p *DOCXParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	var documentXML []byte
	for _, f := range zipReader.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			documentXML, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, err
			}
			break
		}
	}

	if documentXML == nil {
		return nil, err
	}

	text := extractTextFromDocxXML(documentXML)
	if strings.TrimSpace(text) == "" {
		return nil, err
	}

	return []schema.Document{{
		PageContent: text,
		Metadata: map[string]any{
			"content_type": contentType,
		},
	}}, nil
}

func extractTextFromDocxXML(data []byte) string {
	paragraphs := docxParagraphRE.FindAll(data, -1)
	if len(paragraphs) == 0 {
		return string(data)
	}

	var lines []string
	for _, p := range paragraphs {
		matches := docxTextRE.FindAllSubmatch(p, -1)
		var line strings.Builder
		for _, m := range matches {
			line.Write(m[1])
		}
		if text := strings.TrimSpace(line.String()); text != "" {
			lines = append(lines, text)
		}
	}
	return strings.Join(lines, "\n\n")
}
