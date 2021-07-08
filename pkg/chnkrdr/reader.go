package chnkrdr

import (
	"io"
)

var _ io.Reader = &chunkedReader{}

type chunkedReader struct {
	r    io.Reader
	size int
}

func NewChunkedReader(r io.Reader, maxReadSize int) io.Reader {
	return &chunkedReader{r: r, size: maxReadSize}
}

func (r *chunkedReader) Read(p []byte) (n int, err error) {
	if len(p) > r.size {
		return r.r.Read(p[:r.size])
	}
	return r.r.Read(p)
}
