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

	var prefix string
	if f.str == "" {
		prefix = ""
	} else {
		prefix = f.str + f.delimiter
	}

	return Folder{
		str:       prefix + other.str,
		delimiter: f.delimiter,
	}
}

func buildFolderName(path []string, delimiter string) (name string) {
	name = strings.Join(path, delimiter)
	if delimiter != "" {
		name = strings.Trim(name, delimiter[0:1])
	}
	return
}

func (cl *Client) folderName(path []string) Folder {
	return Folder{
		buildFolderName(path, cl.delimiter),
		cl.delimiter,
	}
}

func (cl *Client) NewFolder(path []string) Folder {
	return cl.toplevel.Append(cl.folderName(path))
}
