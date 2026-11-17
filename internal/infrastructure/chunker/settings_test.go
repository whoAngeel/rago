package chunker

import (
	"testing"

	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

type warnCounter struct {
	ports.Logger
	warns int
}

func (w *warnCounter) Warn(string, ...any) { w.warns++ }

func TestFromConfig_Valid(t *testing.T) {
	log := &warnCounter{}
	c, s := FromConfig(config.Config{Chunker: "recursive", ChunkSize: 800, ChunkOverlap: 100, ChunkContextHeader: false}, log)
	if _, ok := c.(*RecursiveChunker); !ok {
		t.Errorf("chunker = %T, want *RecursiveChunker", c)
	}
	if want := (Settings{Kind: "recursive", Size: 800, Overlap: 100, Header: false}); s != want {
		t.Errorf("settings = %+v, want %+v", s, want)
	}
	if log.warns != 0 {
		t.Errorf("warnings = %d, want 0", log.warns)
	}
	c, s = FromConfig(config.Config{Chunker: "fixed", ChunkSize: 1000, ChunkOverlap: 200, ChunkContextHeader: true}, log)
	if _, ok := c.(*FixedChunker); !ok || !s.Header {
		t.Errorf("fixed config gave %T, %+v", c, s)
	}
}

func TestFromConfig_InvalidFallsBackToDefaults(t *testing.T) {
	for name, cfg := range map[string]config.Config{
		"size zero":        {Chunker: "recursive", ChunkSize: 0, ChunkOverlap: 0},
		"negative overlap": {Chunker: "recursive", ChunkSize: 500, ChunkOverlap: -1},
		"overlap >= size":  {Chunker: "recursive", ChunkSize: 500, ChunkOverlap: 500},
		"unknown chunker":  {Chunker: "magic", ChunkSize: 500, ChunkOverlap: 50},
	} {
		t.Run(name, func(t *testing.T) {
			log := &warnCounter{}
			c, s := FromConfig(cfg, log)
			if s != DefaultSettings() {
				t.Errorf("settings = %+v, want defaults %+v", s, DefaultSettings())
			}
			if _, ok := c.(*RecursiveChunker); !ok {
				t.Errorf("chunker = %T, want *RecursiveChunker", c)
			}
			if log.warns != 1 {
				t.Errorf("warnings = %d, want 1", log.warns)
			}
		})
	}
}

func TestDefaultSettingsMatchConfigDefaults(t *testing.T) {
	t.Setenv("OPEN_ROUTER_API", "key")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("MINIO_ENDPOINT", "minio:9000")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	s, warn := SettingsFromConfig(*cfg)
	if warn != nil || s != DefaultSettings() {
		t.Errorf("config defaults = %+v (warn %v), want %+v", s, warn, DefaultSettings())
	}
}
