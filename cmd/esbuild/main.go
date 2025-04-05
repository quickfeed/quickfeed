package main

import (
	"github.com/quickfeed/quickfeed/internal/ui"
)

// Tool to execute esbuild externally.
// Meant for development purposes.
// Used to rebuild the UI after modifying the source code.
// The frontend src code is located in the public directory.
// The compiled code is placed in the dist directory.
func main() {
	// Errors and warnings will be logged by Esbuild
	_ = ui.Build("", true)
}
