package feed

import (
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
	"github.com/Necoro/feed2imap-go/pkg/util"
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
	updateOnly bool
	reasons    []string
}

// Creator returns the name of the creating author.
// MUST NOT have `*feeditem` has the receiver, because the template breaks then.
func (item feeditem) Creator() string {
	if item.Item.Author != nil {
		return item.Item.Author.Name
	}
	return ""
}

func (item *feeditem) addReason(reason string) {
	if !util.StrContains(item.reasons, reason) {
		item.reasons = append(item.reasons, reason)
	}
}

func (feed *Feed) descriptor() feedDescriptor {
	return feedDescriptor{
		Name: feed.Name,
		Url:  feed.Url,
	}
}

func (feed *Feed) NeedsUpdate(updateTime time.Time) bool {
	if *feed.MinFreq == 0 { // shortcut
		return true
	}
	if !updateTime.IsZero() && int(time.Since(updateTime).Hours()) < *feed.MinFreq {
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
