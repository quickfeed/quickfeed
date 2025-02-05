//go:build unix

package ci_test

import (
	"os"
	"syscall"
	"testing"
)

func checkOwner(t *testing.T, path string) {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	stat := fi.Sys().(*syscall.Stat_t)
	if int(stat.Uid) != os.Getuid() {
		t.Errorf("%s has owner %d, expected %d", fi.Name(), stat.Uid, os.Getuid())
	}
	if int(stat.Gid) != os.Getgid() {
		t.Errorf("%s has group %d, expected %d", fi.Name(), stat.Gid, os.Getgid())
	}
}
