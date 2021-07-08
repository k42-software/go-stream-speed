package fakenet

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"github.com/dustin/go-humanize"
	"hash/crc32"
	"io"
	"io/ioutil"
	"sync"
	"testing"
	"time"
)

func TestNewConn(t *testing.T) {

	const localhost = "127.0.0.1"

	localAddr := NewAddr(localhost, "1234")
	remoteAddr := NewAddr(localhost, "1235")

	t.Logf("fake stream conn: %s <=> %s", localAddr, remoteAddr)

	localConn, remoteConn, err := NewConn(
		context.Background(),
		localAddr,
		remoteAddr,
		MaximumMTU,
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
		t.Logf("local writer started")
		reader := io.TeeReader(bytes.NewReader(fileContents), localChecksum)
		if n, err := io.Copy(localConn, reader); err == nil {
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
		if err = localConn.Close(); err != nil {
			t.Errorf("local close: %s", err)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		started := time.Now()
		t.Logf("remote reader started")
		if n, err := io.Copy(remoteChecksum, remoteConn); err == nil {
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
		if err = remoteConn.Close(); err != nil {
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
