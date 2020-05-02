package feed

import (
	ctxt "context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/pkg/log"
)

// share HTTP clients
var (
	stdHTTPClient    *http.Client
	unsafeHTTPClient *http.Client
)

func init() {
	// std
	stdHTTPClient = &http.Client{Transport: http.DefaultTransport}

	// unsafe
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig
	unsafeHTTPClient = &http.Client{Transport: transport}
}

func context(timeout int) (ctxt.Context, ctxt.CancelFunc) {
	return ctxt.WithTimeout(ctxt.Background(), time.Duration(timeout)*time.Second)
}

func httpClient(disableTLS bool) *http.Client {
	if disableTLS {
		return unsafeHTTPClient
	}
	return stdHTTPClient
}

func (feed *Feed) parse() error {
	ctx, cancel := context(feed.Global.Timeout)
	defer cancel()

	fp := gofeed.NewParser()
	fp.Client = httpClient(feed.NoTLS)

	parsedFeed, err := fp.ParseURLWithContext(feed.Url, ctx)
	if err != nil {
		return fmt.Errorf("while fetching %s from %s: %w", feed.Name, feed.Url, err)
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
