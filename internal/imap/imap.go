package imap

import (
	"fmt"

	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
)

func Connect(url config.Url, maxConnections int) (*Client, error) {
	var err error

	client := newClient(maxConnections)
	client.host = url.Host
	defer func() {
		if err != nil {
			client.Disconnect()
		}
	}()
	client.startCommander()

	var conn *connection // the main connection
	if conn, err = client.connect(url); err != nil {
		return nil, err
	}

	delim, err := conn.fetchDelimiter()
	if err != nil {
		return nil, fmt.Errorf("fetching delimiter: %w", err)
	}
	client.delimiter = delim

	client.toplevel = client.folderName(url.RootPath())

	log.Printf("Determined '%s' as toplevel, with '%s' as delimiter", client.toplevel, client.delimiter)

	if !client.toplevel.IsBlank() {
		if err = conn.ensureFolder(client.toplevel); err != nil {
			return nil, err
		}
	}

	// the other connections
	for i := 1; i < client.maxConnections; i++ {
		go func(id int) {
			if _, err := client.connect(url); err != nil { // explicitly new var 'err', b/c these are now harmless
				log.Warnf("connecting #%d: %s", id, err)
			}
		}(i)
	}

	return client, nil
}
