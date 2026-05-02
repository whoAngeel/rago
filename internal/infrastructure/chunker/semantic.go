package chunker

import (
	"strings"
	"unicode"
)

const (
	DefaultChunkSize    = 1000 // chars
	DefaultChunkOverlap = 200  // solapamientro entre chunks
)

type FixedChuker struct {
	Size    int
	Overlap int
}

func NewFixedChunker(size, overlap int) *FixedChuker {
	if size <= 0 {
		size = DefaultChunkSize
	}
	if overlap < 0 || overlap >= size {
		overlap = DefaultChunkOverlap
	}
	return &FixedChuker{
		Size:    size,
		Overlap: overlap,
	}
}

func (c *FixedChuker) Chunk(text string) ([]string, error) {
	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	var current strings.Builder

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		if current.Len()+len(p) <= c.Size {
			if current.Len() > 0 {
				current.WriteString("\n\n")
			}
			current.WriteString(p)
		} else {
			if current.Len() > 0 {
				chunks = append(chunks, current.String())
				overlap := takeLast(current.String(), c.Overlap)
				current.Reset()
				current.WriteString(overlap)
			}
			if len(p) > c.Size {
				subchunks := splitBySentence(p, c.Size, c.Overlap)
				chunks = append(chunks, subchunks...)
				current.Reset()
			} else {
				current.WriteString(p)
			}
		}
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	return chunks, nil
}

func splitBySentence(text string, size, overlap int) []string {
	var chunks []string
	sentences := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?' || r == '\n'
	})
	var current strings.Builder

	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if current.Len()+len(s) > size && current.Len() > 0 {
			chunks = append(chunks, current.String())
			overlapText := takeLast(current.String(), overlap)
			current.Reset()
			current.WriteString(overlapText)
		}
		if current.Len() > 0 {
			current.WriteString(". ")
		}
		current.WriteString(s)
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	if len(chunks) == 0 {
		return splitBySize(text, size, overlap)
	}

	return chunks
}

func splitBySize(text string, size, overlap int) []string {
	var chunks []string
	runes := []rune(text)
	for i := 0; i < len(runes); i += size - overlap {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	return chunks
}

func takeLast(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	start := 0
	for i := len(runes) - n; i < len(runes); i++ {
		if unicode.IsSpace(runes[i]) {
			start = i + 1
			break
		}
	}
	if start == 0 {
		start = len(runes) - n
	}

	return string(runes[start:])
}

// TODO: semantic chunker
/// 1. dividir texto en oraciones
/// embeddings de cada oracion con el embedder
// calcular cosine similarity entre oraciones consecutivas
// donde la similaridad baja = nuevo chunk
