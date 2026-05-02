package chunker

import (
	"strings"
	"unicode"
)

const (
	DefaultChunkSize    = 1000
	DefaultChunkOverlap = 200
)

type FixedChunker struct {
	Size    int
	Overlap int
}

func NewFixedChunker(size, overlap int) *FixedChunker {
	if size <= 0 {
		size = DefaultChunkSize
	}
	if overlap < 0 || overlap >= size {
		overlap = DefaultChunkOverlap
	}
	return &FixedChunker{
		Size:    size,
		Overlap: overlap,
	}
}

func (c *FixedChunker) Chunk(text string) ([]string, error) {
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
				if overlap != "" {
					current.WriteString(overlap)
					current.WriteString("\n\n")
				}
			}
			if len(p) > c.Size {
				subchunks := splitSentences(p, c.Size, c.Overlap)
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

func splitSentences(text string, size, overlap int) []string {
	sentences := splitByPunctuation(text)
	var chunks []string
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
			if overlapText != "" {
				current.WriteString(overlapText)
				current.WriteString(" ")
			}
		}
		if current.Len() > 0 {
			current.WriteString(" ")
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

func splitByPunctuation(text string) []string {
	var sentences []string
	start := 0
	runes := []rune(text)

	for i, r := range runes {
		if r == '.' || r == '!' || r == '?' {
			end := i + 1
			if end < len(runes) && runes[end] == ' ' {
				end++
			}
			sentence := strings.TrimSpace(string(runes[start:end]))
			if sentence != "" {
				sentences = append(sentences, sentence)
			}
			start = end
		}
	}
	if start < len(runes) {
		remainder := strings.TrimSpace(string(runes[start:]))
		if remainder != "" {
			sentences = append(sentences, remainder)
		}
	}
	return sentences
}

func splitBySize(text string, size, overlap int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
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
	for i := len(runes) - n; i < len(runes); i++ {
		if unicode.IsSpace(runes[i]) {
			trimmed := strings.TrimSpace(string(runes[i+1:]))
			return trimmed
		}
	}
	return strings.TrimSpace(string(runes[len(runes)-n:]))
}
