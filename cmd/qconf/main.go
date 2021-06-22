package main

import (
	"os"
	"text/template"

	"github.com/autograde/quickfeed/config"
)

func main() {
	tmpl, err := template.ParseFiles("envoy/envoy.tmpl")
	if err != nil {
		panic(err)
	}
	envoyConfig := &config.QuickFeed{
		DomainName: "cyclone.meling.me",
		GRPCPort:   9090,
		HTTPPort:   8081,
	}

	err = tmpl.Execute(os.Stdout, envoyConfig)
	if err != nil {
		panic(err)
	}
}
