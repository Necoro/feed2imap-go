package imap

import (
	"sync"

	"github.com/emersion/go-imap"
)

type mailboxes struct {
	mb          map[string]*imap.MailboxInfo
	mu          sync.RWMutex
	changeLocks map[string]chan struct{}
}

func (mbs *mailboxes) unlocking(elem Folder) {
	mbs.mu.Lock()
	defer mbs.mu.Unlock()

	ch, ok := mbs.changeLocks[elem.str]
	if !ok {
		panic("Unlocking where nothing is locked")
	}
	close(ch)
	delete(mbs.changeLocks, elem.str)
}

func (mbs *mailboxes) locking(elem Folder) bool {
	mbs.mu.Lock()
	ch, ok := mbs.changeLocks[elem.str]
	if !ok {
		ch = make(chan struct{})
		mbs.changeLocks[elem.str] = ch
		mbs.mu.Unlock()
		// we created the lock, we are in charge and done here
		return false
	} else {
		// someone else is working, we wait till he's done
		mbs.mu.Unlock() // we are not doing anything...
		<-ch
		return true
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

func NewMailboxes() *mailboxes {
	return &mailboxes{
		mb:          map[string]*imap.MailboxInfo{},
		changeLocks: map[string]chan struct{}{},
		mu:          sync.RWMutex{},
	}
}
