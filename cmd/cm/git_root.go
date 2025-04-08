package main

import (
	"fmt"
	"os/exec"
	"strings"
)

var gitRoot string

func init() {
	var err error
	gitRoot, err = rootDir()
	if err != nil {
		fmt.Println("cm must be run within a git tracked course directory")
		exitErr(err, "Error getting git root")
	}
}

// rootDir determines the root git directory from the current path.
func rootDir() (string, error) {
	dir, err := exec.Command("git", "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil {
		// no error should occur here as long as the code is executed
		// from a Git tracked path
		return "", err
	}
	return strings.TrimSpace(string(dir)), nil
}
