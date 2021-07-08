package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"github.com/k42-software/go-stream-speed/pkg/discard"
	"github.com/k42-software/go-stream-speed/pkg/fakenet"
	"github.com/k42-software/go-stream-speed/pkg/kcpconf"
	"github.com/xtaci/kcp-go/v5"
	"log"
	"net"
	"sync"
)

func dialMockKCP(
	wg *sync.WaitGroup,
) (
	connType string,
	networkConnection net.Conn,
	cancel context.CancelFunc,
	listenCloser context.CancelFunc,
	err error,
) {
	connType = "KCP"

	ctx, ctxCancel := context.WithCancel(context.Background())
	cancel = ctxCancel

	//
	// Setup a fake virtual network
	//
	localAddr := fakenet.NewAddr(serverHost, clientPort)
	remoteAddr := fakenet.NewAddr(serverHost, kcpPort)
	log.Printf("[DEBUG] Fake Packet Connection %s <-> %s", localAddr, remoteAddr)
	var localPacketConn net.PacketConn
	var remotePacketConn net.PacketConn
	localPacketConn, remotePacketConn, err = fakenet.NewPacketConn(
		ctx,
		localAddr,
		remoteAddr,
		fakenet.StandardMTU,
	)
	if err != nil {
		return
	}

	// KCP conversation ID
	var conversationId uint32
	_ = binary.Read(rand.Reader, binary.LittleEndian, &conversationId)

	//
	// Setup an in-process listener for KCP
	//
	ready := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf(
			"[DEBUG] Receive: New KCP %s ==> %s",
			remoteAddr,
			localAddr,
		)
		var remoteSession *kcp.UDPSession
		remoteSession, err = kcp.NewConn3(
			conversationId,
			localAddr,
			kcpconf.NewBlockCrypt(),
			kcpconf.DataShards,
			kcpconf.ParityShards,
			remotePacketConn,
		)
		if err != nil {
			log.Printf("[ERROR] %s NewConn3: %s", connType, err)
			return
		}
		remoteSession.SetNoDelay(kcpconf.NoDelay, kcpconf.Interval, kcpconf.Resend, kcpconf.NoFlowControl)
		remoteSession.SetStreamMode(kcpconf.StreamMode)
		remoteSession.SetWriteDelay(kcpconf.WriteDelay)
		//remoteSession.SetMtu()
		go discard.StreamConnection(ctx, remoteSession)
		ready <- struct{}{}
	}()
	<-ready

	//
	// Setup an in-process KCP client
	//
	log.Printf(
		"[DEBUG]    Send: New KCP %s ==> %s",
		localAddr,
		remoteAddr,
	)
	var session *kcp.UDPSession
	session, err = kcp.NewConn3(
		conversationId,
		remoteAddr,
		kcpconf.NewBlockCrypt(),
		kcpconf.DataShards,
		kcpconf.ParityShards,
		localPacketConn,
	)
	if err != nil {
		log.Printf("[ERROR] %s NewConn3: %s", connType, err)
		return
	}

	session.SetNoDelay(kcpconf.NoDelay, kcpconf.Interval, kcpconf.Resend, kcpconf.NoFlowControl)
	session.SetStreamMode(kcpconf.StreamMode)
	session.SetWriteDelay(kcpconf.WriteDelay)
	//session.SetMtu()

	networkConnection = session

	return
}
