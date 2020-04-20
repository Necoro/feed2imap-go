package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/Necoro/feed2imap-go/internal/feed"
	"github.com/Necoro/feed2imap-go/internal/imap"
	"github.com/Necoro/feed2imap-go/internal/log"
	"github.com/Necoro/feed2imap-go/internal/yaml"
)

var cfgFile = flag.String("f", "config.yml", "configuration file")
var verbose = flag.Bool("v", false, "enable verbose output")

func run() error {
	flag.Parse()
	log.SetDebug(*verbose)

	log.Print("Starting up...")

	log.Printf("Reading configuration file '%s'", *cfgFile)
	cfg, feeds, err := yaml.Load(*cfgFile)
	if err != nil {
		return err
	}

	if err = cfg.Validate(); err != nil {
		return fmt.Errorf("Configuration invalid: %w", err)
	}

	feed.Parse(feeds)

	for _, f := range feeds {
		mails, err := f.ToMails(cfg)
		if err != nil {
			return err
		}
		_ = mails
		break
	}

	imapUrl, err := url.Parse(cfg.Target)
	if err != nil {
		return fmt.Errorf("parsing 'target': %w", err)
	}

	c, err := imap.Connect(imapUrl)
	if err != nil {
		return err
	}

	defer c.Disconnect()

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
