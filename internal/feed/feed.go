package feed

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/internal/config"
	"github.com/Necoro/feed2imap-go/internal/log"
)

type Feed struct {
	Name   string
	Target []string
	Url    string
	config.Options
	feed   *gofeed.Feed
	items  []feeditem
	cached CachedFeed
}

type feeditem struct {
	*gofeed.Feed
	*gofeed.Item
}

type Feeds struct {
	feeds map[string]*Feed
	cache Cache
}

func NewFeeds() *Feeds {
	return &Feeds{
		feeds: map[string]*Feed{},
	}
}

func (feeds *Feeds) String() string {
	var b strings.Builder
	app := func(a ...interface{}) {
		_, _ = fmt.Fprint(&b, a...)
	}
	app("Feeds [")

	first := true
	for k, v := range feeds.feeds {
		if !first {
			app(", ")
		}
		app(`"`, k, `"`, ": ")
		if v == nil {
			app("<nil>")
		} else {
			_, _ = fmt.Fprintf(&b, "%+v", *v)
		}
		first = false
	}
	app("]")

	return b.String()
}

func (feeds *Feeds) Len() int {
	return len(feeds.feeds)
}

func (feeds *Feeds) Contains(name string) bool {
	_, ok := feeds.feeds[name]
	return ok
}

func (feeds *Feeds) Set(name string, feed *Feed) {
	feeds.feeds[name] = feed
}

func (feeds *Feeds) Foreach(f func(*Feed)) {
	for _, feed := range feeds.feeds {
		f(feed)
	}
}

func (feeds *Feeds) ForeachGo(goFunc func(*Feed, *sync.WaitGroup)) {
	var wg sync.WaitGroup
	wg.Add(feeds.Len())

	for _, feed := range feeds.feeds {
		go goFunc(feed, &wg)
	}
	wg.Wait()
}

func (feed *Feed) NeedsUpdate(updateTime time.Time) bool {
	if !updateTime.IsZero() && int(time.Since(updateTime).Hours()) >= feed.MinFreq {
		log.Printf("Feed '%s' does not need updating, skipping.", feed.Name)
		return false
	}
	return true
}

func (feed *Feed) Success() bool {
	return feed.feed != nil
}
