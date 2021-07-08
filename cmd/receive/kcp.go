package main

import (
	"context"
	"github.com/k42-software/go-stream-speed/pkg/discard"
	"github.com/k42-software/go-stream-speed/pkg/kcpconf"
	"github.com/xtaci/kcp-go/v5"
	"log"
	"net"
	"sync"
)

func listenKcp(wg *sync.WaitGroup) {
	defer wg.Done()

	kcpListen := net.JoinHostPort(serverHost, kcpPort)
	log.Printf("[DEBUG] KCP listening on %s", kcpListen)

	listener, err := kcp.ListenWithOptions(
		kcpListen,
		kcpconf.NewBlockCrypt(),
		kcpconf.DataShards,
		kcpconf.ParityShards,
	)
	if err != nil {
		log.Fatalf("[ERROR] %s", err)
	}

	for {
		session, err := listener.AcceptKCP()
		if err != nil {
			log.Printf("[ERROR] %s", err)
			continue
		}
		session.SetNoDelay(kcpconf.NoDelay, kcpconf.Interval, kcpconf.Resend, kcpconf.NoFlowControl)
		session.SetStreamMode(kcpconf.StreamMode)
		session.SetWriteDelay(kcpconf.WriteDelay)
		//session.SetMtu()
		var networkConnection net.Conn = session
		go discard.StreamConnection(
			context.Background(),
			networkConnection,
		)
	}
}
