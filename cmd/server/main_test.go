package main

import (
	"testing"

	"github.com/whoAngeel/rago/internal/application"
)

func TestShouldUpgradeSystemPrompt(t *testing.T) {
	tests := []struct {
		name   string
		stored string
		want   bool
	}{
		{"legacy text is replaced", legacyDefaultSystemPrompt, true},
		{"customised text is kept", legacyDefaultSystemPrompt + "\nExtra.", false},
		{"new default is kept", application.DefaultSystemPrompt, false},
		{"empty is kept", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldUpgradeSystemPrompt(tt.stored); got != tt.want {
				t.Errorf("shouldUpgradeSystemPrompt = %v; want %v", got, tt.want)
			}
		})
	}
}
