package imap

import (
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	imapClient "github.com/emersion/go-imap/client"

	"github.com/Necoro/feed2imap-go/internal/log"
)

type Client struct {
	c         *imapClient.Client
	host      string
	mailboxes mailboxes
	delimiter string
	toplevel  Folder
	commander *commander
}

type Folder struct {
	str       string
	delimiter string
}
type mailboxes map[string]*imap.MailboxInfo

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

func (mbs mailboxes) contains(elem Folder) bool {
	_, ok := mbs[elem.str]
	return ok
}

func (mbs mailboxes) add(elem *imap.MailboxInfo) {
	mbs[elem.Name] = elem
}

func (client *Client) Disconnect() {
	if client != nil {
		client.stopCommander()

		connected := (client.c.State() & imap.ConnectedState) != 0
		_ = client.c.Logout()

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

func (client *Client) fetchDelimiter() error {
	mbox, _, err := client.list("")
	if err != nil {
		return err
	}

	client.delimiter = mbox.Delimiter
	return nil
}

func (client *Client) ensureFolder(folder Folder) error {
	if client.mailboxes.contains(folder) {
		return nil
	}

	log.Printf("Checking for folder '%s'", folder)

	mbox, found, err := client.list(folder.str)
	if err != nil {
		return err
	}

	if mbox != nil && mbox.Delimiter != folder.delimiter {
		panic("Delimiters do not match")
	}

	switch found {
	case 0:
		return client.createFolder(folder.str)
	case 1:
		client.mailboxes.add(mbox)
		return nil
	default:
		return fmt.Errorf("Found multiple folders matching '%s'.", folder)
	}
}

func (client *Client) EnsureFolder(folder Folder, errorHandler ErrorHandler) {
	client.commander.execute(ensureCommando{folder}, errorHandler)
}

func (client *Client) putMessages(folder Folder, messages []string) error {
	if len(messages) == 0 {
		return nil
	}

	now := time.Now()
	for _, msg := range messages {
		reader := strings.NewReader(msg)
		if err := client.c.Append(folder.str, nil, now, reader); err != nil {
			return fmt.Errorf("uploading message to %s: %w", folder, err)
		}
	}

	return nil
}

func (client *Client) PutMessages(folder Folder, messages []string, errorHandler ErrorHandler) {
	client.commander.execute(addCommando{folder, messages}, errorHandler)
}
