package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type config struct {
	GlobalConfig Map `yaml:",inline"`
	Feeds        []configGroupFeed
}

type Group struct {
	Group string
	Feeds []configGroupFeed
}

type configGroupFeed struct {
	Target *string
	Feed   `yaml:",inline"`
	Group  `yaml:",inline"`
}

func (grpFeed *configGroupFeed) isGroup() bool {
	return grpFeed.Group.Group != ""
}

func (grpFeed *configGroupFeed) isFeed() bool {
	return grpFeed.Feed.Name != "" || grpFeed.Feed.Url != ""
}

func (grpFeed *configGroupFeed) target() string {
	if grpFeed.Target != nil {
		return *grpFeed.Target
	}
	if grpFeed.Name != "" {
		return grpFeed.Name
	}

	return grpFeed.Group.Group
}

func parse(buf []byte) (config, error) {
	var parsedCfg config
	if err := yaml.Unmarshal(buf, &parsedCfg); err != nil {
		return parsedCfg, fmt.Errorf("while unmarshalling: %w", err)
	}
	//fmt.Printf("--- parsedCfg:\n%+v\n\n", parsedCfg)

	return parsedCfg, nil
}

func appTarget(target []string, app string) []string {
	switch {
	case len(target) == 0 && app == "":
		return []string{}
	case len(target) == 0:
		return []string{app}
	case app == "":
		return target
	default:
		return append(target, app)
	}
}

// Parse the group structure and populate the `Target` fields in the feeds
func buildFeeds(cfg []configGroupFeed, target []string, feeds Feeds) error {
	for idx := range cfg {
		f := &cfg[idx] // cannot use `_, f := range cfg` as it returns copies(!), but we need the originals
		target := appTarget(target, f.target())
		switch {
		case f.isFeed() && f.isGroup():
			return fmt.Errorf("Entry with Target %s is both a Feed and a group", target)

		case f.isFeed():
			name := f.Feed.Name
			if name == "" {
				return fmt.Errorf("Unnamed feed")
			}

			if _, ok := feeds[name]; ok {
				return fmt.Errorf("Duplicate Feed Name '%s'", name)
			}
			f.Feed.Target = target
			feeds[name] = &f.Feed

		case f.isGroup():
			if err := buildFeeds(f.Group.Feeds, target, feeds); err != nil {
				return err
			}
		}
	}

	return nil
}
