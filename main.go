package main

import (
	"flag"
	"os"

	"github.com/Necoro/feed2imap-go/internal/config"
	"github.com/Necoro/feed2imap-go/internal/log"
	"github.com/Necoro/feed2imap-go/internal/parse"
)

var cfgFile = flag.String("f", "config.yml", "configuration file")
var verbose = flag.Bool("v", false, "enable verbose output")

func run() error {
	flag.Parse()
	log.SetDebug(*verbose)

	log.Print("Starting up...")

	log.Printf("Reading configuration file '%s'", *cfgFile)
	cfg, err := config.Load(*cfgFile)
	if err != nil {
		return err
	}

	parse.Parse(cfg.Feeds)

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
