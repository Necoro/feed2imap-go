package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

func TestUrl_Unmarshal(t *testing.T) {

	tests := []struct {
		name    string
		inp     string
		url     Url
		wantErr bool
		str     string
	}{
		{name: "Empty", inp: `url: ""`, wantErr: true},
		{name: "Simple String", inp: `url: "imap://user:pass@example.net:143/INBOX"`, url: Url{
			Scheme:   "imap",
			User:     "user",
			Password: "pass",
			Host:     "example.net",
			Port:     "143",
			Root:     "/INBOX",
		}, str: "imap://user:******@example.net:143/INBOX"},
		{name: "Simple String with @", inp: `url: "imaps://user@example:pass@example.net:143/INBOX"`, url: Url{
			Scheme:   "imaps",
			User:     "user@example",
			Password: "pass",
			Host:     "example.net",
			Port:     "143",
			Root:     "/INBOX",
		}, str: "imaps://user@example:******@example.net:143/INBOX"},
		{name: "Simple String with %40", inp: `url: "imap://user%40example:pass@example.net:4711/INBOX"`, url: Url{
			Scheme:   "imap",
			User:     "user@example",
			Password: "pass",
			Host:     "example.net",
			Port:     "4711",
			Root:     "/INBOX",
		}, str: "imap://user@example:******@example.net:4711/INBOX"},
		{name: "Simple String without user", inp: `url: "imap://example.net:143/INBOX"`, url: Url{
			Scheme:   "imap",
			User:     "",
			Password: "",
			Host:     "example.net",
			Port:     "143",
			Root:     "/INBOX",
		}, str: "imap://example.net:143/INBOX"},
		{name: "Err: Inv scheme", inp: `url: "smtp://user%40example:pass@example.net:4711/INBOX"`, wantErr: true},
		{name: "Err: No Host", inp: `url: "imap://user%40example:pass/INBOX"`, wantErr: true},
		{name: "Err: Scheme Only", inp: `url: "imap://"`, wantErr: true},
		{name: "No Root", inp: `url: "imap://user:pass@example.net:143"`, url: Url{
			Scheme:   "imap",
			User:     "user",
			Password: "pass",
			Host:     "example.net",
			Port:     "143",
			Root:     "",
		}, str: "imap://user:******@example.net:143"},
		{name: "No Root: Slash", inp: `url: "imap://user:pass@example.net:143/"`, url: Url{
			Scheme:   "imap",
			User:     "user",
			Password: "pass",
			Host:     "example.net",
			Port:     "143",
			Root:     "/",
		}, str: "imap://user:******@example.net:143/"},
		{name: "Full", inp: `url:
  scheme: imap
  host: example.net
  user: user
  password: p4ss
  port: 143
  root: INBOX
`, url: Url{
			Scheme:   "imap",
			User:     "user",
			Password: "p4ss",
			Host:     "example.net",
			Port:     "143",
			Root:     "INBOX",
		}, str: "imap://user:******@example.net:143/INBOX"},
		{name: "Default Port", inp: `url:
  scheme: imap
  host: example.net
  user: user
  password: p4ss
  root: INBOX
`, url: Url{
			Scheme:   "imap",
			User:     "user",
			Password: "p4ss",
			Host:     "example.net",
			Port:     "143",
			Root:     "INBOX",
		}, str: "imap://user:******@example.net:143/INBOX"},
		{name: "Default Scheme", inp: `url:
  host: example.net
  user: user
  password: p4ss
  port: 993
  root: INBOX
`, url: Url{
			Scheme:   "imaps",
			User:     "user",
			Password: "p4ss",
			Host:     "example.net",
			Port:     "993",
			Root:     "INBOX",
		}, str: "imaps://user:******@example.net:993/INBOX"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u struct {
				Url Url `yaml:"url"`
			}
			err := yaml.Unmarshal([]byte(tt.inp), &u)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(u.Url, tt.url); err == nil && diff != "" {
				t.Error(diff)
			}

			if diff := cmp.Diff(u.Url.String(), tt.str); err == nil && diff != "" {
				t.Error(diff)
			}
		})
	}
}
