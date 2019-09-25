package ci_test

import (
	"context"
	"testing"

	"github.com/autograde/aguis/ci"
)

func TestLocal(t *testing.T) {
	const (
		script  = `printf "hello world"`
		wantOut = "hello world"
	)

	local := ci.Local{}
	out, err := local.Run(context.Background(), &ci.Job{
		Commands: []string{script},
	}, "")
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	}
}
