package main

import (
	"flag"
	"log"

	"github.com/quickfeed/quickfeed/internal/mockdata"
)

func main() {
	// TODO(Joachim): remove default admin username
	admin := flag.String("admin", "JoachimTislov", "github username for the admin user")
	flag.Parse()
	if *admin == "" {
		log.Fatal("provide a github username to create the database")
	}
	generator, err := mockdata.NewGenerator()
	if err != nil {
		log.Fatalf("failed to create generator: %v", err)
	}
	if err := generator.Mock(*admin); err != nil {
		log.Fatalf("failed to generate mock data: %v", err)
	}
}
