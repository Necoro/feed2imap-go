package main

import (
	"flag"
	"log"
	"os"

	"github.com/Necoro/feed2imap-go/internal/config"
	"github.com/Necoro/feed2imap-go/internal/parse"
	"github.com/Necoro/feed2imap-go/internal/util"
)

var cfgFile = flag.String("f", "config.yml", "configuration file")

func run() error {
	log.Print("Starting up...")
	flag.Parse()

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
		util.Error(err)
		os.Exit(1)
	}
}
