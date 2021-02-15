package util

import "io"

type fixWriter struct {
	w io.Writer
}

var lf = []byte{'\n'}
var cr = []byte{'\r'}

func (f fixWriter) Write(p []byte) (n int, err error) {
	crFound := false
	start := 0

	write := func(str []byte) {
		var j int
		j, err = f.w.Write(str)
		n = n + j
	}

	for idx, b := range p {
		if crFound && b != '\n' {
			// insert '\n'
			if write(p[start:idx]); err != nil {
				return
			}
			if write(lf); err != nil {
				return
			}

			start = idx
		} else if !crFound && b == '\n' {
			// insert '\r'
			if write(p[start:idx]); err != nil {
				return
			}
			if write(cr); err != nil {
				return
			}

			start = idx
		}
		crFound = b == '\r'
	}

	// write the remainder
	if write(p[start:]); err != nil {
		return
	}

	if crFound { // dangling \r
		write(lf)
	}

	return
}

// Cyrus IMAP really cares about single \r and \n.
// Implement this fixer to change them into \r\n.
func FixWriter(w io.Writer) io.Writer {
	return &fixWriter{w}
}
