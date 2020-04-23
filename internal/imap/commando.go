package imap

const maxPipeDepth = 10

type commander struct {
	client *Client
	pipe   chan<- execution
	done   chan<- struct{}
}

type command interface {
	execute(*connection) error
}

type execution struct {
	cmd          command
	done         chan<- error
}

func (commander *commander) execute(command command) error {
	done := make(chan error)
	commander.pipe <- execution{command, done}
	return <-done
}

func executioner(conn *connection, pipe <-chan execution, done <-chan struct{}) {
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
			err := execution.cmd.execute(conn)
			execution.done <- err
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

	for _, conn := range client.connections {
		if conn != nil {
			go executioner(conn, pipe, done)
		}
	}
}

func (client *Client) stopCommander() {
	if client.commander == nil {
		return
	}

	close(client.commander.done)

	client.commander = nil
}
