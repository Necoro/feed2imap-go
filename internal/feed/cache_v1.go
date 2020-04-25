package feed

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/Necoro/feed2imap-go/pkg/log"
	"github.com/Necoro/feed2imap-go/pkg/util"
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
	LastCheck    time.Time
	currentCheck time.Time
	NumFailures  uint // can't be named `Failures` b/c it'll collide with the interface
	Items        []cachedItem
	newItems     []cachedItem
}

type itemHash [sha256.Size]byte

type cachedItem struct {
	Guid          string
	Title         string
	Link          string
	PublishedDate time.Time
	UpdatedDate   time.Time
	UpdatedCache  time.Time
	Hash          itemHash
}

func (item cachedItem) String() string {
	return fmt.Sprintf(`{
  Title: %q
  Guid: %q
  Link: %q
  Published: %s
  Updated: %s
}`, item.Title, item.Guid, item.Link, util.TimeFormat(item.PublishedDate), util.TimeFormat(item.UpdatedDate))
}

func (cf *cachedFeed) Checked(withFailure bool) {
	cf.currentCheck = time.Now()
	if withFailure {
		cf.NumFailures++
	} else {
		cf.NumFailures = 0
	}
}

func (cf *cachedFeed) Commit() {
	cf.Items = cf.newItems
	cf.newItems = nil
	cf.LastCheck = cf.currentCheck
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

func newCachedItem(item feeditem) cachedItem {
	var ci cachedItem

	ci.Title = item.Item.Title
	ci.Link = item.Item.Link
	if item.Item.PublishedParsed != nil {
		ci.PublishedDate = *item.Item.PublishedParsed
	}
	if item.Item.UpdatedParsed != nil && !item.Item.UpdatedParsed.Equal(ci.PublishedDate) {
		ci.UpdatedDate = *item.Item.UpdatedParsed
	}
	ci.Guid = item.Item.GUID

	contentByte := []byte(item.Item.Description + item.Item.Content)
	ci.Hash = sha256.Sum256(contentByte)

	return ci
}

func (item *cachedItem) similarTo(other *cachedItem, ignoreHash bool) bool {
	return other.Title == item.Title ||
		other.Link == item.Link ||
		other.PublishedDate.Equal(item.PublishedDate) ||
		(!ignoreHash && other.Hash == item.Hash)
}

func (cf *cachedFeed) deleteItem(index int) {
	copy(cf.Items[index:], cf.Items[index+1:])
	cf.Items[len(cf.Items)-1] = cachedItem{}
	cf.Items = cf.Items[:len(cf.Items)-1]
}

func (cf *cachedFeed) filterItems(items []feeditem) []feeditem {
	if len(items) == 0 {
		return items
	}

	cacheItems := make(map[cachedItem]*feeditem, len(items))
	for idx := range items {
		// remove complete duplicates on the go
		cacheItems[newCachedItem(items[idx])] = &items[idx]
	}
	log.Debugf("%d items after deduplication", len(cacheItems))

	filtered := make([]feeditem, 0, len(items))
	cacheadd := make([]cachedItem, 0, len(items))
	app := func(item *feeditem, ci cachedItem, oldIdx *int) {
		if oldIdx != nil {
			item.updateOnly = true
			cf.deleteItem(*oldIdx)
		}
		filtered = append(filtered, *item)
		cacheadd = append(cacheadd, ci)
	}

CACHE_ITEMS:
	for ci, item := range cacheItems {
		log.Debugf("Now checking %s", ci)
		if cf.LastCheck.IsZero() || ci.PublishedDate.After(cf.LastCheck) {
			log.Debug("Newer than last check, including.")

			item.addReason("newer")
			app(item, ci, nil)
			continue
		}

		if ci.Guid != "" {
			for idx, oldItem := range cf.Items {
				if oldItem.Guid == ci.Guid {
					log.Debugf("Guid matches with: %s", oldItem)
					if !oldItem.similarTo(&ci, false) {
						item.addReason("guid (upd)")
						app(item, ci, &idx)
					} else {
						log.Debugf("Similar, ignoring")
					}

					continue CACHE_ITEMS
				}
			}

			log.Debug("Found no matching GUID, including.")
			item.addReason("guid")
			app(item, ci, nil)
			continue
		}

		for idx, oldItem := range cf.Items {
			if oldItem.similarTo(&ci, false) {
				log.Debugf("Similarity matches, ignoring: %s", oldItem)
				continue CACHE_ITEMS
			}

			if oldItem.Link == ci.Link {
				log.Debugf("Link matches, updating: %s", oldItem)
				item.addReason("link (upd)")
				app(item, ci, &idx)

				continue CACHE_ITEMS
			}
		}

		log.Debugf("No match found, inserting.")
		app(item, ci, nil)
	}

	log.Debugf("%d items after filtering", len(filtered))

	cf.newItems = append(cacheadd, cf.Items...)

	return filtered
}
