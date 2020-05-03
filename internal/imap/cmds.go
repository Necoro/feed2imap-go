package imap

type ensureCommando struct {
	folder Folder
}

func (cmd ensureCommando) execute(conn *connection) error {
	return conn.ensureFolder(cmd.folder)
}

func (cl *Client) EnsureFolder(folder Folder) error {
	return cl.commander.execute(ensureCommando{folder})
}

type addCommando struct {
	folder   Folder
	messages []string
}

func (cmd addCommando) execute(conn *connection) error {
	return conn.putMessages(cmd.folder, cmd.messages)
}

func (cl *Client) PutMessages(folder Folder, messages []string) error {
	return cl.commander.execute(addCommando{folder, messages})
}

type replaceCommando struct {
	folder     Folder
	header     string
	value      string
	newContent string
	force      bool
}

func (cmd replaceCommando) execute(conn *connection) error {
	return conn.replace(cmd.folder, cmd.header, cmd.value, cmd.newContent, cmd.force)
}

func (cl *Client) Replace(folder Folder, header, value, newContent string, force bool) error {
	return cl.commander.execute(replaceCommando{folder, header, value, newContent, force})
}
