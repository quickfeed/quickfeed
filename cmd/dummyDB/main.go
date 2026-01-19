package main

import (
	"flag"
	"log"

	"github.com/quickfeed/quickfeed/internal/dummydata"
)

func main() {
	// TODO(Joachim): remove default admin username
	admin := flag.String("admin", "JoachimTislov", "github username for the admin user")
	flag.Parse()
	if *admin == "" {
		log.Fatal("provide a github username to create the database")
	}
	gen, err := dummydata.NewGenerator()
	if err != nil {
		log.Fatalf("failed to create generator: %v", err)
	}
	if err := gen.Data(*admin); err != nil {
		log.Fatalf("failed to generate dummy data: %v", err)
	}
}
