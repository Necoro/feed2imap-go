package feed

import (
	"fmt"

	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/util"
)

type feedImage struct {
	image []byte
	mime  string
}

type item struct {
	*gofeed.Feed
	*gofeed.Item
	feed       *Feed
	Body       string
	updateOnly bool
	reasons    []string
	images     []feedImage
	itemId     string
}

// Creator returns the name of the creating author.
// MUST NOT have `*item` has the receiver, because the template breaks then.
func (item *item) Creator() string {
	if item.Item.Author != nil {
		return item.Item.Author.Name
	}
	return ""
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

func (item *item) messageId() string {
	return fmt.Sprintf("<feed#%s#%s@%s>", item.feed.cached.ID(), item.itemId, config.Hostname())
}
