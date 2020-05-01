package config

// One stored feed
type Feed struct {
	Name   string
	Target []string
	Url    string
	Options
}

// Convenience type for all feeds
type Feeds map[string]*Feed
