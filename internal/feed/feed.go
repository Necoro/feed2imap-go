package feed

import (
	"fmt"
	"strings"

	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/internal/config"
)

type Feed struct {
	Name   string
	Target []string
	Url    string
	config.Options
	feed gofeed.Feed
}

type Feeds map[string]*Feed

func (f Feeds) String() string {
	var b strings.Builder
	app := func(a ...interface{}) {
		_, _ = fmt.Fprint(&b, a...)
	}
	app("Feeds [")

	first := true
	for k, v := range f {
		if !first {
			app(", ")
		}
		app(`"`, k, `"`, ": ")
		if v == nil {
			app("<nil>")
		} else {
			_, _ = fmt.Fprintf(&b, "%+v", *v)
		}
		first = false
	}
	app("]")

	return b.String()
}
