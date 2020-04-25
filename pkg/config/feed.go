package config

import (
	"fmt"
	"strings"
)

// One stored feed
type Feed struct {
	Name    string
	Target  []string `yaml:"-"`
	Url     string
	Options `yaml:",inline"`
}

// Convenience type for all feeds
type Feeds map[string]Feed

func (feeds Feeds) String() string {
	var b strings.Builder
	app := func(a ...interface{}) {
		_, _ = fmt.Fprint(&b, a...)
	}
	app("Feeds [")

	first := true
	for k, v := range feeds {
		if !first {
			app(", ")
		}
		app(`"`, k, `"`, ": ")
		_, _ = fmt.Fprintf(&b, "%+v", v)
		first = false
	}
	app("]")

	return b.String()
}
