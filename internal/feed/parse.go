package feed

import (
	ctxt "context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

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

func parseFeed(feed *Feed) error {
	ctx, cancel := context(feed.Global.Timeout)
	defer cancel()

	fp := gofeed.NewParser()
	fp.Client = httpClient(feed.NoTLS)

	parsedFeed, err := fp.ParseURLWithContext(feed.Url, ctx)
	if err != nil {
		return fmt.Errorf("while fetching %s from %s: %w", feed.Name, feed.Url, err)
	}

	feed.feed = parsedFeed
	feed.items = make([]feeditem, len(parsedFeed.Items))
	for idx, item := range parsedFeed.Items {
		feed.items[idx] = feeditem{Feed: parsedFeed, Item: item}
	}
	return nil
}

func handleFeed(feed *Feed) {
	log.Printf("Fetching %s from %s", feed.Name, feed.Url)

	err := parseFeed(feed)
	if err != nil {
		log.Error(err)
	}
}
