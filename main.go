package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/Necoro/feed2imap-go/internal/feed"
	"github.com/Necoro/feed2imap-go/internal/imap"
	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
	"github.com/Necoro/feed2imap-go/pkg/version"
)

var printVersion = flag.Bool("version", false, "print version and exit")
var cfgFile = flag.String("f", "config.yml", "configuration file")
var cacheFile = flag.String("c", "feed.cache", "cache file")
var verbose = flag.Bool("v", false, "enable verbose output")
var debug = flag.Bool("d", false, "enable debug output")
var dryRun = flag.Bool("dry-run", false, "do everything short of uploading and writing the cache")
var buildCache = flag.Bool("build-cache", false, "only (re)build the cache; useful after migration or when the cache is lost or corrupted")

func processFeed(feed *feed.Feed, client *imap.Client, dryRun bool) {
	msgs, err := feed.Messages()
	if err != nil {
		log.Errorf("Processing items of feed %s: %s", feed.Name, err)
		return
	}

	if dryRun || len(msgs) == 0 {
		feed.MarkSuccess()
		return
	}

	folder := client.NewFolder(feed.Target)
	if err = client.EnsureFolder(folder); err != nil {
		log.Errorf("Creating folder of feed %s: %s", feed.Name, err)
		return
	}

	if err = msgs.Upload(client, folder, feed.Reupload); err != nil {
		log.Errorf("Uploading messages of feed %s: %s", feed.Name, err)
		return
	}

	log.Printf("Uploaded %d messages to '%s' @ %s", len(msgs), feed.Name, folder)

	feed.MarkSuccess()
}

func run() error {
	flag.Parse()
	if *printVersion {
		println("Feed2Imap-Go, " + version.FullVersion())
		return nil
	}

	if *debug {
		log.SetDebug()
	} else if *verbose {
		log.SetVerbose()
	}

	log.Printf("Starting up (%s)...", version.FullVersion())

	cfg, err := config.Load(*cfgFile)
	if err != nil {
		return err
	}

	if err = cfg.Validate(); err != nil {
		return fmt.Errorf("Configuration invalid: %w", err)
	}

	state := feed.NewState(cfg)

	err = state.LoadCache(*cacheFile, *buildCache)
	if err != nil {
		return err
	}

	state.RemoveUndue()

	if state.NumFeeds() == 0 {
		log.Print("Nothing to do, exiting.")
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

	var c *imap.Client
	if !*dryRun && !*buildCache {
		if c, err = imap.Connect(imapUrl); err != nil {
			return err
		}

		defer c.Disconnect()
	}

	if !*buildCache {
		state.ForeachGo(func(f *feed.Feed) {
			processFeed(f, c, *dryRun)
		})
	}

	if !*dryRun {
		if err = state.StoreCache(*cacheFile); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
