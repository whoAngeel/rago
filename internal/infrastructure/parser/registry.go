package parser

import (
	"fmt"
	"sync"

	"github.com/whoAngeel/rago/internal/core/ports"
)

type Registry struct {
	mu      sync.RWMutex
	parsers map[string]ports.Parser
}

func NewRegistry() *Registry {
	return &Registry{
		parsers: make(map[string]ports.Parser),
	}
}

func (r *Registry) Register(contentType string, parser ports.Parser) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.parsers[contentType] = parser
}

func (r *Registry) Get(contentType string) (ports.Parser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.parsers[contentType]
	if !ok {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
	return p, nil
}

func (r *Registry) Resolve(contentType string) (ports.Parser, error) {
	return r.Get(contentType)
}
