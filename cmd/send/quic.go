package main

import (
	"context"
	"github.com/k42-software/go-stream-speed/pkg/quicconf"
	"github.com/k42-software/go-stream-speed/pkg/quicnet"
	"github.com/lucas-clemente/quic-go"
	"log"
	"net"
	"sync"
)

func dialQUIC(
	wg *sync.WaitGroup,
) (
	connType string,
	networkConnection net.Conn,
	cancel context.CancelFunc,
	listenCloser context.CancelFunc,
	err error,
) {
	connType = "QUIC"

	clientCtx, clientCtxCancel := context.WithCancel(context.Background())
	cancel = clientCtxCancel

	quicAddr := net.JoinHostPort(
		serverHost,
		quicPort,
	)

	log.Printf(
		"[DEBUG] Connecting via %s to %s",
		connType,
		quicAddr,
	)

	var session quic.Session

	session, err = quic.DialAddrContext(
		clientCtx,
		quicAddr,
		quicconf.ClientTlsConfig(),
		quicconf.ClientConfig(false),
	)
	if err != nil {
		log.Printf("[ERROR] dial: %s", err)
		return
	}

	if session == nil {
		log.Printf("[ERROR] quic session is nil")
		return
	}

	var stream quic.SendStream
	stream, err = session.OpenUniStreamSync(clientCtx)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return
	}

	networkConnection = quicnet.NewQuicSendConn(
		session,
		stream,
		true,
	)

	return
}
