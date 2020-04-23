package imap

import (
	"sync"

	"github.com/emersion/go-imap"
)

type mailboxes struct {
	mb map[string]*imap.MailboxInfo
	mu sync.RWMutex
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

func NewMailboxes() *mailboxes {
	return &mailboxes{
		mb: map[string]*imap.MailboxInfo{},
		mu: sync.RWMutex{},
	}
}