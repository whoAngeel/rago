package eval

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tmc/langchaingo/schema"
	"github.com/whoAngeel/rago/internal/core/ports"
	parserpkg "github.com/whoAngeel/rago/internal/infrastructure/parser"
)

// ParseSignature identifies how documents are parsed; a cache entry is only
// valid for the signature it was written with.
func ParseSignature(engine, model, mode string, forceOCR bool) string {
	return fmt.Sprintf("parser=%s;engine=%s;model=%s;mode=%s;force=%t", parserpkg.Version, engine, model, mode, forceOCR)
}

// ParseCache stores parsed pages on disk, one JSON file per (file bytes,
// signature) key. An empty Dir disables it.
type ParseCache struct {
	Dir       string
	Signature string
	Warn      func(string) // cache write problems; defaults to stderr

	Parsed int // files parsed by the parser
	Cached int // files served from the cache
}

type cachedPage struct {
	Text     string         `json:"text"`
	Metadata map[string]any `json:"metadata"`
}

func (c *ParseCache) path(content []byte) string {
	sum := sha256.Sum256(content)
	key := sha256.Sum256([]byte(hex.EncodeToString(sum[:]) + "\x00" + c.Signature))
	return filepath.Join(c.Dir, hex.EncodeToString(key[:])+".json")
}

func (c *ParseCache) warn(msg string) {
	if c.Warn != nil {
		c.Warn(msg)
		return
	}
	fmt.Fprintln(os.Stderr, "warning:", msg)
}

// Docs returns the parsed pages of a document, from the cache when possible.
func (c *ParseCache) Docs(ctx context.Context, reg ports.ParserRegistry, name string, content []byte) ([]schema.Document, error) {
	if c.Dir == "" {
		c.Parsed++
		return ParseDocs(ctx, reg, name, content)
	}
	path := c.path(content)
	if data, err := os.ReadFile(path); err == nil {
		var pages []cachedPage
		if json.Unmarshal(data, &pages) == nil {
			c.Cached++
			docs := make([]schema.Document, len(pages))
			for i, p := range pages {
				docs[i] = schema.Document{PageContent: p.Text, Metadata: p.Metadata}
			}
			return docs, nil
		}
	}
	docs, err := ParseDocs(ctx, reg, name, content)
	if err != nil {
		return nil, err
	}
	c.Parsed++
	c.store(path, docs)
	return docs, nil
}

func (c *ParseCache) store(path string, docs []schema.Document) {
	pages := make([]cachedPage, len(docs))
	for i, d := range docs {
		pages[i] = cachedPage{Text: d.PageContent, Metadata: d.Metadata}
	}
	data, err := json.Marshal(pages)
	if err == nil {
		err = os.MkdirAll(c.Dir, 0o755)
	}
	if err == nil {
		tmp := path + ".tmp"
		if err = os.WriteFile(tmp, data, 0o644); err == nil {
			if err = os.Rename(tmp, path); err != nil {
				os.Remove(tmp)
			}
		}
	}
	if err != nil {
		c.warn(fmt.Sprintf("parse cache write failed: %v", err))
	}
}

// Summary is the one-line stderr report of cache usage.
func (c *ParseCache) Summary() string {
	return fmt.Sprintf("parse cache: %d files parsed, %d from cache", c.Parsed, c.Cached)
}
