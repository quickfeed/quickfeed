package main

import (
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/ui"
)

// Tool to execute esbuild externally.
// Meant for development purposes.
// Used to rebuild the UI after modifying the source code.
// The frontend src code is located in the public directory.
// The compiled code is placed in the dist directory.
func main() {
	// Load environment variables from .env file
	// This is necessary to ensure the build has access to the correct environment variables.
	if err := env.Load(env.RootEnv(".env")); err != nil {
		panic(err)
	}
	// Errors and warnings will be logged by Esbuild
	_ = ui.Build("", true)
}
