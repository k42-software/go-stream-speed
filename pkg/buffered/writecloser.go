// Coped from https://github.com/lucas-clemente/quic-go
// MIT License; https://github.com/lucas-clemente/quic-go/blob/master/LICENSE

// An instance of bufio.Writer which also implements io.Closer so that you
// can be assured that the buffer is flushed on close.
package buffered

import (
	"bufio"
	"io"
)

type bufferedWriteCloser struct {
	*bufio.Writer
	io.Closer
}

// NewWriteCloser creates an io.WriteCloser from a bufio.Writer and an io.Closer
func NewWriteCloser(writer *bufio.Writer, closer io.Closer) io.WriteCloser {
	return &bufferedWriteCloser{
		Writer: writer,
		Closer: closer,
	}
}

func (h bufferedWriteCloser) Close() error {
	if err := h.Writer.Flush(); err != nil {
		return err
	}
	return h.Closer.Close()
}
