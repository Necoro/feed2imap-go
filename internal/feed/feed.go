package feed

import (
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
)

type Feed struct {
	*config.Feed
	feed   *gofeed.Feed
	items  []feeditem
	cached CachedFeed
}

type feedDescriptor struct {
	Name string
	Url  string
}

type feeditem struct {
	*gofeed.Feed
	*gofeed.Item
}

func (feed *Feed) descriptor() feedDescriptor {
	return feedDescriptor{
		Name: feed.Name,
		Url:  feed.Url,
	}
}

func (feed *Feed) NeedsUpdate(updateTime time.Time) bool {
	if !updateTime.IsZero() && int(time.Since(updateTime).Hours()) >= *feed.MinFreq {
		log.Printf("Feed '%s' does not need updating, skipping.", feed.Name)
		return false
	}
	return true
}

func (feed *Feed) Success() bool {
	return feed.feed != nil
}
