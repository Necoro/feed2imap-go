package imap

import (
	"sync/atomic"

	uidplus "github.com/emersion/go-imap-uidplus"
	imapClient "github.com/emersion/go-imap/client"

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

func NewClient() *Client {
	return &Client{
		mailboxes:   NewMailboxes(),
		connChannel: make(chan *connection, 0),
	}
}
