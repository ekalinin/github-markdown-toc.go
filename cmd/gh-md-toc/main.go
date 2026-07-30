package main

import (
	"log"
	"os"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/app"
)

// Entry point
func main() {
	cfg, err := parseConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
