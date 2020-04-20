package yaml

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"

	C "github.com/Necoro/feed2imap-go/internal/config"
	F "github.com/Necoro/feed2imap-go/internal/feed"
)

type config struct {
	C.GlobalOptions `yaml:",inline"`
	GlobalConfig    C.Map `yaml:",inline"`
	Feeds           []configGroupFeed
}

type group struct {
	Group string
	Feeds []configGroupFeed
}

type feed struct {
	Name      string
	Url       string
	C.Options `yaml:",inline"`
}

type configGroupFeed struct {
	Target *string
	Feed   feed  `yaml:",inline"`
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

func parse(buf []byte) (config, error) {
	parsedCfg := config{GlobalOptions: C.DefaultGlobalOptions}

	if err := yaml.Unmarshal(buf, &parsedCfg); err != nil {
		return config{}, fmt.Errorf("while unmarshalling: %w", err)
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
func buildFeeds(cfg []configGroupFeed, target []string, feeds F.Feeds) error {
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
			feeds[name] = &F.Feed{
				Name:    f.Feed.Name,
				Target:  target,
				Url:     f.Feed.Url,
				Options: f.Feed.Options,
			}

		case f.isGroup():
			if err := buildFeeds(f.Group.Feeds, target, feeds); err != nil {
				return err
			}
		}
	}

	return nil
}

func Load(path string) (C.Config, F.Feeds, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return C.Config{}, nil, fmt.Errorf("while reading '%s': %w", path, err)
	}

	var parsedCfg config
	if parsedCfg, err = parse(buf); err != nil {
		return C.Config{}, nil, err
	}

	feeds := F.Feeds{}

	if err := buildFeeds(parsedCfg.Feeds, []string{}, feeds); err != nil {
		return C.Config{}, nil, fmt.Errorf("while parsing: %w", err)
	}

	return C.Config{
		GlobalOptions: parsedCfg.GlobalOptions,
		GlobalConfig:  parsedCfg.GlobalConfig,
	}, feeds, nil
}
