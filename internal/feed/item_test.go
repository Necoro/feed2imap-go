package feed

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/internal/feed/template"
)

func TestCreator(t *testing.T) {
	type p = gofeed.Person

	tests := map[string]struct {
		inp []*gofeed.Person
		out string
	}{
		"Empty":      {nil, ""},
		"Nil":        {[]*p{nil}, ""},
		"Name":       {[]*p{{Name: "Foo"}}, "Foo"},
		"Email only": {[]*p{{Email: "foo@example.com"}}, ""},
		"URL only":   {[]*p{{URL: "https://foo.example"}}, "https://foo.example"},
		"Multiple":   {[]*p{{Name: "Foo"}, {Email: "x@example.com"}, {Name: "Bar", URL: "https://bar.example"}}, "Foo, Bar"},
	}

	for name, tt := range tests {
		t.Run(name, func(tst *testing.T) {
			item := Item{Item: &gofeed.Item{Authors: tt.inp}}

			if diff := cmp.Diff(tt.out, item.Creator()); diff != "" {
				tst.Error(diff)
			}
		})
	}
}

func TestHtmlAuthorLinks(t *testing.T) {
	tests := map[string]struct {
		inp      []*gofeed.Person
		contains []string
		excludes []string
	}{
		"Name": {
			inp:      []*gofeed.Person{{Name: "Foo"}},
			contains: []string{"Author:", "Foo"},
			excludes: []string{"Foo</a>"},
		},
		"Name + URL": {
			inp:      []*gofeed.Person{{Name: "Foo", URL: "https://foo.example"}, {Name: "Bar"}},
			contains: []string{`<a style="color: #ababab;" href="https://foo.example">Foo</a>, Bar`},
		},
		"URL only": {
			inp:      []*gofeed.Person{{URL: "https://foo.example"}},
			contains: []string{`<a style="color: #ababab;" href="https://foo.example">https://foo.example</a>`},
		},
		"Escaping": {
			inp:      []*gofeed.Person{{Name: "<b>Foo</b>", URL: "javascript:alert(1)"}},
			contains: []string{`href="#ZgotmplZ">&lt;b&gt;Foo&lt;/b&gt;</a>`},
			excludes: []string{"javascript:", "<b>"},
		},
		"None": {
			inp:      []*gofeed.Person{{Email: "foo@example.com"}},
			excludes: []string{"Author:"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(tst *testing.T) {
			item := &Item{
				Item: &gofeed.Item{Authors: tt.inp},
				Feed: &gofeed.Feed{FeedLink: "https://example.com/feed"},
			}

			var out strings.Builder
			if err := template.Html.Execute(&out, item); err != nil {
				tst.Fatal(err)
			}

			res := out.String()
			for _, s := range tt.contains {
				if !strings.Contains(res, s) {
					tst.Errorf("output does not contain %q:\n%s", s, res)
				}
			}
			for _, s := range tt.excludes {
				if strings.Contains(res, s) {
					tst.Errorf("output contains %q:\n%s", s, res)
				}
			}
			if !strings.Contains(res, `<a style="color: #ababab;" href="https://example.com/feed">https://example.com/feed</a>`) {
				tst.Errorf("output does not contain feed link:\n%s", res)
			}
		})
	}
}
