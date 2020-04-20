package config

import (
	"os"
	"os/user"
	"runtime"
	"strings"
)

type Map map[string]interface{}

type GlobalOptions struct {
	Timeout      int      `yaml:"timeout"`
	DefaultEmail string   `yaml:"default-email"`
	Target       string   `yaml:"target"`
	Parts        []string `yaml:"parts"`
}

var DefaultGlobalOptions = GlobalOptions{
	Timeout:      30,
	DefaultEmail: username() + "@" + hostname(),
	Target:       "",
	Parts:        []string{"text", "html"},
}

type Config struct {
	GlobalOptions
	GlobalConfig Map
}

type Options struct {
	MinFreq    int   `yaml:"min-frequency"`
	InclImages *bool `yaml:"include-images"`
}

func (c *Config) WithPartText() bool {
	for _, part := range c.Parts {
		if part == "text" {
			return true
		}
	}

	return false
}

func (c *Config) WithPartHtml() bool {
	for _, part := range c.Parts {
		if part == "html" {
			return true
		}
	}

	return false
}

func hostname() (hostname string) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	return
}

func username() string {
	u, err := user.Current()
	switch {
	case err != nil:
		return "user"
	case runtime.GOOS == "windows":
		// the domain is attached -- remove it again
		split := strings.Split(u.Username, "\\")
		return split[len(split)-1]
	default:
		return u.Username
	}
}
