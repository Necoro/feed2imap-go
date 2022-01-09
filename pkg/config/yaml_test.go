package config

import (
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"

	"github.com/Necoro/feed2imap-go/internal/http"
)

func t(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ".")
}
func n(s string) (n yaml.Node) {
	n.SetString(s)
	return
}

var null = yaml.Node{Tag: nullTag}

func TestBuildOptions(tst *testing.T) {
	tests := []struct {
		name     string
		inp      Map
		opts     Options
		out      Options
		unknowns []string
	}{
		{"Empty", nil, Options{}, Options{}, []string{}},
		{"Simple copy", nil, Options{MinFreq: 75}, Options{MinFreq: 75}, []string{}},
		{"Unknowns", Map{"foo": 1}, Options{}, Options{}, []string{"foo"}},
		{"Override", Map{"include-images": true}, Options{InclImages: false}, Options{InclImages: true}, []string{}},
		{"Non-Standard Type", Map{"body": "both"}, Options{}, Options{Body: "both"}, []string{}},
		{"Mixed", Map{"min-frequency": 24}, Options{MinFreq: 6, InclImages: true}, Options{MinFreq: 24, InclImages: true}, []string{}},
		{"All",
			Map{"max-frequency": 12, "include-images": true, "ignore-hash": true, "obsolete": 54},
			Options{MinFreq: 6, InclImages: true, IgnHash: false},
			Options{MinFreq: 6, InclImages: true, IgnHash: true},
			[]string{"max-frequency", "obsolete"},
		},
	}

	for _, tt := range tests {
		tst.Run(tt.name, func(tst *testing.T) {
			out, unk := buildOptions(&tt.opts, tt.inp)

			if diff := cmp.Diff(tt.out, out); diff != "" {
				tst.Error(diff)
			}

			sort.Strings(unk)
			sort.Strings(tt.unknowns)

			if diff := cmp.Diff(unk, tt.unknowns); diff != "" {
				tst.Error(diff)
			}
		})
	}
}

func TestBuildFeeds(tst *testing.T) {
	tests := []struct {
		name         string
		wantErr      bool
		target       string
		feeds        []configGroupFeed
		result       Feeds
		noAutoTarget bool
	}{
		{name: "Empty input", wantErr: false, target: "", feeds: nil, result: Feeds{}},
		{name: "Empty Feed", wantErr: true, target: "",
			feeds: []configGroupFeed{
				{Target: n("foo"), Feed: feed{Url: "google.de"}},
			}, result: Feeds{}},
		{name: "Empty Feed", wantErr: true, target: "",
			feeds: []configGroupFeed{
				{Feed: feed{Url: "google.de"}},
			}, result: Feeds{}},
		{name: "Duplicate Feed Name", wantErr: true, target: "",
			feeds: []configGroupFeed{
				{Feed: feed{Name: "Dup"}},
				{Feed: feed{Name: "Dup"}},
			}, result: Feeds{}},
		{name: "Simple", wantErr: false, target: "",
			feeds: []configGroupFeed{
				{Target: n("foo"), Feed: feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("foo")}},
		},
		{name: "Simple With Target", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: n("foo"), Feed: feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("moep.foo")}},
		},
		{name: "Simple With Target and Whitespace", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: n("\r\nfoo "), Feed: feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("moep.foo")}},
		},
		{name: "Simple With Target and NoAutoTarget", wantErr: false, target: "moep", noAutoTarget: true,
			feeds: []configGroupFeed{
				{Target: n("foo"), Feed: feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("moep.foo")}},
		},
		{name: "Simple Without Target", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Feed: feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("moep.muh")}},
		},
		{name: "Simple Without Target and NoAutoTarget", wantErr: false, target: "moep", noAutoTarget: true,
			feeds: []configGroupFeed{
				{Feed: feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("moep")}},
		},
		{name: "Simple With Nil Target", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: null, Feed: feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("moep")}},
		},
		{name: "Simple With Empty Target", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: n(""), Feed: feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("moep")}},
		},
		{name: "Simple With Blank Target", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: n(" "), Feed: feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("moep")}},
		},
		{name: "Multiple Feeds", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: n("foo"), Feed: feed{Name: "muh"}},
				{Feed: feed{Name: "bar"}},
			},
			result: Feeds{
				"muh": &Feed{Name: "muh", Target: t("moep.foo")},
				"bar": &Feed{Name: "bar", Target: t("moep.bar")},
			},
		},
		{name: "URL Target", wantErr: false, target: "",
			feeds: []configGroupFeed{
				{Target: n("imap://foo.bar:443/INBOX/Feed"), Feed: feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("INBOX.Feed")}},
		},
		{name: "Multiple URL Targets", wantErr: false, target: "",
			feeds: []configGroupFeed{
				{Target: n("imap://foo.bar:443/INBOX/Feed"), Feed: feed{Name: "muh"}},
				{Target: n("imap://foo.bar:443/INBOX/Feed2"), Feed: feed{Name: "bar"}},
			},
			result: Feeds{
				"muh": &Feed{Name: "muh", Target: t("INBOX.Feed")},
				"bar": &Feed{Name: "bar", Target: t("INBOX.Feed2")},
			},
		},
		{name: "Mixed URL Targets", wantErr: true, target: "",
			feeds: []configGroupFeed{
				{Target: n("imap://foo.bar:443/INBOX/Feed"), Feed: feed{Name: "muh"}},
				{Target: n("imap://other.bar:443/INBOX/Feed"), Feed: feed{Name: "bar"}},
			},
			result: Feeds{},
		},
		{name: "Maildir URL Target", wantErr: true, target: "",
			feeds: []configGroupFeed{
				{Target: n("maildir:///home/foo/INBOX/Feed"), Feed: feed{Name: "muh"}},
			},
			result: Feeds{},
		},
		{name: "Empty Group", wantErr: false, target: "",
			feeds: []configGroupFeed{
				{Group: group{Group: "G1"}},
			},
			result: Feeds{},
		},
		{name: "Simple Group", wantErr: false, target: "",
			feeds: []configGroupFeed{
				{Group: group{Group: "G1", Feeds: []configGroupFeed{
					{Target: n("bar"), Feed: feed{Name: "F1"}},
					{Target: n(""), Feed: feed{Name: "F2"}},
					{Feed: feed{Name: "F3"}},
				}}},
			},
			result: Feeds{
				"F1": &Feed{Name: "F1", Target: t("G1.bar")},
				"F2": &Feed{Name: "F2", Target: t("G1")},
				"F3": &Feed{Name: "F3", Target: t("G1.F3")},
			},
		},
		{name: "Simple Group, NoAutoTarget", wantErr: false, target: "IN", noAutoTarget: true,
			feeds: []configGroupFeed{
				{Group: group{Group: "G1", Feeds: []configGroupFeed{
					{Target: n("bar"), Feed: feed{Name: "F1"}},
					{Target: n(""), Feed: feed{Name: "F2"}},
					{Feed: feed{Name: "F3"}},
				}}},
			},
			result: Feeds{
				"F1": &Feed{Name: "F1", Target: t("IN.bar")},
				"F2": &Feed{Name: "F2", Target: t("IN")},
				"F3": &Feed{Name: "F3", Target: t("IN")},
			},
		},
		{name: "Nested Groups", wantErr: false, target: "",
			feeds: []configGroupFeed{
				{Group: group{Group: "G1", Feeds: []configGroupFeed{
					{Feed: feed{Name: "F0"}},
					{Target: n("bar"), Group: group{Group: "G2",
						Feeds: []configGroupFeed{{Feed: feed{Name: "F1"}}}}},
					{Target: n(""), Group: group{Group: "G3",
						Feeds: []configGroupFeed{{Target: n("baz"), Feed: feed{Name: "F2"}}}}},
					{Group: group{Group: "G4",
						Feeds: []configGroupFeed{{Feed: feed{Name: "F3"}}}}},
				}}},
			},
			result: Feeds{
				"F0": &Feed{Name: "F0", Target: t("G1.F0")},
				"F1": &Feed{Name: "F1", Target: t("G1.bar.F1")},
				"F2": &Feed{Name: "F2", Target: t("G1.baz")},
				"F3": &Feed{Name: "F3", Target: t("G1.G4.F3")},
			},
		},
	}
	for _, tt := range tests {
		tst.Run(tt.name, func(tst *testing.T) {
			var feeds = Feeds{}
			var opts = Options{}
			var globalTarget = Url{}
			err := buildFeeds(tt.feeds, t(tt.target), feeds, &opts, !tt.noAutoTarget, &globalTarget)
			if (err != nil) != tt.wantErr {
				tst.Errorf("buildFeeds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.result, feeds); !tt.wantErr && diff != "" {
				tst.Error(diff)
			}
		})
	}
}

func defaultConfig(feeds []configGroupFeed, global Map) config {
	defCfg := WithDefault()
	return config{
		Config:       defCfg,
		Feeds:        feeds,
		GlobalConfig: global,
	}
}

func TestUnmarshal(tst *testing.T) {
	tests := []struct {
		name    string
		inp     string
		wantErr bool
		config  config
	}{
		{name: "Empty",
			inp: "", wantErr: false, config: defaultConfig(nil, nil)},
		{name: "Trash", inp: "Something", wantErr: true},
		{name: "Simple config",
			inp: "something: 1\nsomething_else: 2", wantErr: false, config: defaultConfig(nil, Map{"something": 1, "something_else": 2})},
		{name: "Known config",
			inp: "whatever: 2\ndefault-email: foo@foobar.de\ntimeout: 60\nsomething: 1", wantErr: false, config: func() config {
				c := defaultConfig(nil, Map{"something": 1, "whatever": 2})
				c.Timeout = 60
				c.DefaultEmail = "foo@foobar.de"
				return c
			}()},
		{name: "Known config with feed-options",
			inp: "whatever: 2\ntimeout: 60\noptions:\n  min-frequency: 6", wantErr: false, config: func() config {
				c := defaultConfig(nil, Map{"whatever": 2})
				c.Timeout = 60
				c.FeedOptions.MinFreq = 6
				return c
			}()},
		{name: "Known config with invalid feed-options",
			inp: "options:\n  max-frequency: 6", wantErr: true, config: config{}},
		{name: "Nested config",
			inp: `
options:
  cookies:
    - name: foo
      value: bar
`, wantErr: false, config: func() config {
				c := defaultConfig(nil, nil)
				c.FeedOptions.Cookies = []http.Cookie{{Name: "foo", Value: "bar"}}
				return c
			}()},
		{name: "Nested config; multiple",
			inp: `
options:
  cookies:
    - name: foo
      value: bar
    - name: baz
      value: uff
`, wantErr: false, config: func() config {
				c := defaultConfig(nil, nil)
				c.FeedOptions.Cookies = []http.Cookie{{"foo", "bar"}, {"baz", "uff"}}
				return c
			}()},
		{name: "Config with feed",
			inp: `
something: 1
feeds:
  - name: Foo
    url: whatever
    target: bar
    include-images: true
    unknown-option: foo
`,
			wantErr: false,
			config: defaultConfig([]configGroupFeed{{
				Target: n("bar"),
				Feed: feed{
					Name: "Foo",
					Url:  "whatever",
				},
				Options: Map{"include-images": true, "unknown-option": "foo"},
			}}, Map{"something": 1})},

		{name: "Feed with Exec",
			inp: `
feeds:
  - name: Foo
    exec: [whatever, -i, http://foo.bar]
    target: bar
    include-images: true
    unknown-option: foo
`,
			wantErr: false,
			config: defaultConfig([]configGroupFeed{{
				Target: n("bar"),
				Feed: feed{
					Name: "Foo",
					Exec: []string{"whatever", "-i", "http://foo.bar"},
				},
				Options: Map{"include-images": true, "unknown-option": "foo"},
			}}, nil)},

		{name: "Feeds",
			inp: `
feeds:
  - name: Foo
    url: whatever
    min-frequency: 2
  - name: Shrubbery
    url: google.de
    target: bla
    include-images: false
`,
			wantErr: false,
			config: defaultConfig([]configGroupFeed{
				{
					Feed: feed{
						Name: "Foo",
						Url:  "whatever",
					},
					Options: Map{"min-frequency": 2},
				},
				{
					Target: n("bla"),
					Feed: feed{
						Name: "Shrubbery",
						Url:  "google.de",
					},
					Options: Map{"include-images": false},
				},
			}, nil),
		},
		{name: "Empty Group",
			inp: `
feeds:
  - group: Foo
    target: bla
`,
			wantErr: false,
			config:  defaultConfig([]configGroupFeed{{Target: n("bla"), Group: group{"Foo", nil}}}, nil),
		},
		{name: "Feeds and Groups",
			inp: `
feeds:
  - name: Foo
    url: whatever
  - group: G1
    target: target
    feeds:
     - group: G2
       target: ""
       feeds:
        - name: F1
          url: google.de
     - name: F2
     - name: F3
       target:
     - group: G3
`,
			wantErr: false,
			config: defaultConfig([]configGroupFeed{
				{Feed: feed{
					Name: "Foo",
					Url:  "whatever",
				}},
				{Target: n("target"), Group: group{
					Group: "G1",
					Feeds: []configGroupFeed{
						{Target: n(""), Group: group{
							Group: "G2",
							Feeds: []configGroupFeed{
								{Feed: feed{Name: "F1", Url: "google.de"}},
							}},
						},
						{Feed: feed{Name: "F2"}},
						{Feed: feed{Name: "F3"}, Target: null},
						{Group: group{Group: "G3"}},
					}},
				},
			}, nil),
		},
	}

	eqNode := cmp.Comparer(func(l, r yaml.Node) bool {
		return l.Tag == r.Tag && l.Value == r.Value
	})

	for _, tt := range tests {
		tst.Run(tt.name, func(tst *testing.T) {
			in := strings.NewReader(tt.inp)

			got, err := unmarshal(in, WithDefault())
			if (err != nil) != tt.wantErr {
				tst.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if diff := cmp.Diff(tt.config, got, eqNode); diff != "" {
					tst.Error(diff)
				}
			}
		})
	}
}

func TestCompleteFeed(tst *testing.T) {
	inp := `
feeds:
  - name: Foo
    url: whatever
  - group: G1
    target: target
    feeds:
     - group: G2
       target: ""
       feeds:
        - name: F1
          url: google.de
     - name: F2
     - name: F3
       target:
     - name: F4
       target: "G4"
     - name: F5
       target: ~
     - name: F6
       target: ""
     - group: G3
     - group: G4
       feeds:
        - name: F7
`
	res := Feeds{
		"Foo": &Feed{Name: "Foo", Target: t("Foo"), Url: "whatever"},
		"F1":  &Feed{Name: "F1", Target: t("target.F1"), Url: "google.de"},
		"F2":  &Feed{Name: "F2", Target: t("target.F2")},
		"F3":  &Feed{Name: "F3", Target: t("target")},
		"F4":  &Feed{Name: "F4", Target: t("target.G4")},
		"F5":  &Feed{Name: "F5", Target: t("target")},
		"F6":  &Feed{Name: "F6", Target: t("target")},
		"F7":  &Feed{Name: "F7", Target: t("target.G4.F7")},
	}

	c := WithDefault()
	c.FeedOptions = Options{}

	if err := c.parse(strings.NewReader(inp)); err != nil {
		tst.Error(err)
	} else {
		if diff := cmp.Diff(res, c.Feeds); diff != "" {
			tst.Error(diff)
		}
	}
}

func TestURLFeedWithoutGlobalTarget(tst *testing.T) {
	inp := `
feeds:
  - name: Foo
    target: imap://foo.bar:443/INBOX/Feed
`
	res := Feeds{
		"Foo": &Feed{Name: "Foo", Target: t("INBOX.Feed")},
	}

	c := WithDefault()
	c.FeedOptions = Options{}

	if err := c.parse(strings.NewReader(inp)); err != nil {
		tst.Error(err)
	} else {
		if diff := cmp.Diff(res, c.Feeds); diff != "" {
			tst.Error(diff)
		}
		if diff := cmp.Diff("imap://foo.bar:443", c.Target.String()); diff != "" {
			tst.Error(diff)
		}
	}
}

func TestURLFeedWithGlobalTarget(tst *testing.T) {
	inp := `
target: imaps://foo.bar/INBOX/Feeds
feeds:
  - name: Foo
    target: imaps://foo.bar:993/Some/Other/Path
`
	res := Feeds{
		"Foo": &Feed{Name: "Foo", Target: t("Some.Other.Path")},
	}

	c := WithDefault()
	c.FeedOptions = Options{}

	if err := c.parse(strings.NewReader(inp)); err != nil {
		tst.Error(err)
	} else {
		if diff := cmp.Diff(res, c.Feeds); diff != "" {
			tst.Error(diff)
		}
		if diff := cmp.Diff("imaps://foo.bar:993/INBOX/Feeds", c.Target.String()); diff != "" {
			tst.Error(diff)
		}
	}
}

func TestURLFeedWithDifferentGlobalTarget(tst *testing.T) {
	inp := `
target: imaps://foo.bar/INBOX/Feeds
feeds:
  - name: Foo
    target: imaps://other.bar/INBOX/Feeds
`
	errorText := "while parsing: Line 5: Given URL endpoint 'imaps://other.bar:993' does not match previous endpoint 'imaps://foo.bar:993'."
	c := WithDefault()
	c.FeedOptions = Options{}

	err := c.parse(strings.NewReader(inp))
	if err == nil {
		tst.Error("Expected error.")
	} else {
		if diff := cmp.Diff(errorText, err.Error()); diff != "" {
			tst.Error(diff)
		}
	}
}
