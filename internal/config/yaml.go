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
	return grpFeed.Name != ""
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
	fmt.Printf("--- parsedCfg:\n%+v\n\n", parsedCfg)

	return parsedCfg, nil
}

func appTarget(target, app string) string {
	if target == "" {
		return app
	}

	if app == "" {
		return target
	}

	return target + "/" + app
}

// Parse the group structure and populate the `Target` fields in the feeds
func buildFeeds(cfg []configGroupFeed, target string, feeds Feeds) error {
	for _, f := range cfg {
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
