package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

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

	imapUrl, err := url.Parse(cfg.Target)
	if err != nil {
		return fmt.Errorf("parsing 'target': %w", err)
	}

	c, err := imap.Connect(imapUrl)
	if err != nil {
		return err
	}

	defer c.Disconnect()

	for _, f := range feeds {
		mails, err := f.ToMails(cfg)
		if err != nil {
			return err
		}
		if len(mails) == 0 {
			continue
		}
		folder := c.NewFolder(f.Target)
		if err = c.EnsureFolder(folder); err != nil {
			return err
		}
		for _, mail := range mails {
			if err = c.PutMessage(folder, mail, time.Now()); err != nil {
				return err
			} // TODO
		}
		log.Printf("Uploaded %d messages to '%s' @ %s", len(mails), f.Name, folder)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
