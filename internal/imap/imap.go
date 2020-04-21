package imap

import (
	"fmt"
	"net/url"
	"strings"

	imapClient "github.com/emersion/go-imap/client"

	"github.com/Necoro/feed2imap-go/internal/log"
)

const (
	imapsPort   = "993"
	imapPort    = "143"
	imapsSchema = "imaps"
	imapSchema  = "imap"
)

func forceTLS(url *url.URL) bool {
	return url.Scheme == imapsSchema || url.Port() == imapsPort
}

func setDefaultScheme(url *url.URL) {
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

func setDefaultPort(url *url.URL) {
	if url.Port() == "" {
		var port string
		if url.Scheme == imapsSchema {
			port = imapsPort
		} else {
			port = imapPort
		}
		url.Host += ":" + port
	}
}

func sanitizeUrl(url *url.URL) {
	setDefaultScheme(url)
	setDefaultPort(url)
}

func Connect(url *url.URL) (*Client, error) {
	var c *imapClient.Client
	var err error

	sanitizeUrl(url)

	forceTls := forceTLS(url)

	if forceTls {
		c, err = imapClient.DialTLS(url.Host, nil)
		if err != nil {
			return nil, fmt.Errorf("connecting (TLS) to %s: %w", url.Host, err)
		}
		log.Print("Connected to ", url.Host, " (TLS)")
	} else {
		c, err = imapClient.Dial(url.Host)
		if err != nil {
			return nil, fmt.Errorf("connecting to %s: %w", url.Host, err)
		}
	}

	var client = Client{c: c, host: url.Host, mailboxes: mailboxes{}}

	defer func() {
		if err != nil {
			client.Disconnect()
		}
	}()

	if !forceTls {
		var hasStartTls bool // explicit to avoid shadowing err

		hasStartTls, err = c.SupportStartTLS()
		if err != nil {
			return nil, fmt.Errorf("checking for starttls for %s: %w", url.Host, err)
		}

		if hasStartTls {
			if err = c.StartTLS(nil); err != nil {
				return nil, fmt.Errorf("enabling starttls for %s: %w", url.Host, err)
			}

			log.Print("Connected to ", url.Host, " (STARTTLS)")
		} else {
			log.Print("Connected to ", url.Host, " (Plain)")
		}
	}

	pwd, _ := url.User.Password()
	if err = c.Login(url.User.Username(), pwd); err != nil {
		return nil, fmt.Errorf("login to %s: %w", url.Host, err)
	}

	if err = client.fetchDelimiter(); err != nil {
		return nil, fmt.Errorf("fetching delimiter: %w", err)
	}

	toplevel := url.Path
	if toplevel[0] == '/' {
		toplevel = toplevel[1:]
	}
	client.toplevel = client.folderName(strings.Split(toplevel, "/"))

	log.Printf("Determined '%s' as toplevel, with '%s' as delimiter", client.toplevel, client.delimiter)

	if err = client.EnsureFolder(client.toplevel); err != nil {
		return nil, err
	}

	return &client, nil
}
