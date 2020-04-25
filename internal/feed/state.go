package feed

import (
	"sync"

	"github.com/Necoro/feed2imap-go/pkg/config"
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

func (state *State) ForeachGo(goFunc func(*Feed, *sync.WaitGroup)) {
	var wg sync.WaitGroup
	wg.Add(len(state.feeds))

	for _, feed := range state.feeds {
		go goFunc(feed, &wg)
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
		success := feed.Success()
		feed.cached.Checked(!success)

		if success {
			ctr++
		}
	}

	return ctr
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
		if !feed.NeedsUpdate(feed.cached.Last()) {
			delete(state.feeds, name)
		}
	}
}

func (state *State) NumFeeds() int {
	return len(state.feeds)
}
