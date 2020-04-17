package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var cfg = flag.String("f", "config.yml", "configuration file")

type ConfigMap map[string]interface{}

type configFeed struct {
	Name      string
	Url       string
	MinFreq   int `yaml:"min-frequency"`
	ConfigMap `yaml:",inline"`
}

type configGroup struct {
	Group string
	Feeds []configGroupFeed
}

type configGroupFeed struct {
	Target      *string
	configFeed  `yaml:",inline"`
	configGroup `yaml:",inline"`
}

func (grpFeed *configGroupFeed) isGroup() bool {
	return grpFeed.configGroup.Group != ""
}

func (grpFeed *configGroupFeed) isFeed() bool {
	return grpFeed.configFeed.Name != ""
}

func (grpFeed *configGroupFeed) target() string {
	if grpFeed.Target != nil {
		return *grpFeed.Target
	}
	if grpFeed.Name != "" {
		return grpFeed.Name
	}

	return grpFeed.Group
}

type config struct {
	GlobalConfig ConfigMap `yaml:",inline"`
	Feeds        []configGroupFeed
}

type feed struct {
	name    string
	target  string
	url     string
	minFreq int
	config  ConfigMap
}

var feeds = make(map[string]*feed)

func appTarget(target, app string) string {
	if target == "" {
		return app
	}

	if app == "" {
		return target
	}

	return target + "/" + app
}

func buildFeeds(cfg []configGroupFeed, target string) error {
	for _, f := range cfg {
		target := appTarget(target, f.target())
		switch {
		case f.isFeed() && f.isGroup():
			return fmt.Errorf("Entry with Target %s is both a feed and a group", target)
		case f.isFeed():
			name := f.Name
			if _, ok := feeds[name]; !ok {
				return fmt.Errorf("Duplicate feed name '%s'", name)
			}
			feeds[name] = &feed{
				name:    name,
				target:  target,
				url:     f.Url,
				minFreq: f.MinFreq,
				config:  f.ConfigMap,
			}
		case f.isGroup():
			if err := buildFeeds(f.Feeds, target); err != nil {
				return err
			}
		}
	}

	return nil
}

func run() error {
	log.Print("Starting up...")
	flag.Parse()

	log.Printf("Reading configuration file '%s'", *cfg)
	buf, err := ioutil.ReadFile(*cfg)
	if err != nil {
		return fmt.Errorf("while reading '%s': %w", *cfg, err)
	}

	var parsedCfg config
	if err := yaml.Unmarshal(buf, &parsedCfg); err != nil {
		return fmt.Errorf("while unmarshalling: %w", err)
	}
	fmt.Printf("--- parsedCfg:\n%+v\n\n", parsedCfg)

	if err := buildFeeds(parsedCfg.Feeds, ""); err != nil {
		return fmt.Errorf("while parsing: %w", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.SetOutput(os.Stderr)
		log.Print("Error: ", err)
		os.Exit(1)
	}
}
