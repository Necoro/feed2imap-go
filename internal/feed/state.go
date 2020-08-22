package feed

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/internal/feed/filter"
	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
)

type State struct {
	feeds map[string]*Feed
	cache Cache
	cfg   *config.Config
}

func (state *State) Foreach(f func(*Feed)) {
	for _, feed := range state.feeds {
		f(feed)
	}
}

func (state *State) ForeachGo(goFunc func(*Feed)) {
	var wg sync.WaitGroup
	wg.Add(len(state.feeds))

	f := func(feed *Feed, wg *sync.WaitGroup) {
		goFunc(feed)
		wg.Done()
	}

	for _, feed := range state.feeds {
		go f(feed, &wg)
	}
	wg.Wait()
}

func (state *State) LoadCache(fileName string, forceNew bool) error {
	var (
		cache Cache
		err   error
	)

	if forceNew {
		cache, err = newCache()
	} else {
		cache, err = LoadCache(fileName)
	}

	if err != nil {
		return err
	}
	state.cache = cache

	for _, feed := range state.feeds {
		feed.cached = cache.findItem(feed)
	}
	return nil
}

func (state *State) StoreCache(fileName string) error {
	return state.cache.store(fileName)
}

func (state *State) UnlockCache() {
	_ = state.cache.Unlock()
}

func (state *State) Fetch() int {
	state.ForeachGo(handleFeed)

	ctr := 0
	for _, feed := range state.feeds {
		success := feed.FetchSuccessful()
		feed.cached.Checked(!success)

		if success {
			ctr++
		}
	}

	return ctr
}

func printItem(item *gofeed.Item) string {
	// analogous to gofeed.Feed.String
	json, _ := json.MarshalIndent(item, "", "    ")
	return string(json)
}

func (feed *Feed) filterItems() []item {
	if feed.filter == nil {
		return feed.items
	}

	items := make([]item, 0, len(feed.items))

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

func filterFeed(feed *Feed) {
	if len(feed.items) > 0 {
		origLen := len(feed.items)

		log.Debugf("Filtering %s. Starting with %d items", feed.Name, origLen)

		items := feed.filterItems()
		newLen := len(items)
		if newLen < origLen {
			log.Printf("Item filter on %s: Reduced from %d to %d items.", feed.Name, origLen, newLen)
			origLen = newLen
		}

		items = feed.cached.filterItems(items, feed.IgnHash, feed.AlwaysNew)
		feed.items = items

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

func (state *State) Filter() {
	if log.IsDebug() {
		// single threaded for better output
		state.Foreach(filterFeed)
	} else {
		state.ForeachGo(filterFeed)
	}
}

func NewState(cfg *config.Config) (*State, error) {
	state := State{
		feeds: map[string]*Feed{},
		cache: Cache{}, // loaded later on
		cfg:   cfg,
	}

	for name, parsedFeed := range cfg.Feeds {
		var itemFilter *filter.Filter
		var err error
		if parsedFeed.ItemFilter != "" {
			if itemFilter, err = filter.New(parsedFeed.ItemFilter); err != nil {
				return nil, fmt.Errorf("Feed %s: Parsing item-filter: %w", parsedFeed.Name, err)
			}
		}
		state.feeds[name] = &Feed{Feed: parsedFeed, Global: cfg.GlobalOptions, filter: itemFilter}
	}

	return &state, nil
}

func (state *State) RemoveUndue() {
	for name, feed := range state.feeds {
		if feed.Disable || !feed.NeedsUpdate(feed.cached.Last()) {
			delete(state.feeds, name)
		}
	}
}

func (state *State) NumFeeds() int {
	return len(state.feeds)
}
