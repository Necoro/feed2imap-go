package config

type Map map[string]interface{}

type Config struct {
	GlobalConfig Map
}

type Options struct {
	MinFreq    int   `yaml:"min-frequency"`
	InclImages *bool `yaml:"include-images"`
}
