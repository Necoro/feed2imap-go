package yaml

import (
	"reflect"
	"strings"
	"testing"

	C "github.com/Necoro/feed2imap-go/internal/config"
	F "github.com/Necoro/feed2imap-go/internal/feed"
)

func s(s string) *string { return &s }
func b(b bool) *bool     { return &b }
func t(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ".")
}

func TestBuildFeeds(tst *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		target  string
		feeds   []configGroupFeed
		result  F.Feeds
	}{
		{name: "Empty input", wantErr: false, target: "", feeds: nil, result: F.Feeds{}},
		{name: "Empty Feed", wantErr: true, target: "",
			feeds: []configGroupFeed{
				{Target: s("foo"), Feed: feed{Url: "google.de"}},
			}, result: F.Feeds{}},
		{name: "Empty Feed", wantErr: true, target: "",
			feeds: []configGroupFeed{
				{Target: nil, Feed: feed{Url: "google.de"}},
			}, result: F.Feeds{}},
		{name: "Duplicate Feed Name", wantErr: true, target: "",
			feeds: []configGroupFeed{
				{Target: nil, Feed: feed{Name: "Dup"}},
				{Target: nil, Feed: feed{Name: "Dup"}},
			}, result: F.Feeds{}},
		{name: "Simple", wantErr: false, target: "",
			feeds: []configGroupFeed{
				{Target: s("foo"), Feed: feed{Name: "muh"}},
			},
			result: F.Feeds{"muh": &F.Feed{Name: "muh", Target: t("foo")}},
		},
		{name: "Simple With Target", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: s("foo"), Feed: feed{Name: "muh"}},
			},
			result: F.Feeds{"muh": &F.Feed{Name: "muh", Target: t("moep.foo")}},
		},
		{name: "Simple With Nil Target", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: nil, Feed: feed{Name: "muh"}},
			},
			result: F.Feeds{"muh": &F.Feed{Name: "muh", Target: t("moep.muh")}},
		},
		{name: "Simple With Empty Target", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: s(""), Feed: feed{Name: "muh"}},
			},
			result: F.Feeds{"muh": &F.Feed{Name: "muh", Target: t("moep")}},
		},
		{name: "Multiple Feeds", wantErr: false, target: "moep",
			feeds: []configGroupFeed{
				{Target: s("foo"), Feed: feed{Name: "muh"}},
				{Target: nil, Feed: feed{Name: "bar"}},
			},
			result: F.Feeds{
				"muh": &F.Feed{Name: "muh", Target: t("moep.foo")},
				"bar": &F.Feed{Name: "bar", Target: t("moep.bar")},
			},
		},
		{name: "Empty Group", wantErr: false, target: "",
			feeds: []configGroupFeed{
				{Target: nil, Group: group{Group: "G1"}},
			},
			result: F.Feeds{},
		},
		{name: "Simple Group", wantErr: false, target: "",
			feeds: []configGroupFeed{
				{Target: nil, Group: group{Group: "G1", Feeds: []configGroupFeed{
					{Target: s("bar"), Feed: feed{Name: "F1"}},
					{Target: s(""), Feed: feed{Name: "F2"}},
					{Target: nil, Feed: feed{Name: "F3"}},
				}}},
			},
			result: F.Feeds{
				"F1": &F.Feed{Name: "F1", Target: t("G1.bar")},
				"F2": &F.Feed{Name: "F2", Target: t("G1")},
				"F3": &F.Feed{Name: "F3", Target: t("G1.F3")},
			},
		},
		{name: "Nested Groups", wantErr: false, target: "",
			feeds: []configGroupFeed{
				{Target: nil, Group: group{Group: "G1", Feeds: []configGroupFeed{
					{Target: nil, Feed: feed{Name: "F0"}},
					{Target: s("bar"), Group: group{Group: "G2",
						Feeds: []configGroupFeed{{Target: nil, Feed: feed{Name: "F1"}}}}},
					{Target: s(""), Group: group{Group: "G3",
						Feeds: []configGroupFeed{{Target: s("baz"), Feed: feed{Name: "F2"}}}}},
					{Target: nil, Group: group{Group: "G4",
						Feeds: []configGroupFeed{{Target: nil, Feed: feed{Name: "F3"}}}}},
				}}},
			},
			result: F.Feeds{
				"F0": &F.Feed{Name: "F0", Target: t("G1.F0")},
				"F1": &F.Feed{Name: "F1", Target: t("G1.bar.F1")},
				"F2": &F.Feed{Name: "F2", Target: t("G1.baz")},
				"F3": &F.Feed{Name: "F3", Target: t("G1.G4.F3")},
			},
		},
	}
	for _, tt := range tests {
		tst.Run(tt.name, func(tst *testing.T) {
			var feeds F.Feeds = F.Feeds{}
			err := buildFeeds(tt.feeds, t(tt.target), feeds)
			if (err != nil) != tt.wantErr {
				tst.Errorf("buildFeeds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(feeds, tt.result) {
				tst.Errorf("buildFeeds() got = %v, want %v", feeds, tt.result)
			}
		})
	}
}

//noinspection GoNilness,GoNilness
func TestParse(tst *testing.T) {
	tests := []struct {
		name         string
		inp          string
		wantErr      bool
		feeds        []configGroupFeed
		globalConfig C.Map
	}{
		{name: "Empty",
			inp: "", wantErr: false, feeds: nil, globalConfig: nil},
		{name: "Trash", inp: "Something", wantErr: true},
		{name: "Simple config",
			inp: "something: 1\nsomething_else: 2", wantErr: false, feeds: nil, globalConfig: C.Map{"something": 1, "something_else": 2}},
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
			wantErr:      false,
			globalConfig: C.Map{"something": 1},
			feeds: []configGroupFeed{
				{Target: s("bar"), Feed: feed{
					Name: "Foo",
					Url:  "whatever",
					Options: C.Options{
						MinFreq:    0,
						InclImages: b(true),
					},
				}}}},

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
			feeds: []configGroupFeed{
				{Target: nil, Feed: feed{
					Name: "Foo",
					Url:  "whatever",
					Options: C.Options{
						MinFreq:    2,
						InclImages: nil,
					},
				}},
				{Target: s("bla"), Feed: feed{
					Name: "Shrubbery",
					Url:  "google.de",
					Options: C.Options{
						MinFreq:    0,
						InclImages: b(false),
					},
				}},
			},
		},
		{name: "Empty Group",
			inp: `
feeds:
   - group: Foo
     target: bla
`,
			wantErr: false,
			feeds:   []configGroupFeed{{Target: s("bla"), Group: group{"Foo", nil}}},
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
			feeds: []configGroupFeed{
				{Target: nil, Feed: feed{
					Name: "Foo",
					Url:  "whatever",
				}},
				{Target: s("target"), Group: group{
					Group: "G1",
					Feeds: []configGroupFeed{
						{Target: s(""), Group: group{
							Group: "G2",
							Feeds: []configGroupFeed{
								{Target: nil, Feed: feed{Name: "F1", Url: "google.de"}},
							}},
						},
						{Target: nil, Feed: feed{Name: "F2"}},
						{Target: nil, Group: group{Group: "G3"}},
					}},
				},
			},
		},
	}

	for _, tt := range tests {
		tst.Run(tt.name, func(tst *testing.T) {
			var buf = []byte(tt.inp)
			got, err := parse(buf)
			if (err != nil) != tt.wantErr {
				tst.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Feeds, tt.feeds) {
				tst.Errorf("parse() got = %v, want %v", got.Feeds, tt.feeds)
			}
			if !reflect.DeepEqual(got.GlobalConfig, tt.globalConfig) {
				tst.Errorf("parse() got = %v, want %v", got.GlobalConfig, tt.globalConfig)
			}
		})
	}
}
