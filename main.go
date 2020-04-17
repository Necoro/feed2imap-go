package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

var cfg = flag.String("f", "config.yml", "configuration file")

type ConfigMap map[string]interface{}

type Feed struct {
	Name      string
	Url       string
	MinFreq   int `yaml:"min-frequency"`
	ConfigMap `yaml:",inline"`
}

type Group struct {
	Group string
	Feeds []GroupFeed
}

type GroupFeed struct {
	Target string
	Feed   `yaml:",inline"`
	Group  `yaml:",inline"`
}

type Yaml struct {
	GlobalConfig ConfigMap `yaml:",inline"`
	Feeds        []GroupFeed
}

func main() {
	log.Print("Starting up...")
	flag.Parse()

	log.Printf("Reading configuration file '%s'", *cfg)
	buf, err := ioutil.ReadFile(*cfg)
	if err != nil {
		msg := fmt.Sprint("No file found: ", *cfg)
		panic(msg)
	}

	var m Yaml
	yaml.Unmarshal(buf, &m)
	fmt.Printf("--- m:\n%+v\n\n", m)

}
