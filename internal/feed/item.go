package feed

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Necoro/gofeed"
	"github.com/google/uuid"

	"github.com/Necoro/feed2imap-go/pkg/config"
)

type feedImage struct {
	image []byte
	mime  string
	name  string
}

type ItemID uuid.UUID

func newItemID() ItemID {
	return ItemID(uuid.New())
}

type Item struct {
	*gofeed.Item              // access fields implicitly
	Feed         *gofeed.Feed // named explicitly to not shadow common fields with Item
	feed         *Feed
	Body         string
	TextBody     string
	UpdateOnly   bool
	ID           ItemID
	reasons      []string
	images       []feedImage
}

func (item *Item) DateParsed() *time.Time {
	if item.UpdatedParsed == nil || item.UpdatedParsed.IsZero() {
		return item.PublishedParsed
	}
	return item.UpdatedParsed
}

func (item *Item) Date() string {
	if item.Updated == "" {
		return item.Published
	}
	return item.Updated
}

// Creator returns the name of the creating authors (comma separated).
func (item *Item) Creator() string {
	names := make([]string, len(item.Authors))
	for i, p := range item.Authors {
		names[i] = p.Name
	}
	return strings.Join(names, ", ")
}

func (item *Item) FeedLink() string {
	if item.Feed.FeedLink != "" {
		// the one in the feed itself
		return item.Feed.FeedLink
	}
	// the one in the config
	return item.feed.Url
}

func (item *Item) AddReason(reason string) {
	if !slices.Contains(item.reasons, reason) {
		item.reasons = append(item.reasons, reason)
	}
}

func (item *Item) addImage(img []byte, mime string, name string) int {
	i := feedImage{img, mime, name}
	item.images = append(item.images, i)
	return len(item.images)
}

func (item *Item) clearImages() {
	clear(item.images)
	item.images = []feedImage{}
}

func (item *Item) defaultEmail() string {
	return item.feed.Global.DefaultEmail
}

func (item *Item) Id() string {
	idStr := base64.RawURLEncoding.EncodeToString(item.ID[:])
	return item.feed.id() + "#" + idStr
}

func (item *Item) messageId() string {
	return fmt.Sprintf("<feed#%s@%s>", item.Id(), config.Hostname())
}

func printItem(item *gofeed.Item) string {
	// analogous to gofeed.Feed.String
	json, _ := json.MarshalIndent(item, "", "    ")
	return string(json)
}
