package config

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

func i(i int) *int   { return &i }
func b(b bool) *bool { return &b }
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

func TestBuildFeeds(tst *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		target  string
		feeds   []configGroupFeed
		result  Feeds
	}{
		{name: "Empty input", wantErr: false, target: "", feeds: nil, result: Feeds{}},
		{name: "Empty Feed", wantErr: true, target: "",
			feeds: []configGroupFeed{
				{Target: n("foo"), Feed: Feed{Url: "google.de"}},
			}, result: Feeds{}},
		{name: "Empty Feed", wantErr: true, target: "",
			feeds: []configGroupFeed{
				{Feed: Feed{Url: "google.de"}},
			}, result: Feeds{}},
		{name: "Duplicate Feed Name", wantErr: true, target: "",
			feeds: []configGroupFeed{
				{Feed: Feed{Name: "Dup"}},
				{Feed: Feed{Name: "Dup"}},
			}, result: Feeds{}},
		{name: "Simple", wantErr: false, target: "",
			feeds: []configGroupFeed{
				{Target: n("foo"), Feed: Feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("foo")}},
		},
		{name: "Simple With Target", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: n("foo"), Feed: Feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("moep.foo")}},
		},
		{name: "Simple Without Target", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Feed: Feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("moep.muh")}},
		},
		{name: "Simple With Nil Target", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: yaml.Node{Tag: "!!null"}, Feed: Feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("moep")}},
		},
		{name: "Simple With Empty Target", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: n(""), Feed: Feed{Name: "muh"}},
			},
			result: Feeds{"muh": &Feed{Name: "muh", Target: t("moep")}},
		},
		{name: "Multiple Feeds", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: n("foo"), Feed: Feed{Name: "muh"}},
				{Feed: Feed{Name: "bar"}},
			},
			result: Feeds{
				"muh": &Feed{Name: "muh", Target: t("moep.foo")},
				"bar": &Feed{Name: "bar", Target: t("moep.bar")},
			},
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
					{Target: n("bar"), Feed: Feed{Name: "F1"}},
					{Target: n(""), Feed: Feed{Name: "F2"}},
					{Feed: Feed{Name: "F3"}},
				}}},
			},
			result: Feeds{
				"F1": &Feed{Name: "F1", Target: t("G1.bar")},
				"F2": &Feed{Name: "F2", Target: t("G1")},
				"F3": &Feed{Name: "F3", Target: t("G1.F3")},
			},
		},
		{name: "Nested Groups", wantErr: false, target: "",
			feeds: []configGroupFeed{
				{Group: group{Group: "G1", Feeds: []configGroupFeed{
					{Feed: Feed{Name: "F0"}},
					{Target: n("bar"), Group: group{Group: "G2",
						Feeds: []configGroupFeed{{Feed: Feed{Name: "F1"}}}}},
					{Target: n(""), Group: group{Group: "G3",
						Feeds: []configGroupFeed{{Target: n("baz"), Feed: Feed{Name: "F2"}}}}},
					{Group: group{Group: "G4",
						Feeds: []configGroupFeed{{Feed: Feed{Name: "F3"}}}}},
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
			err := buildFeeds(tt.feeds, t(tt.target), feeds)
			if (err != nil) != tt.wantErr {
				tst.Errorf("buildFeeds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(feeds, tt.result); !tt.wantErr && diff != "" {
				tst.Error(diff)
			}
		})
	}
}

func defaultConfig(feeds []configGroupFeed, global Map) config {
	defCfg := WithDefault()
	if global != nil {
		defCfg.GlobalConfig = global
	}
	return config{
		Config:       defCfg,
		Feeds:        feeds,
		GlobalConfig: global,
	}
}

//noinspection GoNilness,GoNilness
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
				c.FeedOptions.MinFreq = i(6)
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
			config: defaultConfig([]configGroupFeed{
				{Target: n("bar"), Feed: Feed{
					Name: "Foo",
					Url:  "whatever",
					Options: Options{
						MinFreq:    nil,
						InclImages: b(true),
					},
				}}}, Map{"something": 1})},

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
				{Feed: Feed{
					Name: "Foo",
					Url:  "whatever",
					Options: Options{
						MinFreq:    i(2),
						InclImages: nil,
					},
				}},
				{Target: n("bla"), Feed: Feed{
					Name: "Shrubbery",
					Url:  "google.de",
					Options: Options{
						MinFreq:    nil,
						InclImages: b(false),
					},
				}},
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
      - group: G3
`,
			wantErr: false,
			config: defaultConfig([]configGroupFeed{
				{Feed: Feed{
					Name: "Foo",
					Url:  "whatever",
				}},
				{Target: n("target"), Group: group{
					Group: "G1",
					Feeds: []configGroupFeed{
						{Target: n(""), Group: group{
							Group: "G2",
							Feeds: []configGroupFeed{
								{Feed: Feed{Name: "F1", Url: "google.de"}},
							}},
						},
						{Feed: Feed{Name: "F2"}},
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
			var buf = []byte(tt.inp)
			got, err := unmarshal(buf, WithDefault())
			if (err != nil) != tt.wantErr {
				tst.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				diff := cmp.Diff(got, tt.config, eqNode)
				if diff != "" {
					tst.Error(diff)
				}
			}
		})
	}
}
