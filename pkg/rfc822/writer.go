// Package rfc822 provides a writer that ensures the intrinsics of RFC 822.
//
// Rationale
//
// Cyrus IMAP really cares about the hard specifics of RFC 822, namely not allowing single \r and \n.
//
// See also: https://www.cyrusimap.org/imap/reference/faqs/interop-barenewlines.html
// and: https://github.com/Necoro/feed2imap-go/issues/46
//
// NB: This package currently only cares about the newlines.
package rfc822

import "io"

type rfc822Writer struct {
	w io.Writer
}

var lf = []byte{'\n'}
var cr = []byte{'\r'}

func (f rfc822Writer) Write(p []byte) (n int, err error) {
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

// Writer creates a new RFC 822 conform writer.
func Writer(w io.Writer) io.Writer {
	return rfc822Writer{w}
}
