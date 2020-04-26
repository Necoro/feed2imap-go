package feed

import (
	"sync"

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

func (state *State) LoadCache(fileName string) error {
	cache, err := loadCache(fileName)
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
	return storeCache(state.cache, fileName)
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

func filterFeed(feed *Feed) {
	if len(feed.items) > 0 {
		origLen := len(feed.items)

		log.Debugf("Filtering %s. Starting with %d items", feed.Name, origLen)
		items := feed.cached.filterItems(feed.items, *feed.Options.IgnHash, *feed.Options.AlwaysNew)
		feed.items = items

		newLen := len(feed.items)
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

func NewState(cfg *config.Config) *State {
	state := State{
		feeds: map[string]*Feed{},
		cache: nil, // loaded later on
		cfg:   cfg,
	}

	for name, parsedFeed := range cfg.Feeds {
		state.feeds[name] = &Feed{Feed: parsedFeed}
	}

	return &state
}

func (state *State) RemoveUndue() {
	for name, feed := range state.feeds {
		if *feed.Options.Disable || !feed.NeedsUpdate(feed.cached.Last()) {
			delete(state.feeds, name)
		}
	}
}

func (state *State) NumFeeds() int {
	return len(state.feeds)
}
