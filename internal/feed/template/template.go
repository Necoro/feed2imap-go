package template

import (
	"fmt"
	html "html/template"
	"io"
	"strconv"
	"strings"
	text "text/template"

	"github.com/Necoro/feed2imap-go/pkg/log"
)

type Template interface {
	Execute(wr io.Writer, data interface{}) error
}

//go:embed html.tpl
var htmlTpl string

//go:embed text.tpl
var textTpl string

var Html = fromString("Feed", htmlTpl, true)
var Text = fromString("Feed", textTpl, false)

func must(t Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return t
}

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

func _html(s string) html.HTML {
	return html.HTML(s)
}

var funcMap = html.FuncMap{
	"dict":        dict,
	"join":        join,
	"lastUrlPart": lastUrlPart,
	"byteCount":   byteCount,
	"html":        _html,
}

func fromString(name, templateStr string, useHtml bool) Template {
	if useHtml {
		return must(html.New(name).Funcs(funcMap).Parse(templateStr))
	} else {
		return must(text.New(name).Funcs(text.FuncMap(funcMap)).Parse(templateStr))
	}
}
