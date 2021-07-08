package main

import (
	"context"
	"crypto/tls"
	"github.com/k42-software/go-stream-speed/pkg/discard"
	"github.com/k42-software/go-stream-speed/pkg/tlsconf"
	"log"
	"net"
	"sync"
)

func listenTls(wg *sync.WaitGroup) {
	defer wg.Done()

	tlsListen := net.JoinHostPort(serverHost, tlsPort)
	log.Printf("[DEBUG] TLS listening on %s", tlsListen)

	listener, err := tls.Listen(
		"tcp",
		tlsListen,
		tlsconf.ServerConfig(),
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
