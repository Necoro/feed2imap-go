package main

import (
	"flag"
	"log"
	"os"

	"github.com/Necoro/feed2imap-go/internal/config"
)

var cfgFile = flag.String("f", "config.yml", "configuration file")

func run() error {
	log.Print("Starting up...")
	flag.Parse()

	log.Printf("Reading configuration file '%s'", *cfgFile)
	if _, err := config.Load(*cfgFile); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.SetOutput(os.Stderr)
		log.Print("Error: ", err)
		os.Exit(1)
	}
}
