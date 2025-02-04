package main

import (
	"fmt"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/rand"
)

// This tool is used to generate a random secret for signing JWT tokens.
func main() {
	fmt.Println("Generating random secret for signing JWT tokens...")
	if err := env.Save(env.RootEnv(".env"), map[string]string{
		"QUICKFEED_AUTH_SECRET": rand.String(),
	}); err != nil {
		panic(err)
	}
	fmt.Println("Done.")
}
