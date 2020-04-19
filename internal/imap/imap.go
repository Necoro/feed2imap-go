package imap

import (
	"fmt"
	"net/url"

	"github.com/emersion/go-imap"
	imapClient "github.com/emersion/go-imap/client"

	"github.com/Necoro/feed2imap-go/internal/log"
)

const (
	imapsPort = "993"
	imapPort  = "143"
)

type Client struct {
	c    *imapClient.Client
	host string
}

func forceTLS(url *url.URL) bool {
	return url.Scheme == "imaps" || url.Port() == imapsPort
}

func setDefaultScheme(url *url.URL) {
	switch url.Scheme {
	case "imap", "imaps":
		return
	default:
		oldScheme := url.Scheme
		if url.Port() == imapsPort {
			url.Scheme = "imaps"
		} else {
			url.Scheme = "imap"
		}

		if oldScheme != "" {
			log.Warnf("Unknown scheme '%s', defaulting to '%s'", oldScheme, url.Scheme)
		}
	}
}

func setDefaultPort(url *url.URL) {
	if url.Port() == "" {
		var port string
		if url.Scheme == "imaps" {
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

func (client *Client) Disconnect() {
	if client != nil {
		connected := (client.c.State() & imap.ConnectedState) != 0
		_ = client.c.Logout()

		if connected {
			log.Print("Disconnected from ", client.host)
		}
	}
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

	var client = Client{c, url.Host}

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
		}
		log.Print("Connected to ", url.Host, " (Plain)")
	}

	pwd, _ := url.User.Password()
	if err = c.Login(url.User.Username(), pwd); err != nil {
		return nil, fmt.Errorf("login to %s: %w", url.Host, err)
	}

	return &client, nil
}
