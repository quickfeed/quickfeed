package ui

import (
	"fmt"
	"os/exec"

	"github.com/quickfeed/quickfeed/internal/env"
)

var (
	build = "tailwind"
	watch = []string{build, "--", "--watch=always"}
)

func tailwindBuild() error {
	return run(build)
}

func tailwindWatch() {
	err := run(watch...)
	if err != nil {
		fmt.Printf("Error running tailwind watcher: %v\n", err)
	}
}

func run(scriptTarget ...string) error {
	cmd := exec.Command("npm", append([]string{"run"}, scriptTarget...)...)
	cmd.Dir = env.PublicDir()
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
