package feed

import (
	"bufio"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Necoro/feed2imap-go/pkg/log"
)

const (
	currentVersion byte   = 1
	startFeedId    uint64 = 1
)

type Cache interface {
	findItem(*Feed) CachedFeed
	Version() byte
	transformToCurrent() (Cache, error)
}

type feedId uint64

type feedDescriptor struct {
	Name string
	Url  string
}

type CachedFeed interface {
	Checked(withFailure bool)
	Failures() uint
}

type cachedFeed struct {
	LastCheck   time.Time
	NumFailures uint // can't be named `Failures` b/c it'll collide with the interface
	Items       []cachedItem
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

type v1Cache struct {
	version byte
	Ids     map[feedDescriptor]feedId
	NextId  uint64
	Feeds   map[feedId]*cachedFeed
}

func (cache *v1Cache) Version() byte {
	return cache.version
}

func newCache() Cache {
	cache := v1Cache{
		Ids:    map[feedDescriptor]feedId{},
		Feeds:  map[feedId]*cachedFeed{},
		NextId: startFeedId,
	}
	cache.version = currentVersion
	return &cache
}

func cacheForVersion(version byte) (Cache, error) {
	switch version {
	case 1:
		return newCache(), nil
	default:
		return nil, fmt.Errorf("unknown cache version '%d'", version)
	}
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

func storeCache(cache Cache, fileName string) error {
	if cache == nil {
		return fmt.Errorf("trying to store nil cache")
	}
	if cache.Version() != currentVersion {
		return fmt.Errorf("trying to store cache with unsupported version '%d' (current: '%d')", cache.Version(), currentVersion)
	}

	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("trying to store cache to '%s': %w", fileName, err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	if err = writer.WriteByte(currentVersion); err != nil {
		return fmt.Errorf("writing to '%s': %w", fileName, err)
	}

	encoder := gob.NewEncoder(writer)
	if err = encoder.Encode(cache); err != nil {
		return fmt.Errorf("encoding cache: %w", err)
	}

	writer.Flush()
	log.Printf("Stored cache to '%s'.", fileName)

	return nil
}

func loadCache(fileName string) (Cache, error) {
	f, err := os.Open(fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// no cache there yet -- make new
			return newCache(), nil
		}
		return nil, fmt.Errorf("opening cache at '%s': %w", fileName, err)
	}
	defer f.Close()

	log.Printf("Loading cache from '%s'", fileName)

	reader := bufio.NewReader(f)
	version, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("reading from '%s': %w", fileName, err)
	}

	cache, err := cacheForVersion(version)
	if err != nil {
		return nil, err
	}

	decoder := gob.NewDecoder(reader)
	if err = decoder.Decode(cache); err != nil {
		return nil, fmt.Errorf("decoding for version '%d' from '%s': %w", version, fileName, err)
	}

	if cache, err = cache.transformToCurrent(); err != nil {
		return nil, fmt.Errorf("cannot transform from version %d to %d: %w", version, currentVersion, err)
	}

	log.Printf("Loaded cache (version %d), transformed to version %d.", version, currentVersion)

	return cache, nil
}
