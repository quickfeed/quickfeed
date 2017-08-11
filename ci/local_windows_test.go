package ci_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/autograde/aguis/ci"
)

func TestLocalWindows(t *testing.T) {
	const (
		cmd      = `printf "Hello World"`
		expected = "Hello World"
	)

	local := ci.Local{}

	out, err := local.Run(context.Background(), &ci.Job{
		Commands: []string{cmd},
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(out)
	if expected != out {
		t.Errorf("Have %#v want %#v", out, expected)
	}
}
