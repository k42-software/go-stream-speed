package main

import (
	"context"
	"crypto/tls"
	"github.com/k42-software/go-stream-speed/pkg/tlsconf"
	"log"
	"net"
	"sync"
)

func dialTLS(
	wg *sync.WaitGroup,
) (
	connType string,
	networkConnection net.Conn,
	cancel context.CancelFunc,
	listenCloser context.CancelFunc,
	err error,
) {
	connType = "TLS"
	var tcpAddr = net.JoinHostPort(
		serverHost,
		tlsPort,
	)
	log.Printf(
		"[DEBUG] Connecting via %s to %s",
		connType,
		tcpAddr,
	)
	networkConnection, err = tls.Dial(
		"tcp",
		tcpAddr,
		tlsconf.ClientConfig(),
	)
	return
}
