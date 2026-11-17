package parser

import (
	"context"
	"errors"
	"testing"
)

type stubOCR struct {
	text  string
	err   error
	calls int
}

func (s *stubOCR) Recognize(context.Context, []byte) (string, error) {
	s.calls++
	return s.text, s.err
}

func TestFallbackOCR(t *testing.T) {
	t.Run("primary error uses secondary", func(t *testing.T) {
		p, s := &stubOCR{err: errors.New("down")}, &stubOCR{text: "backup"}
		got, err := NewFallbackOCR(p, s).Recognize(context.Background(), nil)
		if err != nil || got != "backup" {
			t.Errorf("got %q, %v", got, err)
		}
	})
	t.Run("primary success skips secondary", func(t *testing.T) {
		p, s := &stubOCR{text: "main"}, &stubOCR{text: "backup"}
		got, err := NewFallbackOCR(p, s).Recognize(context.Background(), nil)
		if err != nil || got != "main" || s.calls != 0 {
			t.Errorf("got %q, %v, secondary calls %d", got, err, s.calls)
		}
	})
}
