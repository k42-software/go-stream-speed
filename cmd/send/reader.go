package main

import (
	"bufio"
	"github.com/dustin/go-humanize"
	"io"
	"log"
	"github.com/k42-software/go-stream-speed/pkg/counter"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const readerBufferSize = 4096

func mainReader(
	wg *sync.WaitGroup,
	connType string,
	networkConnection net.Conn,
	totalWritten *uint64,
	writingDone *uint64,
) {
	defer wg.Done()

	started := time.Now()

	var totalRead uint64
	var expectedSize uint64

	//readCounter := counter.NewCounter("Stream read")
	//defer readCounter.Close()
	//reader := io.TeeReader(bufio.NewReaderSize(networkConnection, readerBufferSize), readCounter)
	reader := bufio.NewReaderSize(networkConnection, readerBufferSize)

	writeCounter := counter.NewCounter(" Stream read")
	defer writeCounter.Close()
	writer := writeCounter

	// This is a modified copy of the io.Copy logic.
	// This was modified to check if it has read enough yet.
	var err error
	buf := make([]byte, 32*1024)
	for {
		if atomic.LoadUint64(writingDone) == 1 {
			if expectedSize == 0 {
				expectedSize = atomic.LoadUint64(totalWritten)
				log.Printf("[DEBUG] Discovered expected size is %d", expectedSize)
			}
			if totalRead == expectedSize {
				log.Printf("[DEBUG] %s read exactly expected size", connType)
				break
			}
		}
		nr, er := reader.Read(buf)
		if nr > 0 {
			nw, ew := writer.Write(buf[0:nr])
			if nw > 0 {
				totalRead += uint64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	if err != nil {
		log.Printf("[ERROR] Stream read: %s", err)
	}

	log.Printf("[DEBUG] %s stopped reading", connType)
	seconds := time.Since(started).Seconds()
	bytesPerSecond := uint64(float64(totalRead) / seconds)
	log.Printf(
		"[DEBUG] %s read %s in %s (%s/s)",
		connType,
		humanize.IBytes(totalRead),
		time.Since(started),
		humanize.IBytes(bytesPerSecond),
	)
}
