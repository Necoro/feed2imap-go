package feed

import (
	"crypto/sha256"
	"time"

	"github.com/Necoro/feed2imap-go/pkg/log"
)

const (
	v1Version   Version = 1
	startFeedId uint64  = 1
)

type feedId uint64

type v1Cache struct {
	Ids    map[feedDescriptor]feedId
	NextId uint64
	Feeds  map[feedId]*cachedFeed
}

type cachedFeed struct {
	LastCheck   time.Time
	NumFailures uint // can't be named `Failures` b/c it'll collide with the interface
	Items       []cachedItem
}

type itemHash [sha256.Size]byte

type cachedItem struct {
	Uid     string
	Title   string
	Link    string
	Date    time.Time
	Updated time.Time
	Creator string
	Hash    itemHash
}

func (cf *cachedFeed) Checked(withFailure bool) {
	cf.LastCheck = time.Now()
	if withFailure {
		cf.NumFailures++
	} else {
		cf.NumFailures = 0
	}
}

func (cf *cachedFeed) Failures() uint {
	return cf.NumFailures
}

func (cf *cachedFeed) Last() time.Time {
	return cf.LastCheck
}

func (cache *v1Cache) Version() Version {
	return v1Version
}

func newV1Cache() *v1Cache {
	cache := v1Cache{
		Ids:    map[feedDescriptor]feedId{},
		Feeds:  map[feedId]*cachedFeed{},
		NextId: startFeedId,
	}
	return &cache
}

func (cache *v1Cache) transformToCurrent() (Cache, error) {
	return cache, nil
}

func (cache *v1Cache) getItem(id feedId) CachedFeed {
	feed, ok := cache.Feeds[id]
	if !ok {
		feed = &cachedFeed{}
		cache.Feeds[id] = feed
	}
	return feed
}

func (cache *v1Cache) findItem(feed *Feed) CachedFeed {
	if feed.cached != nil {
		return feed.cached.(*cachedFeed)
	}

	fDescr := feed.descriptor()
	id, ok := cache.Ids[fDescr]
	if !ok {
		var otherId feedDescriptor
		changed := false
		for otherId, id = range cache.Ids {
			if otherId.Name == fDescr.Name {
				log.Warnf("Feed %s seems to have changed URLs: newCache '%s', old '%s'. Updating.",
					fDescr.Name, fDescr.Url, otherId.Url)
				changed = true
				break
			} else if otherId.Url == fDescr.Url {
				log.Warnf("Feed with URL '%s' seems to have changed its name: newCache '%s', old '%s'. Updating",
					fDescr.Url, fDescr.Name, otherId.Name)
				changed = true
				break
			}
		}
		if changed {
			delete(cache.Ids, otherId)
		} else {
			id = feedId(cache.NextId)
			cache.NextId++
		}

		cache.Ids[fDescr] = id
	}

	item := cache.getItem(id)
	feed.cached = item
	return item
}
