package imap

import (
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

func (client *Client) Disconnect() {
	if client != nil {
		client.stopCommander()

		connected := false
		for _, conn := range client.connections {
			connected = conn.disconnect() || connected
		}

		if connected {
			log.Print("Disconnected from ", client.host)
		}
	}
}

func (client *Client) createConnection(c *imapClient.Client) *connection {
	if client.nextFreeIndex >= len(client.connections) {
		panic("Too many connections")
	}

	conn := &connection{
		connConf:  &client.connConf,
		mailboxes: client.mailboxes,
		c:         c,
	}

	client.connections[client.nextFreeIndex] = conn
	client.nextFreeIndex++

	return conn
}

func NewClient() *Client {
	return &Client{mailboxes: NewMailboxes()}
}
