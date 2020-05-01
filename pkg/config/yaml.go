package config

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Necoro/feed2imap-go/pkg/log"
)

const (
	strTag   = "!!str"
	nullTag  = "!!null"
	emptyTag = ""
)

type config struct {
	*Config      `yaml:",inline"`
	GlobalConfig Map `yaml:",inline"`
	Feeds        []configGroupFeed
}

type group struct {
	Group string
	Feeds []configGroupFeed
}

type configGroupFeed struct {
	Target  yaml.Node
	Feed    Feed  `yaml:",inline"`
	Group   group `yaml:",inline"`
	Options Map   `yaml:",inline"`
}

func (grpFeed *configGroupFeed) isGroup() bool {
	return grpFeed.Group.Group != ""
}

func (grpFeed *configGroupFeed) isFeed() bool {
	return grpFeed.Feed.Name != "" || grpFeed.Feed.Url != ""
}

func (grpFeed *configGroupFeed) target() string {
	tag := grpFeed.Target.ShortTag()
	switch tag {
	case strTag:
		return grpFeed.Target.Value
	case nullTag:
		return ""
	case emptyTag:
		// tag not set: continue on
	default:
		panic("unexpected tag " + tag + " for target node")
	}

	if grpFeed.Feed.Name != "" {
		return grpFeed.Feed.Name
	}

	return grpFeed.Group.Group
}

func unmarshal(in io.Reader, cfg *Config) (config, error) {
	parsedCfg := config{Config: cfg}

	d := yaml.NewDecoder(in)
	if err := d.Decode(&parsedCfg); err != nil && err != io.EOF {
		return config{}, err
	}

	return parsedCfg, nil
}

func (cfg *Config) parse(in io.Reader) error {
	var (
		err       error
		parsedCfg config
	)

	if parsedCfg, err = unmarshal(in, cfg); err != nil {
		return fmt.Errorf("while unmarshalling: %w", err)
	}

	if err := buildFeeds(parsedCfg.Feeds, []string{}, cfg.Feeds, &cfg.FeedOptions); err != nil {
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

func buildOptions(globalFeedOptions *Options, options Map) (feedOptions Options, unknownFields []string) {
	if options == nil {
		// no options set for the feed: copy global options and be done
		return *globalFeedOptions, unknownFields
	}

	fv := reflect.ValueOf(&feedOptions).Elem()
	gv := reflect.ValueOf(globalFeedOptions).Elem()

	n := gv.NumField()
	for i := 0; i < n; i++ {
		val := fv.Field(i)
		f := fv.Type().Field(i)

		if f.PkgPath != "" && !f.Anonymous {
			continue
		}

		tag := f.Tag.Get("yaml")
		if tag == "" {
			continue
		}

		name := strings.Split(tag, ",")[0]

		set, ok := options[name]
		if ok { // in the map -> copy and delete
			val.Set(reflect.ValueOf(set))
			delete(options, name)
		} else { // not in the map -> copy from global
			val.Set(gv.Field(i))
		}
	}

	// remaining fields are unknown
	for k := range options {
		unknownFields = append(unknownFields, k)
	}

	return feedOptions, unknownFields
}

// Fetch the group structure and populate the `targetStr` fields in the feeds
func buildFeeds(cfg []configGroupFeed, target []string, feeds Feeds, globalFeedOptions *Options) error {
	for _, f := range cfg {
		target := appTarget(target, f.target())
		switch {
		case f.isFeed() && f.isGroup():
			return fmt.Errorf("Entry with targetStr %s is both a Feed and a group", target)

		case f.isFeed():
			feedCopy := f.Feed
			name := f.Feed.Name
			if name == "" {
				return fmt.Errorf("Unnamed feed")
			}
			if _, ok := feeds[name]; ok {
				return fmt.Errorf("Duplicate Feed Name '%s'", name)
			}

			opt, unknown := buildOptions(globalFeedOptions, f.Options)

			for _, optName := range unknown {
				log.Warnf("Unknown option '%s' for feed '%s'. Ignored!", optName, name)
			}

			feedCopy.Options = opt
			feedCopy.Target = target
			feeds[name] = &feedCopy

		case f.isGroup():
			if err := buildFeeds(f.Group.Feeds, target, feeds, globalFeedOptions); err != nil {
				return err
			}

			for optName := range f.Options {
				log.Warnf("Unknown option '%s' for group '%s'. Ignored!", optName, f.Group.Group)
			}
		}
	}

	return nil
}
