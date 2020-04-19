package feed

import (
	ctxt "context"
	"fmt"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/internal/config"
	"github.com/Necoro/feed2imap-go/internal/log"
)

func context() (ctxt.Context, ctxt.CancelFunc) {
	return ctxt.WithTimeout(ctxt.Background(), 60*time.Second)
}

func parseFeed(feed *config.Feed) error {
	ctx, cancel := context()
	defer cancel()
	fp := gofeed.NewParser()
	if _, err := fp.ParseURLWithContext(feed.Url, ctx); err != nil {
		return fmt.Errorf("while fetching %s from %s: %w", feed.Name, feed.Url, err)
	}

	return nil
}

func handleFeed(feed *config.Feed, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("Fetching %s from %s", feed.Name, feed.Url)

	if err := parseFeed(feed); err != nil {
		log.Error(err)
	}
}

func Parse(feeds config.Feeds) {
	var wg sync.WaitGroup
	wg.Add(len(feeds))

	for _, feed := range feeds {
		go handleFeed(feed, &wg)
	}

	wg.Wait()
}
