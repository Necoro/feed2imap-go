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
	}

	defer conn.mailboxes.unlocking(folder)

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

func (conn *connection) delete(uids []uint32) error {
	storeItem := imap.FormatFlagsOp(imap.AddFlags, true)
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)

	if err := conn.c.UidStore(seqSet, storeItem, imap.DeletedFlag, nil); err != nil {
		return fmt.Errorf("marking as deleted: %w", err)
	}

	if err := conn.c.Expunge(nil); err != nil {
		return fmt.Errorf("expunging: %w", err)
	}

	return nil
}

func (conn *connection) fetchFlags(uid uint32) ([]string, error) {
	fetchItem := []imap.FetchItem{imap.FetchFlags}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uid)

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- conn.c.UidFetch(seqSet, fetchItem, messages)
	}()

	msg := <-messages
	err := <-done

	if err != nil {
		return nil, fmt.Errorf("fetching flags: %w", err)
	}
	return msg.Flags, nil
}

func (conn *connection) replace(folder Folder, header, value, newContent string, force bool) error {
	var err error
	var msgIds []uint32

	if err = conn.selectFolder(folder); err != nil {
		return err
	}

	if msgIds, err = conn.searchHeader(header, value); err != nil {
		return err
	}

	if len(msgIds) == 0 {
		if force {
			return conn.append(folder, nil, newContent)
		}
		return nil // nothing to do
	}

	var flags []string
	if flags, err = conn.fetchFlags(msgIds[0]); err != nil {
		return err
	}

	if err = conn.delete(msgIds); err != nil {
		return err
	}

	if err = conn.append(folder, flags, newContent); err != nil {
		return err
	}

	return nil
}

func (conn *connection) searchHeader(header, value string) ([]uint32, error) {
	criteria := imap.NewSearchCriteria()
	criteria.Header.Set(header, value)
	ids, err := conn.search(criteria)
	if err != nil {
		return nil, fmt.Errorf("searching for header %q=%q: %w", header, value, err)
	}
	return ids, nil
}

func (conn *connection) search(criteria *imap.SearchCriteria) ([]uint32, error) {
	return conn.c.UidSearch(criteria)
}

func (conn *connection) selectFolder(folder Folder) error {
	if _, err := conn.c.Select(folder.str, false); err != nil {
		return fmt.Errorf("selecting folder %s: %w", folder, err)
	}

	return nil
}

func (conn *connection) append(folder Folder, flags []string, msg string) error {
	reader := strings.NewReader(msg)
	if err := conn.c.Append(folder.str, flags, time.Now(), reader); err != nil {
		return fmt.Errorf("uploading message to %s: %w", folder, err)
	}

	return nil
}

func (conn *connection) putMessages(folder Folder, messages []string) error {
	if len(messages) == 0 {
		return nil
	}

	for _, msg := range messages {
		if err := conn.append(folder, nil, msg); err != nil {
			return err
		}
	}

	return nil
}
