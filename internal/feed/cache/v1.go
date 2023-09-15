package cache

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Necoro/feed2imap-go/internal/feed"
	"github.com/Necoro/feed2imap-go/pkg/log"
	"github.com/Necoro/feed2imap-go/pkg/util"
)

const (
	v1Version    Version = 1
	startFeedId  uint64  = 1
	maxCacheSize         = 1000
	maxCacheDays         = 180
)

type feedId uint64

func (id feedId) String() string {
	return strconv.FormatUint(uint64(id), 16)
}

func idFromString(s string) feedId {
	id, _ := strconv.ParseUint(s, 16, 64)
	return feedId(id)
}

type v1Cache struct {
	Ids    map[feed.Descriptor]feedId
	NextId uint64
	Feeds  map[feedId]*cachedFeed
}

type cachedFeed struct {
	feed         *feed.Feed
	id           feedId // not saved, has to be set on loading
	LastCheck    time.Time
	currentCheck time.Time
	NumFailures  int // can't be named `Failures` b/c it'll collide with the interface
	Items        []cachedItem
	newItems     []cachedItem
}

type itemHash [sha256.Size]byte

func (h itemHash) String() string {
	return hex.EncodeToString(h[:])
}

type cachedItem struct {
	Guid         string
	Title        string
	Link         string
	Date         time.Time
	UpdatedCache time.Time
	Hash         itemHash
	ID           uuid.UUID
	deleted      bool
}

func (item cachedItem) String() string {
	return fmt.Sprintf(`{
  ID: %s
  Title: %q
  Guid: %q
  Link: %q
  Date: %s
  Hash: %s
}`,
		base64.RawURLEncoding.EncodeToString(item.ID[:]),
		item.Title, item.Guid, item.Link, util.TimeFormat(item.Date), item.Hash)
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
	if cf.newItems != nil {
		cf.Items = cf.newItems
		cf.newItems = nil
	}
	cf.LastCheck = cf.currentCheck
}

func (cf *cachedFeed) Failures() int {
	return cf.NumFailures
}

func (cf *cachedFeed) Last() time.Time {
	return cf.LastCheck
}

func (cf *cachedFeed) Feed() *feed.Feed {
	return cf.feed
}

func (cache *v1Cache) Version() Version {
	return v1Version
}

func (cache *v1Cache) Info() string {
	descriptors := make([]feed.Descriptor, len(cache.Ids))
	i := 0
	for descr := range cache.Ids {
		descriptors[i] = descr
		i++
	}

	sort.Slice(descriptors, func(i, j int) bool {
		return descriptors[i].Name < descriptors[j].Name
	})

	b := strings.Builder{}
	for _, descr := range descriptors {
		id := cache.Ids[descr]
		feed := cache.Feeds[id]
		b.WriteString(fmt.Sprintf("%3s: %s (%s) (%d items)\n", id.String(), descr.Name, descr.Url, len(feed.Items)))
	}
	return b.String()
}

func (cache *v1Cache) SpecificInfo(i any) string {
	id := idFromString(i.(string))

	b := strings.Builder{}
	feed := cache.Feeds[id]

	for descr, fId := range cache.Ids {
		if id == fId {
			b.WriteString(descr.Name)
			b.WriteString(" -- ")
			b.WriteString(descr.Url)
			b.WriteByte('\n')
			break
		}
	}

	b.WriteString(fmt.Sprintf(`
Last Check: %s
Num Failures: %d
Num Items: %d
`,
		util.TimeFormat(feed.LastCheck),
		feed.NumFailures,
		len(feed.Items)))

	for _, item := range feed.Items {
		b.WriteString("\n--------------------\n")
		b.WriteString(item.String())
	}
	return b.String()
}

func newV1Cache() *v1Cache {
	cache := v1Cache{
		Ids:    map[feed.Descriptor]feedId{},
		Feeds:  map[feedId]*cachedFeed{},
		NextId: startFeedId,
	}
	return &cache
}

func (cache *v1Cache) transformTo(v Version) (Impl, error) {
	switch v {
	case v2Version:
		return (*v2Cache)(cache), nil
	default:
		return nil, fmt.Errorf("Transformation not supported")
	}
}

func (cache *v1Cache) getItem(id feedId) *cachedFeed {
	feed, ok := cache.Feeds[id]
	if !ok {
		feed = &cachedFeed{}
		cache.Feeds[id] = feed
	}
	feed.id = id
	return feed
}

func (cache *v1Cache) cachedFeed(f *feed.Feed) CachedFeed {
	fDescr := f.Descriptor()
	id, ok := cache.Ids[fDescr]
	if !ok {
		var otherId feed.Descriptor
		changed := false
		for otherId, id = range cache.Ids {
			if otherId.Name == fDescr.Name {
				log.Warnf("Feed %s seems to have changed URLs: new '%s', old '%s'. Updating.",
					fDescr.Name, fDescr.Url, otherId.Url)
				changed = true
				break
			} else if otherId.Url == fDescr.Url {
				log.Warnf("Feed with URL '%s' seems to have changed its name: new '%s', old '%s'. Updating.",
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

	cf := cache.getItem(id)
	cf.feed = f
	f.SetExtID(id)
	return cf
}

func (cf *cachedFeed) cachedItem(item *feed.Item) cachedItem {
	var ci cachedItem

	ci.ID = item.ID
	ci.Title = item.Item.Title
	ci.Link = item.Item.Link
	if item.DateParsed() != nil {
		ci.Date = *item.DateParsed()
	}
	ci.Guid = item.Item.GUID

	contentByte := []byte(item.Item.Description + item.Item.Content)
	ci.Hash = sha256.Sum256(contentByte)

	return ci
}

func (item *cachedItem) similarTo(other *cachedItem, ignoreHash bool) bool {
	return other.Title == item.Title &&
		other.Link == item.Link &&
		other.Date.Equal(item.Date) &&
		(ignoreHash || other.Hash == item.Hash)
}

func (cf *cachedFeed) markItemDeleted(index int) {
	cf.Items[index].deleted = true
}

func (cf *cachedFeed) Filter(items []feed.Item, ignoreHash, alwaysNew bool) []feed.Item {
	if len(items) == 0 {
		return items
	}

	cacheItems := make(map[cachedItem]*feed.Item, len(items))
	for idx := range items {
		i := &items[idx]
		ci := cf.cachedItem(i)

		// remove complete duplicates on the go
		cacheItems[ci] = i
	}
	log.Debugf("%d items after deduplication", len(cacheItems))

	filtered := make([]feed.Item, 0, len(items))
	cacheadd := make([]cachedItem, 0, len(items))
	app := func(item *feed.Item, ci cachedItem, oldIdx *int) {
		if oldIdx != nil {
			item.UpdateOnly = true
			prevId := cf.Items[*oldIdx].ID
			ci.ID = prevId
			item.ID = prevId
			log.Debugf("oldIdx: %d, prevId: %s, item.id: %s", *oldIdx, prevId, item.Id())
			cf.markItemDeleted(*oldIdx)
		}
		filtered = append(filtered, *item)
		cacheadd = append(cacheadd, ci)
	}

	seen := func(oldIdx int) {
		ci := cf.Items[oldIdx]
		cf.markItemDeleted(oldIdx)
		cacheadd = append(cacheadd, ci)
	}

CACHE_ITEMS:
	for ci, item := range cacheItems {
		log.Debugf("Now checking %s", ci)

		if ci.Guid != "" {
			for idx, oldItem := range cf.Items {
				if oldItem.Guid == ci.Guid {
					log.Debugf("Guid matches with: %s", oldItem)
					if !oldItem.similarTo(&ci, ignoreHash) {
						item.AddReason("guid (upd)")
						app(item, ci, &idx)
					} else {
						log.Debugf("Similar, ignoring item %s", base64.RawURLEncoding.EncodeToString(oldItem.ID[:]))
						seen(idx)
					}

					continue CACHE_ITEMS
				}
			}

			log.Debug("Found no matching GUID, including.")
			item.AddReason("guid")
			app(item, ci, nil)
			continue
		}

		for idx, oldItem := range cf.Items {
			if oldItem.similarTo(&ci, ignoreHash) {
				log.Debugf("Similarity matches, ignoring: %s", oldItem)
				seen(idx)
				continue CACHE_ITEMS
			}

			if oldItem.Link == ci.Link {
				if alwaysNew {
					log.Debugf("Link matches, but `always-new`.")
					item.AddReason("always-new")
					continue
				}
				log.Debugf("Link matches, updating: %s", oldItem)
				item.AddReason("link (upd)")
				app(item, ci, &idx)

				continue CACHE_ITEMS
			}
		}

		log.Debugf("No match found, inserting.")
		item.AddReason("new")
		app(item, ci, nil)
	}

	log.Debugf("%d items after filtering", len(filtered))

	cf.newItems = append(cacheadd, filterItems(cf.Items)...)

	return filtered
}

func filterItems(items []cachedItem) []cachedItem {
	n := min(len(items), maxCacheSize)

	copiedItems := make([]cachedItem, 0, n)
	for _, item := range items {
		if !item.deleted {
			copiedItems = append(copiedItems, item)
			if len(copiedItems) >= n {
				break
			}
		}
	}

	return copiedItems
}

func (cache *v1Cache) cleanup(knownDescriptors map[feed.Descriptor]struct{}) {
	for descr, id := range cache.Ids {
		if _, ok := knownDescriptors[descr]; ok {
			// do not delete stuff still known to us
			continue
		}

		cf := cache.Feeds[id]
		if cf.LastCheck.IsZero() || util.Days(time.Since(cf.LastCheck)) > maxCacheDays {
			delete(cache.Feeds, id)
			delete(cache.Ids, descr)
		}
	}
}

func (cache *v1Cache) load(reader io.Reader) error {
	decoder := gob.NewDecoder(reader)
	return decoder.Decode(cache)
}

func (cache *v1Cache) store(writer io.Writer) error {
	encoder := gob.NewEncoder(writer)
	return encoder.Encode(cache)
}
