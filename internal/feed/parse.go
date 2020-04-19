package feed

import (
	ctxt "context"
	"fmt"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/internal/log"
)

func context() (ctxt.Context, ctxt.CancelFunc) {
	return ctxt.WithTimeout(ctxt.Background(), 60*time.Second)
}

func parseFeed(feed *Feed) error {
	ctx, cancel := context()
	defer cancel()
	fp := gofeed.NewParser()
	parsedFeed, err := fp.ParseURLWithContext(feed.Url, ctx)
	if err != nil {
		return fmt.Errorf("while fetching %s from %s: %w", feed.Name, feed.Url, err)
	}

	feed.feed = parsedFeed
	feed.items = make([]feeditem, len(parsedFeed.Items))
	for _, item := range parsedFeed.Items {
		feed.items = append(feed.items, feeditem{parsedFeed, item})
	}
	return nil
}

func handleFeed(feed *Feed, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("Fetching %s from %s", feed.Name, feed.Url)

	if err := parseFeed(feed); err != nil {
		log.Error(err)
	}
}

func Parse(feeds Feeds) {
	var wg sync.WaitGroup
	wg.Add(len(feeds))

	for _, feed := range feeds {
		go handleFeed(feed, &wg)
	}

	wg.Wait()
}
