package rfc822

import (
	"bytes"
	"io"
	"testing"
)

func TestRfc822Writer_Write(t *testing.T) {
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
			w := Writer(&b)
			n, err := io.WriteString(w, tt.before)
			if err != nil {
				t.Errorf("Error: %v", err)
				return
			}
			if n != len(tt.before) {
				t.Errorf("Unexpected number of bytes written: %d, expected: %d", n, len(tt.before))
			}
			res := b.String()
			if tt.after != res {
				t.Errorf("Expected: %q, got: %q", tt.after, res)
			}
		})
	}
}
