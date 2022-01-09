package template

import (
	_ "embed"
	"errors"
	"fmt"
	html "html/template"
	"io"
	"io/fs"
	"os"
	text "text/template"

	"github.com/Necoro/feed2imap-go/pkg/log"
)

type template interface {
	Execute(wr io.Writer, data interface{}) error
	Name() string
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
	useHtml:  true,
	dflt:     defaultHtmlTpl,
	template: html.New("Html").Funcs(funcMap),
}

var Text = Template{
	useHtml:  false,
	dflt:     defaultTextTpl,
	template: text.New("Text").Funcs(funcMap),
}

func (tpl *Template) loadDefault() {
	if err := tpl.load(tpl.dflt); err != nil {
		panic(err)
	}
}

func (tpl *Template) load(content string) (err error) {
	if tpl.useHtml {
		_, err = tpl.template.(*html.Template).Parse(content)
	} else {
		_, err = tpl.template.(*text.Template).Parse(content)
	}
	return
}

func (tpl *Template) LoadFile(file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.Errorf("Template file '%s' does not exist, keeping default.", file)
			return nil
		} else {
			return fmt.Errorf("reading template file '%s': %w", file, err)
		}
	}

	return tpl.load(string(content))
}

func init() {
	Html.loadDefault()
	Text.loadDefault()
}
