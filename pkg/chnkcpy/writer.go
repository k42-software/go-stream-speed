package chnkcpy

import "io"

// Express an anonymous function as an io.Writer
type WriterFunc func(p []byte) (n int, err error)

var _ io.Writer = WriterFunc(nil)

func (f WriterFunc) Write(p []byte) (n int, err error) {
	return f(p)
}
