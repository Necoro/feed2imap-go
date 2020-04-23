package imap

import "strings"

type Folder struct {
	str       string
	delimiter string
}

func (f Folder) String() string {
	return f.str
}

func (f Folder) Append(other Folder) Folder {
	if f.delimiter != other.delimiter {
		panic("Delimiters do not match")
	}
	return Folder{
		str:       f.str + f.delimiter + other.str,
		delimiter: f.delimiter,
	}
}

func (client *Client) folderName(path []string) Folder {
	return Folder{
		strings.Join(path, client.delimiter),
		client.delimiter,
	}
}

func (client *Client) NewFolder(path []string) Folder {
	return client.toplevel.Append(client.folderName(path))
}
