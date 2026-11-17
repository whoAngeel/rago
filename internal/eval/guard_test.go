package eval

import "testing"

func TestCheckHost(t *testing.T) {
	tests := []struct {
		host        string
		allowRemote bool
		wantErr     bool
	}{
		{"localhost", false, false},
		{"LOCALHOST", false, false},
		{"127.0.0.1", false, false},
		{"::1", false, false},
		{"[::1]", false, false},
		{"qdrant.prod.example.com", false, true},
		{"10.0.0.5", false, true},
		{"", false, true},
		{"qdrant.prod.example.com", true, false},
	}
	for _, tt := range tests {
		err := CheckHost(tt.host, tt.allowRemote)
		if (err != nil) != tt.wantErr {
			t.Errorf("CheckHost(%q, %v) error = %v, wantErr %v", tt.host, tt.allowRemote, err, tt.wantErr)
		}
	}
}
