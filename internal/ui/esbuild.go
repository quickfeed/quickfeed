package ui

import (
	"os"
	"os/exec"

	"github.com/quickfeed/quickfeed/internal/env"
)

// Build builds the UI with esbuild and outputs to the public/dist folder
func Build() error {
	return run("build")
}

// Watch starts a watch process for the frontend, rebuilding on changes
func Watch() error {
	return run("watch")
}

func run(arg string) error {
	cmd := exec.Command("/home/linuxbrew/.linuxbrew/bin/npm", "run", arg)
	cmd.Dir = env.PublicDir()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}
