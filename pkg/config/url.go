package config

import (
	"fmt"
	"net"
	"net/url"

	"gopkg.in/yaml.v3"
)

type Url struct {
	Scheme   string `yaml:"scheme"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Root     string `yaml:"root"`
}

func (u *Url) Empty() bool {
	return u.Host == ""
}

func (u *Url) EmptyRoot() bool {
	return u.Root == "" || u.Root == "/"
}

func (u *Url) UnmarshalYAML(value *yaml.Node) (err error) {
	if value.ShortTag() == strTag {
		var val string
		var rawUrl *url.URL

		if err = value.Decode(&val); err != nil {
			return err
		}
		if rawUrl, err = url.Parse(val); err != nil {
			return err
		}

		u.Scheme = rawUrl.Scheme
		u.User = rawUrl.User.Username()
		u.Password, _ = rawUrl.User.Password()
		u.Host = rawUrl.Hostname()
		u.Port = rawUrl.Port()
		u.Root = rawUrl.Path
	} else {
		type _url Url // avoid recursion
		wrapped := (*_url)(u)
		if err = value.Decode(wrapped); err != nil {
			return err
		}
	}

	u.sanitize()

	if errors := u.validate(); len(errors) > 0 {
		errs := make([]string, len(errors)+1)
		copy(errs[1:], errors)
		errs[0] = fmt.Sprintf("line %d: Invalid target:", value.Line)
		return &yaml.TypeError{Errors: errs}
	}

	return nil
}

func (u *Url) String() string {
	var pwd string
	if u.Password != "" {
		pwd = ":******"
	}

	return fmt.Sprintf("%s://%s%s@%s%s", u.Scheme, u.User, pwd, u.HostPort(), u.Root)
}

func (u *Url) HostPort() string {
	if u.Port != "" {
		return net.JoinHostPort(u.Host, u.Port)
	}
	return u.Host
}

const (
	imapsPort   = "993"
	imapPort    = "143"
	imapsSchema = "imaps"
	imapSchema  = "imap"
)

func (u *Url) ForceTLS() bool {
	return u.Scheme == imapsSchema || u.Port == imapsPort
}

func (u *Url) setDefaultScheme() {
	if u.Scheme == "" {
		if u.Port == imapsPort {
			u.Scheme = imapsSchema
		} else {
			u.Scheme = imapSchema
		}
	}
}

func (u *Url) setDefaultPort() {
	if u.Port == "" {
		if u.Scheme == imapsSchema {
			u.Port = imapsPort
		} else {
			u.Port = imapPort
		}
	}
}

func (u *Url) sanitize() {
	u.setDefaultScheme()
	u.setDefaultPort()
}

func (u *Url) validate() (errors []string) {
	if u.Scheme != imapSchema && u.Scheme != imapsSchema {
		errors = append(errors, fmt.Sprintf("Unknown scheme %q", u.Scheme))
	}

	if u.Host == "" {
		errors = append(errors, "Host not set")
	}

	return
}
