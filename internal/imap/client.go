package imap

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	uidplus "github.com/emersion/go-imap-uidplus"
	imapClient "github.com/emersion/go-imap/client"

	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
)

const numberConns = 5

type connConf struct {
	host      string
	delimiter string
	toplevel  Folder
}

type Client struct {
	connConf
	mailboxes   *mailboxes
	commander   *commander
	connections [numberConns]*connection
	idxCounter  int32
	connChannel chan *connection
}

var dialer imapClient.Dialer

func init() {
	dialer = &net.Dialer{Timeout: 30 * time.Second}
}

func newImapClient(url config.Url) (c *imapClient.Client, err error) {
	if url.ForceTLS() {
		if c, err = imapClient.DialWithDialerTLS(dialer, url.HostPort(), nil); err != nil {
			return nil, fmt.Errorf("connecting (TLS) to %s: %w", url.Host, err)
		}
		log.Print("Connected to ", url.HostPort(), " (TLS)")
	} else {
		if c, err = imapClient.DialWithDialer(dialer, url.HostPort()); err != nil {
			return nil, fmt.Errorf("connecting to %s: %w", url.Host, err)
		}
	}

	return
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

func (cl *Client) Disconnect() {
	if cl != nil {
		cl.stopCommander()
		close(cl.connChannel)

		connected := false
		for _, conn := range cl.connections {
			connected = conn.disconnect() || connected
		}

		if connected {
			log.Print("Disconnected from ", cl.host)
		}
	}
}

func (cl *Client) createConnection(c *imapClient.Client) *connection {
	nextIndex := int(atomic.AddInt32(&cl.idxCounter, 1)) - 1

	if nextIndex >= len(cl.connections) {
		panic("Too many connections")
	}

	client := &client{c, uidplus.NewClient(c)}

	conn := &connection{
		connConf:  &cl.connConf,
		mailboxes: cl.mailboxes,
		c:         client,
	}

	cl.connections[nextIndex] = conn

	return conn
}

func newClient() *Client {
	return &Client{
		mailboxes:   NewMailboxes(),
		connChannel: make(chan *connection, 0),
	}
}
