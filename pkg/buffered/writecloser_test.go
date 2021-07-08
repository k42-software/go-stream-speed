// Coped from github.com/lucas-clemente/quic-go@v0.21.1/internal/utils
// MIT License

package buffered

import (
	"bufio"
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

var _ = Describe("buffered io.WriteCloser", func() {
	It("flushes before closing", func() {
		buf := &bytes.Buffer{}

		w := bufio.NewWriter(buf)
		wc := NewWriteCloser(w, &nopCloser{})
		wc.Write([]byte("foobar"))
		Expect(buf.Len()).To(BeZero())
		Expect(wc.Close()).To(Succeed())
		Expect(buf.String()).To(Equal("foobar"))
	})
})
