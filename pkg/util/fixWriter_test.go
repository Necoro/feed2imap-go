package util

import (
	"bytes"
	"io"
	"testing"
)

func TestFixWriter_Write(t *testing.T) {
	tests := []struct {
		before string
		after  string
	}{
		{"", ""},
		{"foo", "foo"},
		{"foo\r", "foo\r\n"},
		{"foo\n", "foo\r\n"},
		{"foo\r\n", "foo\r\n"},
		{"\r", "\r\n"},
		{"\n", "\r\n"},
		{"\r\n", "\r\n"},
		{"foo\rbar", "foo\r\nbar"},
		{"foo\nbar", "foo\r\nbar"},
		{"foo\r\nbar", "foo\r\nbar"},
		{"\r\r", "\r\n\r\n"},
		{"\n\n", "\r\n\r\n"},
		{"\r\r\n", "\r\n\r\n"},
		{"\n\r", "\r\n\r\n"},
		{"\rbar", "\r\nbar"},
		{"\nbar", "\r\nbar"},
		{"\r\nbar", "\r\nbar"},
	}
	for _, tt := range tests {
		t.Run(tt.before, func(t *testing.T) {
			b := bytes.Buffer{}
			w := FixWriter(&b)
			if _, err := io.WriteString(w, tt.before); err != nil {
				t.Errorf("Error: %v", err)
				return
			}
			res := b.String()
			if tt.after != res {
				t.Errorf("Expected: %q, got: %q", tt.after, res)
			}
		})
	}
}
