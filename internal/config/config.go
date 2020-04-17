package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Map map[string]interface{}
type Feeds map[string]*Feed

type config struct {
	GlobalConfig Map `yaml:",inline"`
	Feeds        []configGroupFeed
}

type Config struct {
	GlobalConfig Map `yaml:",inline"`
	Feeds        Feeds
}

type Feed struct {
	Name    string
	Target  string
	Url     string
	MinFreq int
	Config  Map
}

func Load(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("while reading '%s': %w", path, err)
	}

	var parsedCfg config
	if err := yaml.Unmarshal(buf, &parsedCfg); err != nil {
		return nil, fmt.Errorf("while unmarshalling: %w", err)
	}
	fmt.Printf("--- parsedCfg:\n%+v\n\n", parsedCfg)

	var finishedCfg = Config{
		GlobalConfig: parsedCfg.GlobalConfig,
		Feeds:        make(Feeds),
	}

	if err := buildFeeds(parsedCfg.Feeds, "", finishedCfg.Feeds); err != nil {
		return nil, fmt.Errorf("while parsing: %w", err)
	}

	return &finishedCfg, nil
}
