package imap

import (
	"fmt"
	"strings"

	"github.com/emersion/go-imap"
	imapClient "github.com/emersion/go-imap/client"

	"github.com/Necoro/feed2imap-go/internal/log"
)

type Client struct {
	c         *imapClient.Client
	host      string
	folders   folders
	delimiter string
	toplevel  string
}

type folders map[string]*imap.MailboxInfo

func (f folders) contains(elem string) bool {
	_, ok := f[elem]
	return ok
}

func (f folders) add(elem *imap.MailboxInfo) {
	name := elem.Name
	f[name] = elem
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
