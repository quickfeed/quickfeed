//go:build unix

package ci_test

import (
	"context"
	"os"
	"syscall"
	"testing"

	"github.com/quickfeed/quickfeed/ci"
)

func TestLocal(t *testing.T) {
	const (
		script  = `printf "hello world"`
		wantOut = "hello world"
	)

	local := ci.Local{}
	out, err := local.Run(context.Background(), &ci.Job{
		Commands: []string{script},
	})
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	}
}

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
