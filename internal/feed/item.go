package feed

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/util"
)

type feedImage struct {
	image []byte
	mime  string
}

type item struct {
	*gofeed.Item
	Feed       *gofeed.Feed
	feed       *Feed
	Body       string
	updateOnly bool
	reasons    []string
	images     []feedImage
	itemId     uuid.UUID
}

func (item *item) DateParsed() *time.Time {
	if item.UpdatedParsed == nil || item.UpdatedParsed.IsZero() {
		return item.PublishedParsed
	}
	return item.UpdatedParsed
}

func (item *item) Date() string {
	if item.Updated == "" {
		return item.Published
	}
	return item.Updated
}

// Creator returns the name of the creating author.
func (item *item) Creator() string {
	if item.Author != nil {
		return item.Author.Name
	}
	return ""
}

func (item *item) FeedLink() string {
	if item.Feed.Link != "" {
		// the one in the feed itself
		return item.Feed.FeedLink
	}
	// the one in the config
	return item.feed.Url
}

func (item *item) addReason(reason string) {
	if !util.StrContains(item.reasons, reason) {
		item.reasons = append(item.reasons, reason)
	}
}

func (item *item) addImage(img []byte, mime string) int {
	i := feedImage{img, mime}
	item.images = append(item.images, i)
	return len(item.images)
}

func (item *item) clearImages() {
	item.images = []feedImage{}
}

func (item *item) defaultEmail() string {
	return item.feed.Global.DefaultEmail
}

func (item *item) id() string {
	idStr := base64.RawURLEncoding.EncodeToString(item.itemId[:])
	return item.feed.cached.ID() + "#" + idStr
}

func (item *item) messageId() string {
	return fmt.Sprintf("<feed#%s@%s>", item.id(), config.Hostname())
}
