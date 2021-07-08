// send a stream and measure the transfer speed
package main

import (
	"log"
	"sync"
	"time"
)

const echoMode = false

const serverHost = "127.0.0.1"

const kcpPort = "7951"
const tcpPort = "7961"
const quicPort = "7971"
const tlsPort = "7991"

// For mock test networks
const clientPort = "1234"

func main() {

	log.SetFlags(log.Ltime)

	dialer, shutdown := processCliFlags()

	//---------------------------------------------------------------------------------------------
	// Dial

	wg := &sync.WaitGroup{}

	connType, networkConnection, ctxCancel, listenCloser, err := dialer(wg)
	if ctxCancel != nil {
		defer ctxCancel()
	}

	if networkConnection == nil {
		log.Fatal("[ERROR] failed to create network connection object")
	}

	//---------------------------------------------------------------------------------------------
	// Interact

	var (
		totalWritten uint64
		writingDone  uint64
	)

	// writer
	wg.Add(1)
	go mainWriter(
		wg,
		connType,
		networkConnection,
		&totalWritten,
		&writingDone,
		listenCloser,
	)

	// reader
	if echoMode {
		wg.Add(1)
		go mainReader(
			wg,
			connType,
			networkConnection,
			&totalWritten,
			&writingDone,
		)
	}

	//---------------------------------------------------------------------------------------------
	// Shutdown

	wg.Wait()

	if err = networkConnection.Close(); err != nil {
		log.Printf("[ERROR] %s connection close error: %s", connType, err)
	}

	log.Printf("[DEBUG] %s connection closed", connType)

	if shutdown != nil {
		shutdown()
	}

	log.Printf("[DEBUG] Sleeping 2 seconds")
	time.Sleep(2 * time.Second)
	log.Printf("[DEBUG] Exiting")
}
