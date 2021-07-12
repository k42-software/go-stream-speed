package main

import (
	"context"
	"github.com/k42-software/go-stream-speed/pkg/discard"
	"github.com/k42-software/go-stream-speed/pkg/udptcp"
	"log"
	"net"
	"sync"
)

func listenUdpTcp(wg *sync.WaitGroup) {
	defer wg.Done()

	localAddr, _ := net.ResolveUDPAddr(
		"udp",
		net.JoinHostPort(serverHost, udpTcpPort),
	)

	listener, err := udptcp.Listen(localAddr)
	if err != nil {
		log.Fatalf("[ERROR] %s", err)
	}

	log.Printf(
		"[DEBUG] UDPTCP listening on %s",
		listener.Addr(),
	)

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
