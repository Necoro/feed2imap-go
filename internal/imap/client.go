package imap

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-imap"
	imapClient "github.com/emersion/go-imap/client"

	"github.com/Necoro/feed2imap-go/internal/log"
)

const numberConns = 5

type connConf struct {
	host      string
	delimiter string
	toplevel  Folder
}

type connection struct {
	*connConf
	mailboxes *mailboxes
	c         *imapClient.Client
}

type mailboxes struct {
	mb map[string]*imap.MailboxInfo
	mu sync.RWMutex
}

type Client struct {
	connConf
	mailboxes mailboxes
	commander *commander
	connections [numberConns]*connection
	nextFreeIndex int
}

type Folder struct {
	str       string
	delimiter string
}

func (f Folder) String() string {
	return f.str
}

func (f Folder) Append(other Folder) Folder {
	if f.delimiter != other.delimiter {
		panic("Delimiters do not match")
	}
	return Folder{
		str:       f.str + f.delimiter + other.str,
		delimiter: f.delimiter,
	}
}

func (mbs *mailboxes) contains(elem Folder) bool {
	mbs.mu.RLock()
	defer mbs.mu.RUnlock()

	_, ok := mbs.mb[elem.str]
	return ok
}

func (mbs *mailboxes) add(elem *imap.MailboxInfo) {
	mbs.mu.Lock()
	defer mbs.mu.Unlock()

	mbs.mb[elem.Name] = elem
}

func (conn *connection) Disconnect() bool {
	if conn != nil {
		connected := (conn.c.State() & imap.ConnectedState) != 0
		_ = conn.c.Logout()
		return connected
	}
	return false
}

func (client *Client) Disconnect() {
	if client != nil {
		client.stopCommander()

		connected := false
		for _, conn := range client.connections {
			connected = conn.Disconnect() || connected
		}

		if connected {
			log.Print("Disconnected from ", client.host)
		}
	}
}

func (client *Client) folderName(path []string) Folder {
	return Folder{
		strings.Join(path, client.delimiter),
		client.delimiter,
	}
}

func (client *Client) NewFolder(path []string) Folder {
	return client.toplevel.Append(client.folderName(path))
}

func (conn *connection) createFolder(folder string) error {
	err := conn.c.Create(folder)
	if err != nil {
		return fmt.Errorf("creating folder '%s': %w", folder, err)
	}

	err = conn.c.Subscribe(folder)
	if err != nil {
		return fmt.Errorf("subscribing to folder '%s': %w", folder, err)
	}

	log.Printf("Created folder '%s'", folder)

	return nil
}

func (conn *connection) list(folder string) (*imap.MailboxInfo, int, error) {
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- conn.c.List("", folder, mailboxes)
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

func (conn *connection) fetchDelimiter() (string, error) {
	mbox, _, err := conn.list("")
	if err != nil {
		return "", err
	}

	return mbox.Delimiter, nil
}

func (conn *connection) ensureFolder(folder Folder) error {
	if conn.mailboxes.contains(folder) {
		return nil
	}

	log.Printf("Checking for folder '%s'", folder)

	mbox, found, err := conn.list(folder.str)
	if err != nil {
		return err
	}

	if mbox != nil && mbox.Delimiter != folder.delimiter {
		panic("Delimiters do not match")
	}

	switch found {
	case 0:
		return conn.createFolder(folder.str)
	case 1:
		conn.mailboxes.add(mbox)
		return nil
	default:
		return fmt.Errorf("Found multiple folders matching '%s'.", folder)
	}
}

func (client *Client) EnsureFolder(folder Folder) error {
	return client.commander.execute(ensureCommando{folder})
}

func (conn *connection) putMessages(folder Folder, messages []string) error {
	if len(messages) == 0 {
		return nil
	}

	now := time.Now()
	for _, msg := range messages {
		reader := strings.NewReader(msg)
		if err := conn.c.Append(folder.str, nil, now, reader); err != nil {
			return fmt.Errorf("uploading message to %s: %w", folder, err)
		}
	}

	return nil
}

func (client *Client) PutMessages(folder Folder, messages []string) error {
	return client.commander.execute(addCommando{folder, messages})
}

func (client *Client) createConnection(c *imapClient.Client) *connection{
	if client.nextFreeIndex >= len(client.connections) {
		panic("Too many connections")
	}

	conn := &connection{
		connConf:  &client.connConf,
		mailboxes: &client.mailboxes,
		c:         c,
	}

	client.connections[client.nextFreeIndex] = conn
	client.nextFreeIndex++

	return conn
}

func (conn *connection) startTls() error {
	hasStartTls, err := conn.c.SupportStartTLS()
	if err != nil {
		return fmt.Errorf("checking for starttls for %s: %w", conn.host, err)
	}

	if hasStartTls {
		if err = conn.c.StartTLS(nil); err != nil {
			return fmt.Errorf("enabling starttls for %s: %w", conn.host, err)
		}

		log.Print("Connected to ", conn.host, " (STARTTLS)")
	} else {
		log.Print("Connected to ", conn.host, " (Plain)")
	}

	return nil
}

func NewClient() *Client {
	return &Client{
		mailboxes: mailboxes{
			mb: map[string]*imap.MailboxInfo{},
			mu: sync.RWMutex{},
		},
	}
}