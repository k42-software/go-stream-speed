package main

import (
	"context"
	"github.com/k42-software/go-stream-speed/pkg/discard"
	"log"
	"net"
	"sync"
)

func listenTcp(wg *sync.WaitGroup) {
	defer wg.Done()

	tcpListen := net.JoinHostPort(serverHost, tcpPort)
	log.Printf("[DEBUG] TCP listening on %s", tcpListen)

	listener, err := net.Listen(
		"tcp",
		tcpListen,
	)
	if err != nil {
		log.Fatalf("[ERROR] %s", err)
	}

	for {
		networkConnection, err := listener.Accept()
		if err != nil {
			log.Printf("[ERROR] %s", err)
			continue
		}
		go discard.StreamConnection(
			context.Background(),
			networkConnection,
		)
	}
}
