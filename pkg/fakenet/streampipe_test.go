package fakenet

import (
	"bytes"
	"crypto/rand"
	"github.com/dustin/go-humanize"
	"github.com/k42-software/go-stream-speed/pkg/chnkcpy"
	"io"
	"testing"
	"time"
)

func TestPipeLongShort(t *testing.T) {

	pipe, err := NewStreamPipe(MaximumMTU)
	if err != nil {
		t.Errorf("new pipe: %s", err)
	}

	// long write
	started := time.Now()
	t.Logf("filling pipe with a long write")
	writeBuffer := make([]byte, MaximumMTU*1024)
	_, err = io.ReadFull(rand.Reader, writeBuffer)
	if err != nil {
		t.Errorf("read random: %s", err)
	}
	n, err := pipe.Write(writeBuffer)
	if err != nil {
		t.Errorf("pipe write: %s", err)
	}
	if n < len(writeBuffer) {
		t.Errorf("pipe write: %s", io.ErrShortWrite)
	}
	duration := time.Since(started)
	t.Logf(
		"wrote %s in %s (%s/s)",
		humanize.IBytes(uint64(n)),
		duration,
		humanize.IBytes(uint64(float64(n)/duration.Seconds())),
	)

	t.Logf("closing pipe for writing")
	err = pipe.Close()
	if err != nil {
		t.Errorf("pipe close: %s", err)
	}

	// short read
	started = time.Now()
	t.Logf("draining pipe with short reads")
	targetBuffer := bytes.NewBuffer(nil)
	n, err = chnkcpy.ChunkedCopy(targetBuffer, pipe, MinimumMTU/2)
	if err != nil {
		t.Errorf("pipe close: %s", err)
	}
	duration = time.Since(started)
	t.Logf(
		"read %s in %s (%s/s)",
		humanize.IBytes(uint64(n)),
		duration,
		humanize.IBytes(uint64(float64(n)/duration.Seconds())),
	)

	t.Logf("comparing buffers for equality")
	if bytes.Compare(writeBuffer, targetBuffer.Bytes()) != 0 {
		t.Error("original and target buffers do not match")
	}
}

func TestPipeShortLong(t *testing.T) {

	pipe, err := NewStreamPipe(MaximumMTU)
	if err != nil {
		t.Errorf("new pipe: %s", err)
	}

	// long write
	started := time.Now()
	t.Logf("filling pipe with a short writes")
	writeBuffer := make([]byte, MaximumMTU*1024)
	_, err = io.ReadFull(rand.Reader, writeBuffer)
	if err != nil {
		t.Errorf("read random: %s", err)
	}
	n, err := chnkcpy.ChunkedCopy(pipe, bytes.NewReader(writeBuffer), MinimumMTU/2)
	if err != nil {
		t.Errorf("pipe close: %s", err)
	}
	if n < len(writeBuffer) {
		t.Errorf("pipe write: %s", io.ErrShortWrite)
	}
	duration := time.Since(started)
	t.Logf(
		"wrote %s in %s (%s/s)",
		humanize.IBytes(uint64(n)),
		duration,
		humanize.IBytes(uint64(float64(n)/duration.Seconds())),
	)

	t.Logf("closing pipe for writing")
	err = pipe.Close()
	if err != nil {
		t.Errorf("pipe close: %s", err)
	}

	// short read
	started = time.Now()
	t.Logf("draining pipe with a long read")
	targetBuffer := bytes.NewBuffer(nil)
	var n64 int64
	n64, err = io.CopyBuffer(targetBuffer, pipe, make([]byte, len(writeBuffer)*2))
	if err != nil {
		t.Errorf("pipe close: %s", err)
	}
	duration = time.Since(started)
	t.Logf(
		"wrote %s in %s (%s/s)",
		humanize.IBytes(uint64(n64)),
		duration,
		humanize.IBytes(uint64(float64(n64)/duration.Seconds())),
	)

	t.Logf("comparing buffers for equality")
	if bytes.Compare(writeBuffer, targetBuffer.Bytes()) != 0 {
		t.Error("original and target buffers do not match")
	}
}
