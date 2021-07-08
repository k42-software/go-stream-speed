package main

import (
	"context"
	"github.com/k42-software/go-stream-speed/pkg/discard"
	"github.com/k42-software/go-stream-speed/pkg/fakenet"
	"github.com/k42-software/go-stream-speed/pkg/quicconf"
	"github.com/k42-software/go-stream-speed/pkg/quicnet"
	"github.com/lucas-clemente/quic-go"
	"log"
	"net"
	"sync"
)

func dialMockQUIC(
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

	netCtx := clientCtx

	//
	// Setup a fake virtual network
	//
	localAddr := fakenet.NewAddr(serverHost, clientPort)
	remoteAddr := fakenet.NewAddr(serverHost, quicPort)
	log.Printf("[DEBUG] Fake Packet Connection %s <-> %s", localAddr, remoteAddr)
	var localPacketConn net.PacketConn
	var remotePacketConn net.PacketConn
	localPacketConn, remotePacketConn, err = fakenet.NewPacketConn(
		netCtx,
		fakenet.NewAddr(serverHost, clientPort),
		remoteAddr,
		fakenet.StandardMTU,
	)
	if err != nil {
		return
	}

	//
	// Setup an in-process listener for QUIC
	//
	listenCtx, listenCtxCancel := context.WithCancel(netCtx)
	listenCloser = listenCtxCancel
	wg.Add(1)
	go func() {
		defer wg.Done()
		quicListen := net.JoinHostPort(serverHost, quicPort)
		log.Printf("[DEBUG] QUIC listening on %s", quicListen)
		listener, err := quicnet.ListenCtx(
			listenCtx,
			remotePacketConn,
			quicconf.ServerTlsConfig(),
			quicconf.ServerConfig(false),
		)
		defer listener.Close()
		if err != nil {
			log.Printf("[ERROR] %s", err)
			return
		}
		for {
			listenerConnection, err := listener.Accept()
			if err != nil {
				log.Printf("[ERROR] %s", err)
				return
			}
			go discard.StreamConnection(
				listenCtx,
				listenerConnection,
			)
		}
	}()

	//
	// Setup  an in-process QUIC client
	//
	log.Printf("[DEBUG] Connecting via %s to %s", connType, remoteAddr)
	var session quic.Session
	session, err = quic.DialContext(
		clientCtx,
		localPacketConn,
		remoteAddr,
		serverHost,
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
