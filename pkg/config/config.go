package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/Necoro/feed2imap-go/pkg/log"
	"github.com/Necoro/feed2imap-go/pkg/util"
)

// Convenience type for the non-mapped configuration options
// Mostly used for legacy options
type Map map[string]interface{}

// Global options, not feed specific
type GlobalOptions struct {
	Timeout      int      `yaml:"timeout"`
	DefaultEmail string   `yaml:"default-email"`
	Target       string   `yaml:"target"`
	Parts        []string `yaml:"parts"`
}

// Default global options
var DefaultGlobalOptions = GlobalOptions{
	Timeout:      30,
	DefaultEmail: username() + "@" + hostname(),
	Target:       "",
	Parts:        []string{"text", "html"},
}

// Per feed options
type Options struct {
	MinFreq    *int  `yaml:"min-frequency"`
	InclImages *bool `yaml:"include-images"`
	Disable    *bool `yaml:"disable"`
	IgnHash    *bool `yaml:"ignore-hash"`
}

func (opt *Options) mergeFrom(other Options) {
	if opt.MinFreq == nil {
		opt.MinFreq = other.MinFreq
	}
	if opt.InclImages == nil {
		opt.InclImages = other.InclImages
	}
	if opt.IgnHash == nil {
		opt.IgnHash = other.IgnHash
	}
	if opt.Disable == nil {
		opt.Disable = other.Disable
	}
}

// Default feed options
var DefaultFeedOptions Options

func init() {
	one := 1
	fal := false
	DefaultFeedOptions = Options{
		MinFreq:    &one,
		InclImages: &fal,
		IgnHash:    &fal,
		Disable:    &fal,
	}
}

// Config holds the global configuration options and the configured feeds
type Config struct {
	GlobalOptions `yaml:",inline"`
	GlobalConfig  Map     `yaml:",inline"`
	FeedOptions   Options `yaml:"options"`
	Feeds         Feeds   `yaml:"-"`
}

// WithDefault returns a configuration initialized with default values.
func WithDefault() *Config {
	return &Config{
		GlobalOptions: DefaultGlobalOptions,
		FeedOptions:   DefaultFeedOptions,
		GlobalConfig:  Map{},
		Feeds:         Feeds{},
	}
}

// Validates the configuration against common mistakes
func (cfg *Config) Validate() error {
	if cfg.Target == "" {
		return fmt.Errorf("No target set!")
	}

	return nil
}

// Marks whether 'text' part should be included in mails
func (cfg *Config) WithPartText() bool {
	return util.StrContains(cfg.Parts, "text")
}

// Marks whether 'html' part should be included in mails
func (cfg *Config) WithPartHtml() bool {
	return util.StrContains(cfg.Parts, "html")
}

// Current feed2imap version
func Version() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "(unknown)"
	}
	return bi.Main.Version
}

// Load configuration from file
func Load(path string) (*Config, error) {
	log.Printf("Reading configuration file '%s'", path)

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("while reading '%s': %w", path, err)
	}

	cfg := WithDefault()
	if err = cfg.parse(buf); err != nil {
		return nil, fmt.Errorf("while parsing: %w", err)
	}

	cfg.pushFeedOptions()

	return cfg, nil
}

func (cfg *Config) pushFeedOptions() {
	for _, feed := range cfg.Feeds {
		feed.Options.mergeFrom(cfg.FeedOptions)
	}
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
