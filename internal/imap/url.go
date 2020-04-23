package imap

import (
	"net"
	"net/url"

	"github.com/Necoro/feed2imap-go/internal/log"
)

// Our own convenience wrapper
type URL struct {
	*url.URL
	// url.URL has no port field and splits it everytime from Host
	port *string
}

const (
	imapsPort   = "993"
	imapPort    = "143"
	imapsSchema = "imaps"
	imapSchema  = "imap"
)

func (url *URL) Port() string {
	if url.port == nil {
		port := url.URL.Port()
		url.port = &port
	}
	return *url.port
}

func (url *URL) ForceTLS() bool {
	return url.Scheme == imapsSchema || url.Port() == imapsPort
}

func (url *URL) setDefaultScheme() {
	switch url.Scheme {
	case imapSchema, imapsSchema:
		return
	default:
		oldScheme := url.Scheme
		if url.Port() == imapsPort {
			url.Scheme = imapsSchema
		} else {
			url.Scheme = imapSchema
		}

		if oldScheme != "" {
			log.Warnf("Unknown scheme '%s', defaulting to '%s'", oldScheme, url.Scheme)
		}
	}
}

func (url *URL) setDefaultPort() {
	if url.Port() == "" {
		var port string
		if url.Scheme == imapsSchema {
			port = imapsPort
		} else {
			port = imapPort
		}
		url.port = &port
		url.Host = net.JoinHostPort(url.Host, port)
	}
}

func (url *URL) sanitizeUrl() {
	url.setDefaultScheme()
	url.setDefaultPort()
}

func NewUrl(url *url.URL) *URL {
	u := URL{URL: url}
	u.sanitizeUrl()
	return &u
}