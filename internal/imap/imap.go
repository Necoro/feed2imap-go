package imap

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/emersion/go-imap"
	imapClient "github.com/emersion/go-imap/client"

	"github.com/Necoro/feed2imap-go/internal/log"
)

const (
	imapsPort = "993"
	imapPort  = "143"
)

type folders map[string]*imap.MailboxInfo

func (f folders) contains(elem string) bool {
	_, ok := f[elem]
	return ok
}

func (f folders) add(elem *imap.MailboxInfo) {
	name := elem.Name
	f[name] = elem
}

type Client struct {
	c         *imapClient.Client
	host      string
	folders   folders
	delimiter string
	toplevel  string
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

func (client *Client) FolderName(path []string) string {
	return strings.Join(path, client.delimiter)
}

func (client *Client) createFolder(folder string) error {
	err := client.c.Create(folder)
	if err != nil {
		return fmt.Errorf("creating folder '%s': %w", folder, err)
	}

	err = client.c.Subscribe(folder)
	if err != nil {
		return fmt.Errorf("subscribing to folder '%s': %w", folder, err)
	}

	log.Printf("Created folder '%s'", folder)

	return nil
}

func (client *Client) list(folder string) (*imap.MailboxInfo, int, error) {
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- client.c.List("", folder, mailboxes)
	}()

	found := 0
	var mbox *imap.MailboxInfo
	for m := range mailboxes {
		if found == 0 {
			mbox = m
		}
		found++
	}

	if err := <-done; err != nil {
		return nil, 0, fmt.Errorf("while listing '%s': %w", folder, err)
	}

	return mbox, found, nil
}

func (client *Client) selectToplevel() (err error) {
	err = client.EnsureFolder(client.toplevel)

	if err == nil {
		_, err = client.c.Select(client.toplevel, false)
	}

	return
}

func (client *Client) fetchDelimiter() error {
	mbox, _, err := client.list("")
	if err != nil {
		return err
	}

	client.delimiter = mbox.Delimiter
	return nil
}

func (client *Client) EnsureFolder(folder string) error {

	if client.folders.contains(folder) {
		return nil
	}

	log.Printf("Checking for folder '%s'", folder)

	mbox, found, err := client.list(folder)

	switch {
	case err != nil:
		return err
	case found == 0:
		return client.createFolder(folder)
	case found == 1:
		client.folders.add(mbox)
		return nil
	default:
		return fmt.Errorf("Found multiple folders matching '%s'.", folder)
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

	var client = Client{c: c, host: url.Host, folders: folders{}}

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

	if err = client.fetchDelimiter(); err != nil {
		return nil, fmt.Errorf("fetching delimiter: %w", err)
	}

	client.toplevel = url.Path
	if client.toplevel[0] == '/' {
		client.toplevel = client.toplevel[1:]
	}
	client.toplevel = client.FolderName(strings.Split(client.toplevel, "/"))

	log.Printf("Determined '%s' as toplevel, with '%s' as delimiter", client.toplevel, client.delimiter)

	// Go to toplevel folder by default, so that the rest is relative
	if err = client.selectToplevel(); err != nil {
		return nil, err
	}

	return &client, nil
}
