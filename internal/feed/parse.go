package feed

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/internal/http"
	"github.com/Necoro/feed2imap-go/pkg/log"
)

func (feed *Feed) parse() error {
	fp := gofeed.NewParser()

	var reader io.Reader
	var cleanup func() error

	if feed.Url != "" {
		// we do not use the http support in gofeed, so that we can control the behavior of http requests
		// and ensure it to be the same in all places
		resp, cancel, err := http.Get(feed.Url, feed.Global.Timeout, feed.NoTLS)
		if err != nil {
			return fmt.Errorf("while fetching %s from %s: %w", feed.Name, feed.Url, err)
		}
		defer cancel() // includes resp.Body.Close

		reader = resp.Body
		cleanup = func() error { return nil }
	} else { // exec
		// we use the same context as for HTTP
		ctx, cancel := http.Context(feed.Global.Timeout)
		cmd := exec.CommandContext(ctx, feed.Exec[0], feed.Exec[1:]...)
		defer func() {
			cancel()
			// cmd.Wait might have already been called -- but call it again to be sure
			_ = cmd.Wait()
		}()

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("preparing exec for feed '%s': %w", feed.Name, err)
		}

		if err = cmd.Start(); err != nil {
			return fmt.Errorf("starting exec for feed '%s: %w", feed.Name, err)
		}

		reader = stdout
		cleanup = cmd.Wait
	}

	parsedFeed, err := fp.Parse(reader)
	if err != nil {
		return fmt.Errorf("parsing feed '%s': %w", feed.Name, err)
	}

	feed.feed = parsedFeed
	feed.items = make([]item, len(parsedFeed.Items))
	for idx, feedItem := range parsedFeed.Items {
		feed.items[idx] = item{Feed: parsedFeed, Item: feedItem, itemId: uuid.New(), feed: feed}
	}
	return cleanup()
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
