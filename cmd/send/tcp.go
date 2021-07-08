package main

import (
	"context"
	"log"
	"net"
	"sync"
)

func dialTCP(
	wg *sync.WaitGroup,
) (
	connType string,
	networkConnection net.Conn,
	cancel context.CancelFunc,
	listenCloser context.CancelFunc,
	err error,
) {
	connType = "TCP"
	tcpAddr := net.JoinHostPort(
		serverHost,
		tcpPort,
	)
	log.Printf(
		"[DEBUG] Connecting via %s to %s",
		connType,
		tcpAddr,
	)
	networkConnection, err = net.Dial(
		"tcp",
		tcpAddr,
	)
	return
}
