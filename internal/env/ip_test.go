package env_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/internal/env"
)

func TestIsLocal(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"localhost", "localhost", true},
		{"loopback", "127.0.0.1", true},
		{"private", "172.31.120.166", true},
		{"public", "84.22.1.92", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := env.IsLocal(tt.ip)
			if got != tt.want {
				t.Errorf("isLocal(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}
