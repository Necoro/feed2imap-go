package feed

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/internal/http"
	"github.com/Necoro/feed2imap-go/pkg/log"
)

func (feed *Feed) parse() error {
	fp := gofeed.NewParser()

	// we do not use the http support in gofeed, so that we can control the behavior of http requests
	// and ensure it to be the same in all places
	resp, cancel, err := http.Get(feed.Url, feed.Global.Timeout, feed.NoTLS)
	if err != nil {
		return fmt.Errorf("while fetching %s from %s: %w", feed.Name, feed.Url, err)
	}
	defer cancel() // includes resp.Body.Close

	parsedFeed, err := fp.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("parsing feed '%s': %w", feed.Name, err)
	}

	feed.feed = parsedFeed
	feed.items = make([]item, len(parsedFeed.Items))
	for idx, feedItem := range parsedFeed.Items {
		feed.items[idx] = item{Feed: parsedFeed, Item: feedItem, itemId: uuid.New(), feed: feed}
	}
	return nil
}

func handleFeed(feed *Feed) {
	log.Printf("Fetching %s from %s", feed.Name, feed.Url)

	err := feed.parse()
	if err != nil {
		if feed.cached.Failures() >= feed.Global.MaxFailures {
			log.Error(err)
		} else {
			log.Print(err)
		}
	}
}
