package imap

import (
	"fmt"
	"net/url"
	"strings"

	imapClient "github.com/emersion/go-imap/client"

	"github.com/Necoro/feed2imap-go/pkg/log"
)

func newImapClient(url *URL, forceTls bool) (*imapClient.Client, error) {
	var (
		c   *imapClient.Client
		err error
	)

	if forceTls {
		if c, err = imapClient.DialTLS(url.Host, nil); err != nil {
			return nil, fmt.Errorf("connecting (TLS) to %s: %w", url.Host, err)
		}
		log.Print("Connected to ", url.Host, " (TLS)")
	} else {
		if c, err = imapClient.Dial(url.Host); err != nil {
			return nil, fmt.Errorf("connecting to %s: %w", url.Host, err)
		}
	}

	return c, nil
}

func (cl *Client) connect(url *URL, forceTls bool) (*connection, error) {
	c, err := newImapClient(url, forceTls)
	if err != nil {
		return nil, err
	}

	conn := cl.createConnection(c)

	if !forceTls {
		if err = conn.startTls(); err != nil {
			return nil, err
		}
	}

	pwd, _ := url.User.Password()
	if err = c.Login(url.User.Username(), pwd); err != nil {
		return nil, fmt.Errorf("login to %s: %w", url.Host, err)
	}

	return conn, nil
}

func Connect(_url *url.URL) (*Client, error) {
	var err error
	url := NewUrl(_url)
	forceTls := url.ForceTLS()

	client := NewClient()
	client.host = url.Host
	defer func() {
		if err != nil {
			client.Disconnect()
		}
	}()

	var conn *connection // the main connection
	if conn, err = client.connect(url, forceTls); err != nil {
		return nil, err
	}

	delim, err := conn.fetchDelimiter()
	if err != nil {
		return nil, fmt.Errorf("fetching delimiter: %w", err)
	}
	client.delimiter = delim

	toplevel := url.Path
	if toplevel[0] == '/' {
		toplevel = toplevel[1:]
	}
	client.toplevel = client.folderName(strings.Split(toplevel, "/"))

	log.Printf("Determined '%s' as toplevel, with '%s' as delimiter", client.toplevel, client.delimiter)

	if err = conn.ensureFolder(client.toplevel); err != nil {
		return nil, err
	}

	// the other connections
	for i := 1; i < len(client.connections); i++ {
		if _, err := client.connect(url, forceTls); err != nil { // explicitly new var 'err', b/c these are now harmless
			log.Warnf("connecting #%d: %s", i, err)
		}
	}

	client.startCommander()

	return client, nil
}
