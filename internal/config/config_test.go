package config

import (
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    Config
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func s(s string) *string { return &s }
func b(b bool) *bool     { return &b }

//noinspection GoNilness,GoNilness
func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		inp          string
		wantErr      bool
		feeds        []configGroupFeed
		globalConfig Map
	}{
		{name: "Empty",
			inp: "", wantErr: false, feeds: nil, globalConfig: nil},
		{name: "Trash", inp: "Something", wantErr: true},
		{name: "Simple config",
			inp: "something: 1\nsomething_else: 2", wantErr: false, feeds: nil, globalConfig: Map{"something": 1, "something_else": 2}},
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
			globalConfig: Map{"something": 1},
			feeds: []configGroupFeed{
				{Target: s("bar"), Feed: Feed{
					Name:       "Foo",
					Target:     "",
					Url:        "whatever",
					MinFreq:    0,
					InclImages: b(true),
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
				{Target: nil, Feed: Feed{
					Name:       "Foo",
					Url:        "whatever",
					MinFreq:    2,
					InclImages: nil,
				}},
				{Target: s("bla"), Feed: Feed{
					Name:       "Shrubbery",
					Url:        "google.de",
					MinFreq:    0,
					InclImages: b(false),
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
			feeds:   []configGroupFeed{{Target: s("bla"), Group: Group{"Foo", nil}}},
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
				{Target: nil, Feed: Feed{
					Name: "Foo",
					Url:  "whatever",
				}},
				{Target: s("target"), Group: Group{
					Group: "G1",
					Feeds: []configGroupFeed{
						{Target: s(""), Group: Group{
							Group: "G2",
							Feeds: []configGroupFeed{
								{Target: nil, Feed: Feed{Name: "F1", Url: "google.de"}},
							}},
						},
						{Target: nil, Feed: Feed{Name: "F2"}},
						{Target: nil, Group: Group{Group: "G3"}},
					}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf = []byte(tt.inp)
			got, err := parse(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Feeds, tt.feeds) {
				t.Errorf("parse() got = %v, want %v", got.Feeds, tt.feeds)
			}
			if !reflect.DeepEqual(got.GlobalConfig, tt.globalConfig) {
				t.Errorf("parse() got = %v, want %v", got.GlobalConfig, tt.globalConfig)
			}
		})
	}
}
