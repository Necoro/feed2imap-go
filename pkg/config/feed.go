package config

// One stored feed
type Feed struct {
	Name    string
	Target  []string `yaml:"-"`
	Url     string
	Options `yaml:",inline"`
}

// Convenience type for all feeds
type Feeds map[string]*Feed
