package config

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Necoro/feed2imap-go/pkg/log"
)

const (
	strTag  = "!!str"
	nullTag = "!!null"
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

type feed struct {
	Name string
	Url  string
	Exec []string
}

type configGroupFeed struct {
	Target  yaml.Node
	Feed    feed  `yaml:",inline"`
	Group   group `yaml:",inline"`
	Options Map   `yaml:",inline"`
}

func (grpFeed *configGroupFeed) isGroup() bool {
	return grpFeed.Group.Group != ""
}

func (grpFeed *configGroupFeed) isFeed() bool {
	return grpFeed.Feed.Name != "" || grpFeed.Feed.Url != "" || len(grpFeed.Feed.Exec) > 0
}

func (grpFeed *configGroupFeed) target(autoTarget bool) string {
	if !autoTarget || !grpFeed.Target.IsZero() {
		if grpFeed.Target.ShortTag() == nullTag {
			// null may be represented by ~ or NULL or ...
			// Value would hold this representation, which we do not want
			return ""
		}
		return grpFeed.Target.Value
	}

	if grpFeed.Feed.Name != "" {
		return grpFeed.Feed.Name
	}

	return grpFeed.Group.Group
}

func unmarshal(in io.Reader, cfg *Config) (config, error) {
	parsedCfg := config{Config: cfg}

	d := yaml.NewDecoder(in)
	d.KnownFields(true)
	if err := d.Decode(&parsedCfg); err != nil && err != io.EOF {
		return config{}, err
	}

	return parsedCfg, nil
}

func (cfg *Config) fixGlobalOptions(unparsed Map) {
	origMap := Map{}

	// copy map
	for k, v := range unparsed {
		origMap[k] = v
	}

	newOpts, _ := buildOptions(&cfg.FeedOptions, unparsed)

	for k, v := range origMap {
		if _, ok := unparsed[k]; !ok {
			log.Warnf("Global option '%s' should be inside the 'options' map. It currently overwrites the same key there.", k)
		} else if !handleDeprecated(k, v, "", &cfg.GlobalOptions, &newOpts) {
			log.Warnf("Unknown global option '%s'. Ignored!", k)
		}
	}

	cfg.FeedOptions = newOpts
}

func (cfg *Config) parse(in io.Reader) error {
	var (
		err       error
		parsedCfg config
	)

	if parsedCfg, err = unmarshal(in, cfg); err != nil {
		var typeError *yaml.TypeError
		if errors.As(err, &typeError) {
			const sep = "\n\t"
			errMsgs := strings.Join(typeError.Errors, sep)
			return fmt.Errorf("config is invalid: %s%s", sep, errMsgs)
		}

		return fmt.Errorf("while unmarshalling: %w", err)
	}

	cfg.fixGlobalOptions(parsedCfg.GlobalConfig)

	if err := buildFeeds(parsedCfg.Feeds, []string{}, cfg.Feeds, &cfg.FeedOptions, cfg.AutoTarget, &cfg.Target); err != nil {
		return fmt.Errorf("while parsing: %w", err)
	}

	return nil
}

func appTarget(target []string, app string) []string {
	app = strings.TrimSpace(app)
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
		field := fv.Field(i)
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
			value := reflect.ValueOf(set)
			if !value.Type().AssignableTo(field.Type()) {
				value = value.Convert(field.Type())
			}
			field.Set(value)
			delete(options, name)
		} else { // not in the map -> copy from global
			field.Set(gv.Field(i))
		}
	}

	// remaining fields are unknown
	for k := range options {
		unknownFields = append(unknownFields, k)
	}

	return feedOptions, unknownFields
}

// Fetch the group structure and populate the `targetStr` fields in the feeds
func buildFeeds(cfg []configGroupFeed, target []string, feeds Feeds,
	globalFeedOptions *Options, autoTarget bool, globalTarget *Url) (err error) {

	for _, f := range cfg {
		var fTarget []string

		rawTarget := f.target(autoTarget)
		if isRecognizedUrl(rawTarget) {
			// deprecated old-style URLs as target
			// as the full path is specified, `target` is ignored
			if fTarget, err = handleUrlTarget(rawTarget, &f.Target, globalTarget); err != nil {
				return err
			}
		} else {
			// new-style tree-like structure
			fTarget = appTarget(target, rawTarget)
		}

		switch {
		case f.isFeed() && f.isGroup():
			return fmt.Errorf("Entry with targetStr %s is both a Feed and a group", fTarget)

		case f.isFeed():
			name := f.Feed.Name
			if name == "" {
				return fmt.Errorf("Unnamed feed")
			}
			if _, ok := feeds[name]; ok {
				return fmt.Errorf("Duplicate Feed Name '%s'", name)
			}

			opt, unknown := buildOptions(globalFeedOptions, f.Options)

			for _, optName := range unknown {
				if !handleDeprecated(optName, f.Options[optName], name, nil, &opt) {
					log.Warnf("Unknown option '%s' for feed '%s'. Ignored!", optName, name)
				}
			}

			feeds[name] = &Feed{
				Name:    name,
				Url:     f.Feed.Url,
				Exec:    f.Feed.Exec,
				Options: opt,
				Target:  fTarget,
			}

		case f.isGroup():
			opt, unknown := buildOptions(globalFeedOptions, f.Options)

			for _, optName := range unknown {
				log.Warnf("Unknown option '%s' for group '%s'. Ignored!", optName, f.Group.Group)
			}

			if err = buildFeeds(f.Group.Feeds, fTarget, feeds, &opt, autoTarget, globalTarget); err != nil {
				return err
			}
		}
	}

	return nil
}

func handleUrlTarget(targetStr string, targetNode *yaml.Node, globalTarget *Url) ([]string, error) {
	// this whole function is solely for compatibility with old feed2imap
	// there it was common to specify the whole URL for each feed
	if isMaildirUrl(targetStr) {
		// old feed2imap supported maildir, we don't
		return nil, fmt.Errorf("Line %d: Maildir is not supported.", targetNode.Line)
	}

	url := Url{}
	if err := url.UnmarshalYAML(targetNode); err != nil {
		return nil, err
	}

	if globalTarget.Empty() {
		// assign first feed as global url
		*globalTarget = url.BaseUrl()
	} else if !globalTarget.CommonBaseUrl(url) {
		// if we have a url, it must be the same prefix as the global url
		return nil, fmt.Errorf("Line %d: Given URL endpoint '%s' does not match previous endpoint '%s'.",
			targetNode.Line,
			url.BaseUrl(),
			globalTarget.BaseUrl())
	}

	return url.RootPath(), nil
}
