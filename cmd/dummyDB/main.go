package main

import (
	"flag"
	"log"

	"github.com/quickfeed/quickfeed/internal/dummydata"
)

func main() {
	githubUsername := flag.String("admin", "", "github username for the admin user")
	flag.Parse()
	if *githubUsername == "" {
		log.Fatal("Provide a github username to create the database")
	}
	gen, err := dummydata.NewGenerator()
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}
	if err := gen.Data(*githubUsername); err != nil {
		log.Fatalf("Failed to generate dummy data: %v", err)
	}
}
