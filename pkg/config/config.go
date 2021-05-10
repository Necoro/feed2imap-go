package config

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"

	"github.com/Necoro/feed2imap-go/pkg/log"
	"github.com/Necoro/feed2imap-go/pkg/util"
)

// Map is a convenience type for the non-mapped configuration options
// Mostly used for legacy options
type Map map[string]interface{}

// GlobalOptions are not feed specific
type GlobalOptions struct {
	Timeout      int      `yaml:"timeout"`
	DefaultEmail string   `yaml:"default-email"`
	Target       Url      `yaml:"target"`
	Parts        []string `yaml:"parts"`
	MaxFailures  int      `yaml:"max-failures"`
	AutoTarget   bool     `yaml:"auto-target"`
}

var DefaultGlobalOptions = GlobalOptions{
	Timeout:      30,
	MaxFailures:  10,
	DefaultEmail: username() + "@" + Hostname(),
	Target:       Url{},
	Parts:        []string{"text", "html"},
	AutoTarget:   true,
}

// Options are feed specific
// NB: Always specify a yaml name, as it is later used in processing
type Options struct {
	MinFreq     int    `yaml:"min-frequency"`
	InclImages  bool   `yaml:"include-images"`
	EmbedImages bool   `yaml:"embed-images"`
	Disable     bool   `yaml:"disable"`
	IgnHash     bool   `yaml:"ignore-hash"`
	AlwaysNew   bool   `yaml:"always-new"`
	Reupload    bool   `yaml:"reupload-if-updated"`
	NoTLS       bool   `yaml:"tls-no-verify"`
	ItemFilter  string `yaml:"item-filter"`
	Body        Body   `yaml:"body"`
}

var DefaultFeedOptions = Options{
	Body:        "default",
	MinFreq:     0,
	InclImages:  true,
	EmbedImages: false,
	IgnHash:     false,
	AlwaysNew:   false,
	Disable:     false,
	NoTLS:       false,
	ItemFilter:  "",
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

// Validate checks the configuration against common mistakes
func (cfg *Config) Validate() error {
	if cfg.Target.Empty() {
		return fmt.Errorf("No target set!")
	}

	for _, feed := range cfg.Feeds {
		if feed.Url != "" && len(feed.Exec) > 0 {
			return fmt.Errorf("Feed %s: Both 'Url' and 'Exec' set, unsure what to do.", feed.Name)
		}
	}

	return nil
}

// WithPartText marks whether 'text' part should be included in mails
func (opt GlobalOptions) WithPartText() bool {
	return util.StrContains(opt.Parts, "text")
}

// WithPartHtml marks whether 'html' part should be included in mails
func (opt GlobalOptions) WithPartHtml() bool {
	return util.StrContains(opt.Parts, "html")
}

// Load configuration from file and validate it
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

	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("Configuration invalid: %w", err)
	}

	return cfg, nil
}

// Hostname returns the current hostname, or 'localhost' if it cannot be determined
func Hostname() (hostname string) {
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
