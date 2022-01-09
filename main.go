package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Necoro/feed2imap-go/internal/feed/cache"
	"github.com/Necoro/feed2imap-go/internal/feed/template"
	"github.com/Necoro/feed2imap-go/internal/imap"
	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
	"github.com/Necoro/feed2imap-go/pkg/version"
)

// flags
var (
	cfgFile      string = "config.yml"
	cacheFile    string
	printVersion bool = false
	dryRun       bool = false
	buildCache   bool = false
	verbose      bool = false
	debug        bool = false
)

func init() {
	flag.StringVar(&cfgFile, "f", cfgFile, "configuration file")
	flag.StringVar(&cacheFile, "c", "", "override cache file location")
	flag.BoolVar(&printVersion, "version", printVersion, "print version and exit")
	flag.BoolVar(&dryRun, "dry-run", dryRun, "do everything short of uploading and writing the cache")
	flag.BoolVar(&buildCache, "build-cache", buildCache, "only (re)build the cache; useful after migration or when the cache is lost or corrupted")
	flag.BoolVar(&verbose, "v", verbose, "enable verbose output")
	flag.BoolVar(&debug, "d", debug, "enable debug output")
}

func processFeed(cf cache.CachedFeed, client *imap.Client, dryRun bool) {
	feed := cf.Feed()
	msgs, err := feed.Messages()
	if err != nil {
		log.Errorf("Processing items of feed %s: %s", feed.Name, err)
		return
	}

	if dryRun || len(msgs) == 0 {
		cf.Commit()
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

	cf.Commit()
}

func loadTemplate(path string, tpl template.Template) error {
	if path == "" {
		return nil
	}

	log.Printf("Loading custom %s template from %s", tpl.Name(), path)
	if err := tpl.LoadFile(path); err != nil {
		return fmt.Errorf("loading %s template from %s: %w", tpl.Name(), path, err)
	}
	return nil
}

func run() error {
	flag.Parse()
	if printVersion {
		println("Feed2Imap-Go, " + version.FullVersion())
		return nil
	}

	if debug {
		log.SetDebug()
	} else if verbose {
		log.SetVerbose()
	}

	log.Printf("Starting up (%s)...", version.FullVersion())

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	state, err := cache.NewState(cfg)
	if err != nil {
		return err
	}

	cacheLocation := cacheFile
	if cacheLocation == "" {
		cacheLocation = cfg.Cache
	}
	log.Debugf("Using '%s' as cache location", cacheLocation)

	err = state.LoadCache(cacheLocation, buildCache)
	if err != nil {
		return err
	}
	defer state.UnlockCache()

	state.RemoveUndue()

	if state.NumFeeds() == 0 {
		log.Print("Nothing to do, exiting.")
		// nothing to do
		return nil
	}

	if !buildCache {
		if err = loadTemplate(cfg.HtmlTemplate, template.Html); err != nil {
			return err
		}
		if err = loadTemplate(cfg.TextTemplate, template.Text); err != nil {
			return err
		}
	}

	imapErr := make(chan error, 1)
	var c *imap.Client
	if !dryRun && !buildCache {
		go func() {
			var err error
			c, err = imap.Connect(cfg.Target)
			imapErr <- err
		}()

		defer func() {
			// capture c and not evaluate it, before connect has run
			c.Disconnect()
		}()
	}

	if success := state.Fetch(); success == 0 {
		return fmt.Errorf("No successful feed fetch.")
	}

	state.Filter()

	if buildCache {
		state.Foreach(cache.CachedFeed.Commit)
	} else {
		if !dryRun {
			if err = <-imapErr; err != nil {
				return err
			}
		}
		state.ForeachGo(func(f cache.CachedFeed) {
			processFeed(f, c, dryRun)
		})
	}

	if !dryRun {
		if err = state.StoreCache(cacheLocation); err != nil {
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
