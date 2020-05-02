package msg

import (
	"fmt"

	"github.com/Necoro/feed2imap-go/internal/imap"
	"github.com/Necoro/feed2imap-go/pkg/log"
)

// headers
const (
	VersionHeader = "X-Feed2Imap-Version"
	ReasonHeader  = "X-Feed2Imap-Reason"
	IdHeader      = "X-Feed2Imap-Item"
)

type Messages []Message

type Message struct {
	Content  string
	IsUpdate bool
	ID       string
}

func (m Messages) Upload(client *imap.Client, folder imap.Folder, reupload bool) error {
	toStore := make([]string, 0, len(m))

	msgs := make(chan Message, 5)
	ok := make(chan bool)
	go func() {
		errHappened := false
		for msg := range msgs {
			if err := client.Replace(folder, IdHeader, msg.ID, msg.Content, reupload); err != nil {
				log.Errorf("Error while updating mail with id '%s' in folder '%s'. Skipping.: %s",
					msg.ID, folder, err)
				errHappened = true
			}
		}

		ok <- errHappened
	}()

	for _, msg := range m {
		if !msg.IsUpdate {
			toStore = append(toStore, msg.Content)
		} else {
			msgs <- msg
		}
	}

	close(msgs)

	putErr := client.PutMessages(folder, toStore)
	updOk := <-ok

	if putErr != nil {
		return putErr
	}
	if updOk {
		return fmt.Errorf("Errors during updating mails.")
	}

	return nil
}
