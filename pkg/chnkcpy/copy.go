package chnkcpy

import (
	"io"
)

// This copies to dst from src while limiting each read an write to a
// maximum of the given chunk size. This prevents some of the more
// efficient copying options in order to avoid the use of larger buffers.
func ChunkedCopy(dst io.Writer, src io.Reader, chunkSize int) (n int, err error) {
	// This doesn't use the chunkedReader for performance reasons. 
	var written int64
	written, err = io.CopyBuffer(
		WriterFunc(func(p []byte) (n int, err error) {
			return dst.Write(p)
		}),
		ReaderFunc(func(p []byte) (n int, err error) {
			return src.Read(p)
		}),
		make([]byte, chunkSize),
	)
	n = int(written)
	return n, err
}
