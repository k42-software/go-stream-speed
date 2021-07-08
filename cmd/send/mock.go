package main

import (
	"context"
	"github.com/k42-software/go-stream-speed/pkg/discard"
	"github.com/k42-software/go-stream-speed/pkg/fakenet"
	"log"
	"net"
	"sync"
)

func dialMockStream(
	wg *sync.WaitGroup,
) (
	connType string,
	networkConnection net.Conn,
	cancel context.CancelFunc,
	listenCloser context.CancelFunc,
	err error,
) {

	connType = "MOCK Stream"

	ctx, ctxCancel := context.WithCancel(context.Background())
	cancel = ctxCancel

	//////////////////////////////////////////////////////////////////////////
	// Fake Virtual Network
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

	//////////////////////////////////////////////////////////////////////////
	// Receiver
	log.Printf(
		"[DEBUG] %s server receiving on %s",
		connType,
		remoteConn.LocalAddr(),
	)
	go discard.StreamConnection(
		ctx,
		remoteConn,
	)

	//////////////////////////////////////////////////////////////////////////
	// Client
	log.Printf(
		"[DEBUG] Connecting via %s to %s",
		connType,
		localConn.RemoteAddr(),
	)
	networkConnection = localConn

	return
}
