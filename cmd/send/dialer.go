package main

import (
	"context"
	"net"
	"sync"
)

type testDialer func(
	wg *sync.WaitGroup,
) (
	connType string,
	networkConnection net.Conn,
	cancel context.CancelFunc,
	listenCloser context.CancelFunc,
	err error,
)
