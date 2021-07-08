// create a bigfile on disk
package main

import (
	"crypto/rand"
	"github.com/k42-software/go-stream-speed/pkg/counter"
	"io"
	"log"
	"os"
)

func main() {

	log.SetFlags(log.Ltime)

	const outputFile = "bigfile.data"

	const KB = 1024
	const MB = 1024 * KB
	const GB = 1024 * MB

	reader := io.LimitReader(rand.Reader, 2*GB)

	fh, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("[ERROR] %s: %s", outputFile, err)
	}

	writeCounter := counter.NewCounter("File write")
	_, err = io.Copy(io.MultiWriter(fh, writeCounter), reader)
	if err != nil {
		log.Fatalf("[ERROR] %s", err)
	}
	if err = fh.Close(); err != nil {
		log.Fatalf("[ERROR] %s", err)
	}
	if err = writeCounter.Close(); err != nil {
		log.Fatalf("[ERROR] %s", err)
	}

	log.Printf("[DEBUG] Created %s", outputFile)

}
