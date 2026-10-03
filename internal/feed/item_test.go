package feed

import (
	"slices"
	"testing"

	"github.com/Necoro/gofeed"
	ext "github.com/Necoro/gofeed/extensions"
)

func TestMediaImageURLs(t *testing.T) {
	mc := func(url, attr, val string) ext.Extension {
		return ext.Extension{Name: "content", Attrs: map[string]string{"url": url, attr: val}}
	}
	item := &Item{Item: &gofeed.Item{Extensions: ext.Extensions{"media": {
		"content": {mc("http://x/top.jpg", "medium", "image"), mc("http://x/v.mp4", "type", "video/mp4")},
		"group": {{Name: "group", Children: map[string][]ext.Extension{
			"content": {mc("http://x/a.jpg", "medium", "image"), mc("http://x/b.jpg", "type", "image/jpeg"), mc("http://x/top.jpg", "medium", "image")},
		}}},
	}}}}

	want := []string{"http://x/top.jpg", "http://x/a.jpg", "http://x/b.jpg"}
	if got := item.mediaImageURLs(); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
