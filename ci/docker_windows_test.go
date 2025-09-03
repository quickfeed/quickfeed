//go:build windows

package ci_test

import (
	"testing"
)

func checkOwner(t *testing.T, _ string) {
	t.Helper()
	t.Log("Skipping checkOwner on Windows")
}
