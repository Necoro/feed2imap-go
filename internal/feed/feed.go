package feed

import (
	"fmt"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/internal/feed/filter"
	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
)

type Feed struct {
	*config.Feed
	feed   *gofeed.Feed
	filter *filter.Filter
	items  []item
	cached CachedFeed
	Global config.GlobalOptions
}

type feedDescriptor struct {
	Name string
	Url  string
}

func (feed *Feed) descriptor() feedDescriptor {
	var url string
	if feed.Url != "" {
		url = feed.Url
	} else {
		url = "exec://" + strings.Join(feed.Exec, "/")
	}
	return feedDescriptor{
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

func (feed *Feed) MarkSuccess() {
	if feed.cached != nil {
		feed.cached.Commit()
	}
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
