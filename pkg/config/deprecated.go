package config

import (
	"fmt"

	"github.com/Necoro/feed2imap-go/pkg/log"
)

type deprecated struct {
	msg    string
	handle func(any, *GlobalOptions, *Options)
}

var unsupported = deprecated{
	"It won't be supported and is ignored!",
	nil,
}

var deprecatedOpts = map[string]deprecated{
	"dumpdir":       unsupported,
	"debug-updated": {"Use '-d' as option instead.", nil},
	"execurl":       {"Use 'exec' instead.", nil},
	"filter":        {"Use 'item-filter' instead.", nil},
	"disable-ssl-verification": {"Interpreted as 'tls-no-verify'.", func(i any, global *GlobalOptions, opts *Options) {
		if val, ok := i.(bool); ok {
			if val && !opts.NoTLS {
				// do not overwrite the set NoTLS flag!
				opts.NoTLS = val
			}
		} else {
			log.Errorf("disable-ssl-verification: value '%v' cannot be interpreted as a boolean. Ignoring!", i)
		}
	}},
}

func handleDeprecated(option string, value any, feed string, global *GlobalOptions, opts *Options) bool {
	dep, ok := deprecatedOpts[option]
	if !ok {
		return false
	}

	var prefix string
	if feed != "" {
		prefix = fmt.Sprintf("Feed '%s': ", feed)
	} else {
		prefix = "Global "
	}

	log.Warnf("%sOption '%s' is deprecated: %s", prefix, option, dep.msg)

	if dep.handle != nil {
		dep.handle(value, global, opts)
	}

	return true
}
