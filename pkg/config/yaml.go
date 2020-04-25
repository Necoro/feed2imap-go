package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type config struct {
	*Config      `yaml:",inline"`
	GlobalConfig Map `yaml:",inline"` // need to be duplicated, because the Map in Config is not filled
	Feeds        []configGroupFeed
}

type group struct {
	Group string
	Feeds []configGroupFeed
}

type configGroupFeed struct {
	Target *string
	Feed   Feed  `yaml:",inline"`
	Group  group `yaml:",inline"`
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
	if grpFeed.Feed.Name != "" {
		return grpFeed.Feed.Name
	}

	return grpFeed.Group.Group
}

func unmarshal(buf []byte, cfg *Config) (config, error) {
	parsedCfg := config{Config: cfg}

	if err := yaml.Unmarshal(buf, &parsedCfg); err != nil {
		return config{}, err
	}
	//fmt.Printf("--- parsedCfg:\n%+v\n\n", parsedCfg)

	if parsedCfg.GlobalConfig == nil {
		cfg.GlobalConfig = Map{}
	} else {
		cfg.GlobalConfig = parsedCfg.GlobalConfig // need to copy the map explicitly
	}

	return parsedCfg, nil
}

func (cfg *Config) parse(buf []byte) error {
	var (
		err       error
		parsedCfg config
	)

	if parsedCfg, err = unmarshal(buf, cfg); err != nil {
		return fmt.Errorf("while unmarshalling: %w", err)
	}

	if err := buildFeeds(parsedCfg.Feeds, []string{}, cfg.Feeds); err != nil {
		return fmt.Errorf("while parsing: %w", err)
	}

	return nil
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

// Fetch the group structure and populate the `Target` fields in the feeds
func buildFeeds(cfg []configGroupFeed, target []string, feeds Feeds) error {
	for _, f := range cfg {
		target := appTarget(target, f.target())
		switch {
		case f.isFeed() && f.isGroup():
			return fmt.Errorf("Entry with Target %s is both a Feed and a group", target)

		case f.isFeed():
			feedCopy := f.Feed
			name := f.Feed.Name
			if name == "" {
				return fmt.Errorf("Unnamed feed")
			}

			if _, ok := feeds[name]; ok {
				return fmt.Errorf("Duplicate Feed Name '%s'", name)
			}
			feedCopy.Target = target
			feeds[name] = &feedCopy

		case f.isGroup():
			if err := buildFeeds(f.Group.Feeds, target, feeds); err != nil {
				return err
			}
		}
	}

	return nil
}
