package template

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestByteCount(t *testing.T) {
	tests := map[string]struct {
		inp string
		out string
	}{
		"Empty":        {"", "0 B"},
		"Byte":         {"123", "123 B"},
		"KByte":        {"2048", "2.0 KB"},
		"KByte slight": {"2049", "2.0 KB"},
		"KByte round":  {"2560", "2.5 KB"},
		"MByte":        {"2097152", "2.0 MB"},
	}

	for name, tt := range tests {
		t.Run(name, func(tst *testing.T) {
			out := byteCount(tt.inp)

			if diff := cmp.Diff(tt.out, out); diff != "" {
				tst.Error(diff)
			}
		})
	}
}

func TestDict(t *testing.T) {
	type i []interface{}
	type o map[string]interface{}

	tests := map[string]struct {
		inp i
		out o
	}{
		"Empty": {i{}, o{}},
		"One":   {i{"1"}, o{"1": ""}},
		"Two":   {i{"1", 1}, o{"1": 1}},
		"Three": {i{"1", "2", "3"}, o{"1": "2", "3": ""}},
		"Four":  {i{"1", 2, "3", '4'}, o{"1": 2, "3": '4'}},
	}

	for name, tt := range tests {
		t.Run(name, func(tst *testing.T) {
			out := dict(tt.inp...)

			if diff := cmp.Diff(tt.out, o(out)); diff != "" {
				tst.Error(diff)
			}
		})
	}
}
