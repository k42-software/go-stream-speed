package fakenet

import (
	"crypto/rand"
	"io"
	"testing"
)

func TestPacketPipeLongWrite(t *testing.T) {

	pipe, err := NewPacketPipe(StandardMTU)
	if err != nil {
		t.Errorf("new pipe: %s", err)
	}

	// long write
	t.Logf("filling pipe with a long write")
	writeBuffer := make([]byte, MaximumMTU*1024)
	_, err = io.ReadFull(rand.Reader, writeBuffer)
	if err != nil {
		t.Errorf("read random: %s", err)
	}
	n, err := pipe.Write(writeBuffer)
	if err == io.ErrShortBuffer {
		t.Logf("received expected error")
	} else {
		t.Errorf("unexpected error: pipe write: %s", err)
	}
	if n != 0 {
		t.Errorf("unexpected bytes written: %d != %d", n, 0)
	}

	t.Logf("closing pipe for writing")
	err = pipe.Close()
	if err != nil {
		t.Errorf("pipe close: %s", err)
	}

}
