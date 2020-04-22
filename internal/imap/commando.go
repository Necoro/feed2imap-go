package imap

import "github.com/Necoro/feed2imap-go/internal/log"

const maxPipeDepth = 10

type commander struct {
	client *Client
	pipe   chan<- execution
	done   chan<- struct{}
}

type command interface {
	execute(client *Client) error
}

type ErrorHandler func(error) string

type execution struct {
	cmd          command
	done         chan<- struct{}
	errorHandler ErrorHandler
}

type addCommando struct {
	folder   Folder
	messages []string
}

func (cmd addCommando) execute(client *Client) error {
	return client.putMessages(cmd.folder, cmd.messages)
}

type ensureCommando struct {
	folder Folder
}

func (cmd ensureCommando) execute(client *Client) error {
	return client.ensureFolder(cmd.folder)
}

func (commander *commander) execute(command command, handler ErrorHandler) {
	done := make(chan struct{})
	commander.pipe <- execution{command, done, handler}
	<-done
}

func executioner(client *Client, pipe <-chan execution, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case execution := <-pipe:
			select { // break as soon as done is there
			case <-done:
				return
			default:
			}
			if err := execution.cmd.execute(client); err != nil {
				if execution.errorHandler == nil {
					log.Error(err)
				} else {
					log.Error(execution.errorHandler(err))
				}
			}
			close(execution.done)
		}
	}
}

func (client *Client) startCommander() {
	if client.commander != nil {
		return
	}

	pipe := make(chan execution, maxPipeDepth)
	done := make(chan struct{})

	client.commander = &commander{client, pipe, done}

	go executioner(client, pipe, done)
}

func (client *Client) stopCommander() {
	if client.commander == nil {
		return
	}

	close(client.commander.done)

	client.commander = nil
}
