package template

import (
	"fmt"
	"html/template"
	"strings"
)

func dict(v ...string) map[string]string {
	dict := map[string]string{}
	lenv := len(v)
	for i := 0; i < lenv; i += 2 {
		key := v[i]
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

func LastUrlPart(url string) string {
	split := strings.Split(url, "/")
	return split[len(split)-1]
}

func byteCount(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

var funcMap = template.FuncMap{
	"dict":        dict,
	"join":        join,
	"lastUrlPart": LastUrlPart,
	"byteCount":   byteCount,
}

func fromString(name, templateStr string) *template.Template {
	tpl := template.New(name).Funcs(funcMap)
	return template.Must(tpl.Parse(templateStr))
}
