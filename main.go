package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/Necoro/feed2imap-go/internal/cache"
	"github.com/Necoro/feed2imap-go/internal/config"
	"github.com/Necoro/feed2imap-go/internal/feed"
	"github.com/Necoro/feed2imap-go/internal/imap"
	"github.com/Necoro/feed2imap-go/internal/log"
	"github.com/Necoro/feed2imap-go/internal/yaml"
)

var cfgFile = flag.String("f", "config.yml", "configuration file")
var cacheFile = flag.String("c", "feed.cache", "cache file")
var verbose = flag.Bool("v", false, "enable verbose output")

func processFeed(feed *feed.Feed, cfg *config.Config, client *imap.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	mails, err := feed.ToMails(cfg)
	if err != nil {
		log.Errorf("Processing items of feed %s: %s", feed.Name, err)
		return
	}

	if len(mails) == 0 {
		return
	}

	folder := client.NewFolder(feed.Target)
	if err = client.EnsureFolder(folder); err != nil {
		log.Errorf("Creating folder of feed %s: %s", feed.Name, err)
		return
	}

	if err = client.PutMessages(folder, mails); err != nil {
		log.Errorf("Uploading messages of feed %s: %s", feed.Name, err)
		return
	}

	log.Printf("Uploaded %d messages to '%s' @ %s", len(mails), feed.Name, folder)
}

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

	if success := feed.Parse(feeds); success == 0 {
		return fmt.Errorf("No successfull feed fetch.")
	}

	feedCache, err := cache.Read(*cacheFile)
	if err != nil {
		return err
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

	var wg sync.WaitGroup
	wg.Add(len(feeds))
	for _, f := range feeds {
		go processFeed(f, cfg, c, &wg)
	}
	wg.Wait()

	if err = cache.Store(*cacheFile, feedCache); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
