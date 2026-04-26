// logger for test
package logger

import "github.com/whoAngeel/rago/internal/core/ports"

type NoopLogger struct{}

func NewNoop() ports.Logger { return &NoopLogger{} }

func (n *NoopLogger) Debug(_ string, _ ...any)   {}
func (n *NoopLogger) Info(_ string, _ ...any)    {}
func (n *NoopLogger) Warn(_ string, _ ...any)    {}
func (n *NoopLogger) Error(_ string, _ ...any)   {}
func (n *NoopLogger) Fatal(_ string, _ ...any)   {}
func (n *NoopLogger) With(_ ...any) ports.Logger { return &NoopLogger{} }
