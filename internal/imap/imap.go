package imap

import (
	"fmt"
	"strings"

	imapClient "github.com/emersion/go-imap/client"

	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
)

func newImapClient(url config.Url) (*imapClient.Client, error) {
	var (
		c   *imapClient.Client
		err error
	)

	if url.ForceTLS() {
		if c, err = imapClient.DialTLS(url.HostPort(), nil); err != nil {
			return nil, fmt.Errorf("connecting (TLS) to %s: %w", url.Host, err)
		}
		log.Print("Connected to ", url.HostPort(), " (TLS)")
	} else {
		if c, err = imapClient.Dial(url.HostPort()); err != nil {
			return nil, fmt.Errorf("connecting to %s: %w", url.Host, err)
		}
	}

	return c, nil
}

func (cl *Client) connect(url config.Url) (*connection, error) {
	c, err := newImapClient(url)
	if err != nil {
		return nil, err
	}

	conn := cl.createConnection(c)

	if !url.ForceTLS() {
		if err = conn.startTls(); err != nil {
			return nil, err
		}
	}

	if err = c.Login(url.User, url.Password); err != nil {
		return nil, fmt.Errorf("login to %s: %w", url.Host, err)
	}

	cl.connChannel <- conn
	return conn, nil
}

func Connect(url config.Url) (*Client, error) {
	var err error

	client := NewClient()
	client.host = url.Host
	defer func() {
		if err != nil {
			client.Disconnect()
		}
	}()
	client.startCommander()

	var conn *connection // the main connection
	if conn, err = client.connect(url); err != nil {
		return nil, err
	}

	delim, err := conn.fetchDelimiter()
	if err != nil {
		return nil, fmt.Errorf("fetching delimiter: %w", err)
	}
	client.delimiter = delim

	toplevel := url.Root
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
		go func(id int) {
			if _, err := client.connect(url); err != nil { // explicitly new var 'err', b/c these are now harmless
				log.Warnf("connecting #%d: %s", id, err)
			}
		}(i)
	}

	return client, nil
}
