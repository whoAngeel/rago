package chunker

import (
	"fmt"

	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

// Server defaults; they mirror the defaults in the config package.
const (
	KindRecursive = "recursive"
	KindFixed     = "fixed"

	ServerKind    = KindRecursive
	ServerSize    = 1500
	ServerOverlap = 200
	ServerHeader  = true
)

// Settings selects the chunking strategy: the chunker plus whether chunks get
// a context header.
type Settings struct {
	Kind    string // KindRecursive or KindFixed
	Size    int
	Overlap int
	Header  bool
}

// DefaultSettings is what the server uses when nothing is configured.
func DefaultSettings() Settings {
	return Settings{Kind: ServerKind, Size: ServerSize, Overlap: ServerOverlap, Header: ServerHeader}
}

// Validate reports why the settings cannot build a chunker.
func (s Settings) Validate() error {
	if s.Size <= 0 || s.Overlap < 0 || s.Overlap >= s.Size {
		return fmt.Errorf("invalid chunking: size %d, overlap %d (need size > 0 and 0 <= overlap < size)", s.Size, s.Overlap)
	}
	if s.Kind != KindFixed && s.Kind != KindRecursive {
		return fmt.Errorf("unknown chunker %q (want %s or %s)", s.Kind, KindFixed, KindRecursive)
	}
	return nil
}

// Chunker builds the configured chunker.
func (s Settings) Chunker() (ports.Chunker, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	if s.Kind == KindFixed {
		return NewFixedChunker(s.Size, s.Overlap), nil
	}
	return NewRecursiveChunker(s.Size, s.Overlap), nil
}

// String is the one-line description logged at startup.
func (s Settings) String() string {
	return fmt.Sprintf("chunking: %s, size %d, overlap %d, context header %t", s.Kind, s.Size, s.Overlap, s.Header)
}

// SettingsFromConfig reads the chunking settings from cfg, falling back to the
// defaults (and returning the reason as warn) when they are invalid.
func SettingsFromConfig(cfg config.Config) (s Settings, warn error) {
	s = Settings{Kind: cfg.Chunker, Size: cfg.ChunkSize, Overlap: cfg.ChunkOverlap, Header: cfg.ChunkContextHeader}
	if err := s.Validate(); err != nil {
		return DefaultSettings(), err
	}
	return s, nil
}

// FromConfig builds the server's chunker from cfg. Invalid settings fall back
// to the defaults with one warning on log. It returns the effective settings.
func FromConfig(cfg config.Config, log ports.Logger) (ports.Chunker, Settings) {
	s, warn := SettingsFromConfig(cfg)
	if warn != nil {
		log.Warn("invalid chunking configuration, using defaults", "error", warn)
	}
	c, err := s.Chunker()
	if err != nil { // unreachable: s is valid
		panic(err)
	}
	return c, s
}
