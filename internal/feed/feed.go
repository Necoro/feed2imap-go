package feed

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Necoro/gofeed"

	"github.com/Necoro/feed2imap-go/internal/feed/filter"
	"github.com/Necoro/feed2imap-go/internal/http"
	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
)

type Feed struct {
	*config.Feed
	feed   *gofeed.Feed
	filter *filter.Filter
	items  []Item
	Global config.GlobalOptions
	extID  FeedID
}

type FeedID interface {
	String() string
}

type FilterFunc func(items []Item, ignHash, alwaysNew bool) []Item

type Descriptor struct {
	Name string
	Url  string
}

func (feed *Feed) Context() http.Context {
	return http.Context{
		Timeout:    feed.Global.Timeout,
		DisableTLS: feed.NoTLS,
	}
}

func (feed *Feed) Descriptor() Descriptor {
	var url string
	if feed.Url != "" {
		url = feed.Url
	} else {
		url = "exec://" + strings.Join(feed.Exec, "/")
	}
	return Descriptor{
		Name: feed.Name,
		Url:  url,
	}
}

func (feed *Feed) NeedsUpdate(updateTime time.Time) bool {
	if feed.MinFreq == 0 { // shortcut
		return true
	}
	if !updateTime.IsZero() && int(time.Since(updateTime).Hours()) < feed.MinFreq {
		log.Printf("Feed '%s' does not need updating, skipping.", feed.Name)
		return false
	}
	return true
}

func (feed *Feed) FetchSuccessful() bool {
	return feed.feed != nil
}

func Create(parsedFeed *config.Feed, global config.GlobalOptions) (*Feed, error) {
	var itemFilter *filter.Filter
	var err error
	if parsedFeed.ItemFilter != "" {
		if itemFilter, err = filter.New(parsedFeed.ItemFilter); err != nil {
			return nil, fmt.Errorf("Feed %s: Parsing item-filter: %w", parsedFeed.Name, err)
		}
	}
	return &Feed{Feed: parsedFeed, Global: global, filter: itemFilter}, nil
}

func (feed *Feed) filterItems() []Item {
	if feed.filter == nil {
		return feed.items
	}

	items := make([]Item, 0, len(feed.items))

	for _, item := range feed.items {
		res, err := feed.filter.Run(item.Item)
		if err != nil {
			log.Errorf("Feed %s: Item %s: Error applying item filter: %s", feed.Name, printItem(item.Item), err)
			res = true // include
		}

		if res {
			if log.IsDebug() {
				log.Debugf("Filter '%s' matches for item %s", feed.ItemFilter, printItem(item.Item))
			}
			items = append(items, item)
		} else if log.IsDebug() { // printItem is not for free
			log.Debugf("Filter '%s' does not match for item %s", feed.ItemFilter, printItem(item.Item))
		}
	}
	return items
}

func (feed *Feed) Filter(filter FilterFunc) {
	if len(feed.items) > 0 {
		origLen := len(feed.items)

		log.Debugf("Filtering %s. Starting with %d items", feed.Name, origLen)

		items := feed.filterItems()
		newLen := len(items)
		if newLen < origLen {
			log.Printf("Item filter on %s: Reduced from %d to %d items.", feed.Name, origLen, newLen)
			origLen = newLen
		}

		feed.items = filter(items, feed.IgnHash, feed.AlwaysNew)

		newLen = len(feed.items)
		if newLen < origLen {
			log.Printf("Filtered %s. Reduced from %d to %d items.", feed.Name, origLen, newLen)
		} else {
			log.Printf("Filtered %s, no reduction.", feed.Name)
		}

	} else {
		log.Debugf("No items for %s. No filtering.", feed.Name)
	}
}

func (feed *Feed) SetExtID(extID FeedID) {
	feed.extID = extID
}

func (feed *Feed) id() string {
	if feed.extID == nil {
		return feed.Name
	}
	return feed.extID.String()
}

func (feed *Feed) url() *url.URL {
	var feedUrl *url.URL

	tryUrl := func(content, what string) bool {
		var err error
		if content != "" {
			feedUrl, err = url.Parse(content)
			if err != nil {
				log.Errorf("%s '%s' of feed '%s' is not a valid URL.", what, content, feed.Name)
			} else {
				return true
			}
		}
		return false
	}

	if !(tryUrl(feed.feed.FeedLink, "Self-Link") ||
		tryUrl(feed.Url, "URL") ||
		tryUrl(feed.feed.Link, "Link")) {
		panic(fmt.Sprintf("Could not find a valid URL for for feed '%s'. How have we ended up here?", feed.Name))
	}

	return feedUrl
}
