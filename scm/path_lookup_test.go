package scm

import (
	"testing"
)

func TestLookup(t *testing.T) {
	tests := []struct {
		key     string
		pattern string
		url     string
		want    string
	}{
		{"id", "/organizations/{id}", "/organizations/123", "123"},
		{"org", "/orgs/{org}", "/orgs/foobar", "foobar"},
		{"id", "/organizations/{id}/details", "/organizations/123/details", "123"},
		{"user", "/users/{user}/profile", "/users/alice/profile", "alice"},
		{"id", "/organizations/{id}", "/organizations/", ""}, // no ID present
	}

	for _, tt := range tests {
		got := pathValue(tt.key, tt.pattern, tt.url)
		if got != tt.want {
			t.Errorf("lookup(%q, %q, %q) = %q, want %q", tt.key, tt.pattern, tt.url, got, tt.want)
		}
	}
}
