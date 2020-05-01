package config

import (
	"fmt"
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
// NB: Always specify a yaml name, as it is later used in processing
type Options struct {
	MinFreq    int  `yaml:"min-frequency"`
	InclImages bool `yaml:"include-images"`
	Disable    bool `yaml:"disable"`
	IgnHash    bool `yaml:"ignore-hash"`
	AlwaysNew  bool `yaml:"always-new"`
	NoTLS      bool `yaml:"tls-no-verify"`
}

// Default feed options
var DefaultFeedOptions = Options{
	MinFreq:    1,
	InclImages: false,
	IgnHash:    false,
	AlwaysNew:  false,
	Disable:    false,
	NoTLS:      false,
}

// Config holds the global configuration options and the configured feeds
type Config struct {
	GlobalOptions `yaml:",inline"`
	FeedOptions   Options `yaml:"options"`
	Feeds         Feeds   `yaml:"-"`
}

// WithDefault returns a configuration initialized with default values.
func WithDefault() *Config {
	return &Config{
		GlobalOptions: DefaultGlobalOptions,
		FeedOptions:   DefaultFeedOptions,
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

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("while opening '%s': %w", path, err)
	}

	cfg := WithDefault()
	if err = cfg.parse(f); err != nil {
		return nil, fmt.Errorf("while parsing: %w", err)
	}

	return cfg, nil
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
