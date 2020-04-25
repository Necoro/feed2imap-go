package imap

import (
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	imapClient "github.com/emersion/go-imap/client"

	"github.com/Necoro/feed2imap-go/pkg/log"
	"github.com/Necoro/feed2imap-go/pkg/util"
)

type connection struct {
	*connConf
	mailboxes *mailboxes
	c         *imapClient.Client
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

func (conn *connection) disconnect() bool {
	if conn != nil {
		connected := (conn.c.State() & imap.ConnectedState) != 0
		_ = conn.c.Logout()
		return connected
	}
	return false
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

	if conn.mailboxes.locking(folder) {
		// someone else tried to create the MB -- try again, now that he's done
		return conn.ensureFolder(folder)
	} else {
		defer conn.mailboxes.unlocking(folder)
	}

	log.Printf("Checking for folder '%s'", folder)

	mbox, found, err := conn.list(folder.str)
	if err != nil {
		return err
	}

	if mbox != nil && mbox.Delimiter != folder.delimiter {
		panic("Delimiters do not match")
	}

	switch {
	case found == 0 || (found == 1 && util.StrContains(mbox.Attributes, imap.NoSelectAttr)):
		return conn.createFolder(folder.str)
	case found == 1:
		conn.mailboxes.add(mbox)
		return nil
	default:
		return fmt.Errorf("Found multiple folders matching '%s'.", folder)
	}
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
