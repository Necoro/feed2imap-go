package config

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type Map map[string]interface{}
type Feeds map[string]*Feed

func (f Feeds) String() string {
	var b strings.Builder
	app := func(a ...interface{}) {
		_, _ = fmt.Fprint(&b, a...)
	}
	app("Feeds [")

	first := true
	for k, v := range f {
		if !first {
			app(", ")
		}
		app(`"`, k, `"`, ": ")
		if v == nil {
			app("<nil>")
		} else {
			_, _ = fmt.Fprintf(&b, "%+v", *v)
		}
		first = false
	}
	app("]")

	return b.String()
}

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
