package imap

import (
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
	mailboxes     *mailboxes
	commander     *commander
	connections   [numberConns]*connection
	nextFreeIndex int
}

func (cl *Client) Disconnect() {
	if cl != nil {
		cl.stopCommander()

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
	if cl.nextFreeIndex >= len(cl.connections) {
		panic("Too many connections")
	}

	client := &client{c, uidplus.NewClient(c)}

	conn := &connection{
		connConf:  &cl.connConf,
		mailboxes: cl.mailboxes,
		c:         client,
	}

	cl.connections[cl.nextFreeIndex] = conn
	cl.nextFreeIndex++

	return conn
}

func NewClient() *Client {
	return &Client{mailboxes: NewMailboxes()}
}
