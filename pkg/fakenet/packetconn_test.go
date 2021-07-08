package fakenet

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"github.com/dustin/go-humanize"
	"github.com/k42-software/go-stream-speed/pkg/chnkcpy"
	"hash/crc32"
	"io"
	"io/ioutil"
	"sync"
	"testing"
	"time"
)

func TestNewPacketConn(t *testing.T) {

	const localhost = "127.0.0.1"

	const packetSize = MaximumMTU

	localAddr := NewAddr(localhost, "1234")
	remoteAddr := NewAddr(localhost, "1235")

	t.Logf("fake packet conn: %s <=> %s", localAddr, remoteAddr)

	localPacketConn, remotePacketConn, err := NewPacketConn(
		context.Background(),
		localAddr,
		remoteAddr,
		packetSize,
	)
	if err != nil {
		t.Errorf("new connection: %s", err)
	}

	fileContents, err := ioutil.ReadFile("../../bigfile.data")
	if err != nil {
		t.Errorf("file read: %s", err)
	}
	t.Logf("file loaded into memory")

	var wg sync.WaitGroup

	localChecksum := crc32.NewIEEE()
	remoteChecksum := crc32.NewIEEE()

	wg.Add(1)
	go func() {
		started := time.Now()
		t.Logf("local packet writer started")
		reader := io.TeeReader(bytes.NewReader(fileContents), localChecksum)
		n, err := chnkcpy.ChunkedCopy(
			chnkcpy.WriterFunc(func(p []byte) (n int, err error) {
				return localPacketConn.WriteTo(p, remoteAddr)
			}),
			reader,
			packetSize,
		)
		if err == nil {
			duration := time.Since(started)
			t.Logf(
				"local writer wrote %s in %s (%s/s)",
				humanize.IBytes(uint64(n)),
				duration,
				humanize.IBytes(uint64(float64(n)/duration.Seconds())),
			)
		} else {
			t.Errorf("local write: %s", err)
		}
		t.Logf("local writer completed")
		if err = localPacketConn.Close(); err != nil {
			t.Errorf("local close: %s", err)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		started := time.Now()
		t.Logf("remote packet reader started")
		n, err := io.Copy(
			remoteChecksum,
			chnkcpy.ReaderFunc(func(p []byte) (n int, err error) {
				n, _, err = remotePacketConn.ReadFrom(p)
				return n, err
			}),
		)
		if err == nil {
			duration := time.Since(started)
			t.Logf(
				"remote reader read %s in %s (%s/s)",
				humanize.IBytes(uint64(n)),
				duration,
				humanize.IBytes(uint64(float64(n)/duration.Seconds())),
			)
		} else {
			t.Errorf("remote read: %s", err)
		}
		t.Logf("remote reader completed")
		if err = remotePacketConn.Close(); err != nil {
			t.Errorf("remote close: %s", err)
		}
		wg.Done()
	}()

	wg.Wait()

	uint32hex := func(number uint32) string {
		buffer := make([]byte, 4)
		binary.LittleEndian.PutUint32(buffer, number)
		return hex.EncodeToString(buffer)
	}

	localSum := localChecksum.Sum32()
	remoteSum := remoteChecksum.Sum32()

	t.Logf(" local checksum: %s", uint32hex(localSum))
	t.Logf("remote checksum: %s", uint32hex(remoteSum))

	if localSum != remoteSum {
		t.Error("local and remote checksums do not match")
	}

	t.Logf("test completed")
}
