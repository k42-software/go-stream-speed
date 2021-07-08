package main

import (
	"context"
	"crypto/tls"
	"github.com/k42-software/go-stream-speed/pkg/discard"
	"github.com/k42-software/go-stream-speed/pkg/fakenet"
	"github.com/k42-software/go-stream-speed/pkg/tlsconf"
	"log"
	"net"
	"sync"
)

func dialMockTLS(
	wg *sync.WaitGroup,
) (
	connType string,
	networkConnection net.Conn,
	cancel context.CancelFunc,
	listenCloser context.CancelFunc,
	err error,
) {

	connType = "TLS"

	ctx, ctxCancel := context.WithCancel(context.Background())
	cancel = ctxCancel

	//
	// Setup a fake virtual network
	//
	var localConn net.Conn
	var remoteConn net.Conn
	localConn, remoteConn, err = fakenet.NewConn(
		ctx,
		fakenet.NewAddr(serverHost, clientPort),
		fakenet.NewAddr(serverHost, tlsPort),
		fakenet.StandardMTU,
	)
	if err != nil {
		return
	}

	//
	// Setup an in-process listener
	//
	log.Printf(
		"[DEBUG] %s server receiving on %s",
		connType,
		remoteConn.LocalAddr(),
	)
	go discard.StreamConnection(
		ctx,
		tls.Server(
			remoteConn,
			tlsconf.ServerConfig(),
		),
	)

	//
	// Setup an in-process client
	//
	log.Printf(
		"[DEBUG] Connecting via %s to %s",
		connType,
		localConn.RemoteAddr(),
	)
	networkConnection = tls.Client(
		localConn,
		tlsconf.ClientConfig(),
	)

	return
}
