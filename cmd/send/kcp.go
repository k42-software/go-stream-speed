package main

import (
	"context"
	"github.com/k42-software/go-stream-speed/pkg/kcpconf"
	"github.com/xtaci/kcp-go/v5"
	"log"
	"net"
	"sync"
)

func dialKCP(
	wg *sync.WaitGroup,
) (
	connType string,
	networkConnection net.Conn,
	cancel context.CancelFunc,
	listenCloser context.CancelFunc,
	err error,
) {
	connType = "KCP"

	kcpAddr := net.JoinHostPort(
		serverHost,
		kcpPort,
	)
	log.Printf(
		"[DEBUG] Connecting via %s to %s",
		connType,
		kcpAddr,
	)

	var session *kcp.UDPSession
	session, err = kcp.DialWithOptions(
		kcpAddr,
		kcpconf.NewBlockCrypt(),
		kcpconf.DataShards,
		kcpconf.ParityShards,
	)
	if err != nil {
		log.Printf("[ERROR] %s Dial: %s", connType, err)
		return
	}

	session.SetNoDelay(kcpconf.NoDelay, kcpconf.Interval, kcpconf.Resend, kcpconf.NoFlowControl)
	session.SetStreamMode(kcpconf.StreamMode)
	session.SetWriteDelay(kcpconf.WriteDelay)
	//session.SetMtu()

	networkConnection = session

	return
}
