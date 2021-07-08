package chnkcpy

import "io"

// Express an anonymous function as an io.Reader
type ReaderFunc func(p []byte) (n int, err error)

var _ io.Reader = ReaderFunc(nil)

func (f ReaderFunc) Read(p []byte) (n int, err error) {
	return f(p)
}
