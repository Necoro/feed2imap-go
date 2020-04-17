package config

import (
	"fmt"
	"io/ioutil"
)

type Map map[string]interface{}
type Feeds map[string]*Feed

type Config struct {
	GlobalConfig Map
	Feeds        Feeds
}

type Feed struct {
	Name       string
	Target     string `yaml:"-"`
	Url        string
	MinFreq    int   `yaml:"min-frequency"`
	InclImages *bool `yaml:"include-images"`
}

func Load(path string) (Config, error) {
	var finishedCfg Config

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return finishedCfg, fmt.Errorf("while reading '%s': %w", path, err)
	}

	var parsedCfg config
	if parsedCfg, err = parse(buf); err != nil {
		return finishedCfg, err
	}

	finishedCfg = Config{
		GlobalConfig: parsedCfg.GlobalConfig,
		Feeds:        make(Feeds),
	}

	if err := buildFeeds(parsedCfg.Feeds, "", finishedCfg.Feeds); err != nil {
		return finishedCfg, fmt.Errorf("while parsing: %w", err)
	}

	return finishedCfg, nil
}
