package main

import (
	"context"
	"github.com/k42-software/go-stream-speed/pkg/discard"
	"github.com/k42-software/go-stream-speed/pkg/quicconf"
	"github.com/k42-software/go-stream-speed/pkg/quicnet"
	"log"
	"net"
	"sync"
)

func listenQuic(wg *sync.WaitGroup) {
	defer wg.Done()

	quicListen := net.JoinHostPort(serverHost, quicPort)
	log.Printf("[DEBUG] QUIC listening on %s", quicListen)

	listener, err := quicnet.ListenAddr(
		quicListen,
		quicconf.ServerTlsConfig(),
		quicconf.ServerConfig(false),
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
