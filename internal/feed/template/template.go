package template

import (
	_ "embed"
	html "html/template"
	"io"
	text "text/template"
)

type template interface {
	Execute(wr io.Writer, data interface{}) error
}

type Template struct {
	template
	useHtml bool
	dflt    string
}

//go:embed html.tpl
var defaultHtmlTpl string

//go:embed text.tpl
var defaultTextTpl string

var Html = Template{
	useHtml: true,
	dflt:    defaultHtmlTpl,
}

var Text = Template{
	useHtml: false,
	dflt:    defaultTextTpl,
}

func (tpl *Template) loadDefault() {
	if tpl.useHtml {
		tpl.template = html.Must(html.New("Html").Funcs(funcMap).Parse(tpl.dflt))
	} else {
		tpl.template = text.Must(text.New("Text").Funcs(funcMap).Parse(tpl.dflt))
	}
}

func init() {
	Html.loadDefault()
	Text.loadDefault()
}
