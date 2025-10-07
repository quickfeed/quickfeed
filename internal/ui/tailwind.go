package ui

import (
	"github.com/quickfeed/quickfeed/internal/env"
	"os"
	"os/exec"
)

func tailwindBuild() error {
	cmd := exec.Command("npm", "run", "tailwind")
	cmd.Stdout = os.Stdout
	cmd.Dir = env.PublicDir()
	return cmd.Run()
}
