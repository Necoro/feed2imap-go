package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/Necoro/feed2imap-go/internal/feed"
	"github.com/Necoro/feed2imap-go/internal/imap"
	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
)

var cfgFile = flag.String("f", "config.yml", "configuration file")
var cacheFile = flag.String("c", "feed.cache", "cache file")
var verbose = flag.Bool("v", false, "enable verbose output")
var debug = flag.Bool("d", false, "enable debug output")

func processFeed(feed *feed.Feed, cfg *config.Config, client *imap.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	mails, err := feed.ToMails(cfg)
	if err != nil {
		log.Errorf("Processing items of feed %s: %s", feed.Name, err)
		return
	}

	if len(mails) == 0 {
		feed.MarkSuccess()
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

	feed.MarkSuccess()
}

func run() error {
	flag.Parse()
	if *debug {
		log.SetDebug()
	} else if *verbose {
		log.SetVerbose()
	}

	log.Print("Starting up...")

	cfg, err := config.Load(*cfgFile)
	if err != nil {
		return err
	}

	if err = cfg.Validate(); err != nil {
		return fmt.Errorf("Configuration invalid: %w", err)
	}

	state := feed.NewState(cfg)

	err = state.LoadCache(*cacheFile)
	if err != nil {
		return err
	}

	state.RemoveUndue()

	if state.NumFeeds() == 0 {
		// nothing to do
		return nil
	}

	if success := state.Fetch(); success == 0 {
		return fmt.Errorf("No successful feed fetch.")
	}

	state.Filter()

	imapUrl, err := url.Parse(cfg.Target)
	if err != nil {
		return fmt.Errorf("parsing 'target': %w", err)
	}

	c, err := imap.Connect(imapUrl)
	if err != nil {
		return err
	}

	defer c.Disconnect()

	state.ForeachGo(func(f *feed.Feed, wg *sync.WaitGroup) {
		processFeed(f, cfg, c, wg)
	})

	if err = state.StoreCache(*cacheFile); err != nil {
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
