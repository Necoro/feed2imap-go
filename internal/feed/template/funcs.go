package template

import (
	"fmt"
	html "html/template"
	"strconv"
	"strings"
	text "text/template"

	"github.com/Necoro/feed2imap-go/pkg/log"
)

// dict creates a map out of the passed in key/value pairs.
func dict(v ...any) map[string]any {
	dict := make(map[string]any)
	lenv := len(v)
	for i := 0; i < lenv; i += 2 {
		key := v[i].(string)
		if i+1 >= lenv {
			dict[key] = ""
			continue
		}
		dict[key] = v[i+1]
	}
	return dict
}

// join takes a separator and a list of strings and puts the former in between each pair of the latter.
func join(sep string, parts []string) string {
	return strings.Join(parts, sep)
}

// lastUrlPart returns the last part of a URL string
func lastUrlPart(url string) string {
	split := strings.Split(url, "/")
	return split[len(split)-1]
}

// byteCount receives an integer as a string, that is interpreted as a size in bytes.
// This size is then equipped with the corresponding unit:
// 1023 --> 1023 B; 1024 --> 1.0 KB; ...
func byteCount(str string) string {
	var b uint64
	if str != "" {
		var err error
		if b, err = strconv.ParseUint(str, 10, 64); err != nil {
			log.Printf("Cannot convert '%s' to byte count: %s", str, err)
		}
	}

	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func _html(s string) html.HTML {
	return html.HTML(s)
}

var funcMap = text.FuncMap{
	"dict":        dict,
	"join":        join,
	"lastUrlPart": lastUrlPart,
	"byteCount":   byteCount,
	"html":        _html,
}
