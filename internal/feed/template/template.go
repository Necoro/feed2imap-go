package template

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/Necoro/feed2imap-go/pkg/log"
)

func dict(v ...interface{}) map[string]interface{} {
	dict := make(map[string]interface{})
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

func join(sep string, parts []string) string {
	return strings.Join(parts, sep)
}

func lastUrlPart(url string) string {
	split := strings.Split(url, "/")
	return split[len(split)-1]
}

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

func html(s string) template.HTML {
	return template.HTML(s)
}

var funcMap = template.FuncMap{
	"dict":        dict,
	"join":        join,
	"lastUrlPart": lastUrlPart,
	"byteCount":   byteCount,
	"html":        html,
}

func fromString(name, templateStr string) *template.Template {
	tpl := template.New(name).Funcs(funcMap)
	return template.Must(tpl.Parse(templateStr))
}
