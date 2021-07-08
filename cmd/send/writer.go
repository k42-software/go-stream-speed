package main

import (
	"bufio"
	"bytes"
	"context"
	"github.com/dustin/go-humanize"
	"github.com/k42-software/go-stream-speed/pkg/chnkrdr"
	"github.com/k42-software/go-stream-speed/pkg/counter"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const writerBufferSize = 4096

const fileReadingChunkSize = writerBufferSize

func mainWriter(
	wg *sync.WaitGroup,
	connType string,
	networkConnection net.Conn,
	totalWritten *uint64,
	writingDone *uint64,
	listenCloser context.CancelFunc,
) {
	started := time.Now()
	defer func() {
		log.Printf("[DEBUG] %s stopped writing", connType)
		written := atomic.LoadUint64(totalWritten)
		seconds := time.Since(started).Seconds()
		bytesPerSecond := uint64(float64(written) / seconds)
		log.Printf(
			"[DEBUG] %s written %s in %s (%s/s)",
			connType,
			humanize.IBytes(written),
			time.Since(started),
			humanize.IBytes(bytesPerSecond),
		)
		atomic.StoreUint64(writingDone, 1)
		wg.Done()

		if listenCloser != nil {
			listenCloser()
		}
	}()

	var fileReader io.Reader

	if fh, err := os.OpenFile(*testFilePath, os.O_RDONLY, 0); err == nil {

		// Load all the file in to memory here so that the disk/file I/O
		// doesn't impact the network writing speed. Additionally report
		// how quickly we read the file, which should help us understand
		// what our upper performance limit probably is (assuming that
		// the disk is faster than the network).
		readCounter := counter.NewCounter("   File read")
		reader := io.TeeReader(fh, readCounter)
		writer := bytes.NewBuffer(nil)
		if _, err = io.Copy(writer, reader); err != nil {
			log.Fatalf("[ERROR] File read: %s", err)
		}
		if err = fh.Close(); err != nil {
			log.Fatalf("[ERROR] File close: %s", err)
		}
		_ = readCounter.Close()

		//fileReader = bytes.NewReader(writer.Bytes())
		//fileReader = writer

		// Eliminate the io.ReadFrom and io.WriteTo optimisations and
		// force all reads from the file to be of our specific chunk
		// size; so that we have a consistent read performance in
		// all of the testing modes.
		fileReader = chnkrdr.NewChunkedReader(writer, fileReadingChunkSize)

		log.Printf("[DEBUG] File contents buffered in memory")

	} else {
		log.Fatalf("[ERROR] %s write: %s", connType, err)
	}

	var writer io.Writer
	var reader io.Reader

	reader = fileReader

	//readCounter := counter.NewCounter("   Copy read")
	//defer readCounter.Close()
	//reader = io.TeeReader(reader, readCounter)

	//reader = NewSlowReader(reader, 1024 * 1024)

	writer = networkConnection
	writer = bufio.NewWriterSize(writer, writerBufferSize)

	writeCounter := counter.NewCounter("Send: Write")
	defer writeCounter.Close()
	writer = io.MultiWriter(writeCounter, writer)

	// Update started so we only report the transfer time
	started = time.Now()

	if n, err := io.Copy(writer, reader); err == nil {
		atomic.StoreUint64(totalWritten, uint64(n))
	} else {
		log.Printf("[ERROR] %s write: %s", connType, err)
	}

	if flusher, ok := writer.(interface{ Flush() error }); ok {
		log.Printf("[DEBUG] Stream writer flush")
		if err := flusher.Flush(); err != nil {
			log.Printf("[ERROR] %s write: %s", connType, err)
		}
	}

}
