package cache

import (
	"sync"

	"github.com/Necoro/feed2imap-go/internal/feed"
	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
)

type State struct {
	feeds       map[string]*feed.Feed
	cachedFeeds map[string]CachedFeed
	cache       Cache
	cfg         *config.Config
}

func (state *State) Foreach(f func(CachedFeed)) {
	for _, feed := range state.cachedFeeds {
		f(feed)
	}
}

func (state *State) ForeachGo(goFunc func(CachedFeed)) {
	var wg sync.WaitGroup
	wg.Add(len(state.feeds))

	f := func(feed CachedFeed, wg *sync.WaitGroup) {
		goFunc(feed)
		wg.Done()
	}

	for _, feed := range state.cachedFeeds {
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

	for name, feed := range state.feeds {
		state.cachedFeeds[name] = cache.cachedFeed(feed)
	}

	// state.feeds should not be used after loading the cache --> enforce a panic
	state.feeds = nil

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
	for _, cf := range state.cachedFeeds {
		success := cf.Feed().FetchSuccessful()
		cf.Checked(!success)

		if success {
			ctr++
		}
	}

	return ctr
}

func handleFeed(cf CachedFeed) {
	feed := cf.Feed()
	log.Printf("Fetching %s from %s", feed.Name, feed.Url)

	err := feed.Parse()
	if err != nil {
		if feed.Url == "" || cf.Failures() >= feed.Global.MaxFailures {
			log.Error(err)
		} else {
			log.Print(err)
		}
	}
}

func filterFeed(cf CachedFeed) {
	cf.Feed().Filter(cf.Filter)
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
		feeds:       map[string]*feed.Feed{},
		cachedFeeds: map[string]CachedFeed{},
		cache:       Cache{}, // loaded later on
		cfg:         cfg,
	}

	for name, parsedFeed := range cfg.Feeds {
		feed, err := feed.Create(parsedFeed, cfg.GlobalOptions)
		if err != nil {
			return nil, err
		}
		state.feeds[name] = feed
	}

	return &state, nil
}

func (state *State) RemoveUndue() {
	for name, feed := range state.cachedFeeds {
		if feed.Feed().Disable || !feed.Feed().NeedsUpdate(feed.Last()) {
			delete(state.cachedFeeds, name)
		}
	}
}

func (state *State) NumFeeds() int {
	return len(state.cachedFeeds)
}
