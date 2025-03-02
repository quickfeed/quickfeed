package main

import (
	"log"

	"github.com/quickfeed/quickfeed/internal/ui"
)

// Tool to execute esbuild externally.
// Used to rebuild the UI after modifying the source code.
// The frontend src code is located in the public directory.
// The compiled code is placed in the dist directory.
func main() {
	if err := ui.Build(); err != nil {
		log.Fatalf("failed to build UI: %v", err)
	}
}
