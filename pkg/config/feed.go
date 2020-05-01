package config

// One stored feed
type Feed struct {
	Name    string
	Target  []string `yaml:"-"`
	Url     string
	Options `yaml:"-"` // not parsed directly
}

// Convenience type for all feeds
type Feeds map[string]*Feed
