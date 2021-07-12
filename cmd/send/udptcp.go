package main

import (
	"context"
	"github.com/k42-software/go-stream-speed/pkg/udptcp"
	"log"
	"net"
	"sync"
)

func dialUDPTCP(
	wg *sync.WaitGroup,
) (
	connType string,
	networkConnection net.Conn,
	cancel context.CancelFunc,
	listenCloser context.CancelFunc,
	err error,
) {
	connType = "UDPTCP"

	localAddr, _ := udptcp.ParseAddress(
		net.JoinHostPort(serverHost, "0"),
	)
	remoteAddr, _ := udptcp.ParseAddress(
		net.JoinHostPort(serverHost, udpTcpRemotePort),
	)

	log.Printf(
		"[DEBUG] UDPTCP Dialing from %s to %s",
		localAddr,
		remoteAddr,
	)

	networkConnection, err = udptcp.Dial(
		localAddr.UDP(),
		remoteAddr.UDP(),
	)

	return
}
