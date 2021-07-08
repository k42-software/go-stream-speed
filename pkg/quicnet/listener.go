package quicnet

import (
	"context"
	"crypto/tls"
	"github.com/jbenet/go-context/dag"
	"github.com/k42-software/go-stream-speed/pkg/quicconf"
	"github.com/lucas-clemente/quic-go"
	"log"
	"net"
	"sync"
)

// Represent a quic.Listener as a net.Listener
func Listen(packetConn net.PacketConn, tlsConf *tls.Config, config *quic.Config) (*Listener, error) {
	return ListenCtx(context.Background(), packetConn, tlsConf, config)
}

// Represent a quic.Listener as a net.Listener
func ListenCtx(ctx context.Context, packetConn net.PacketConn, tlsConf *tls.Config, config *quic.Config) (*Listener, error) {
	tlsConf, config = quicconf.DefaultServerConfig(tlsConf, config)
	listener, err := quic.Listen(packetConn, tlsConf, config)
	if err != nil {
		return nil, err
	}
	return NewListener(ctx, listener), nil
}

// Represent a quic.Listener as a net.Listener
func ListenAddr(addr string, tlsConf *tls.Config, config *quic.Config) (*Listener, error) {
	return ListenAddrCtx(context.Background(), addr, tlsConf, config)
}

// Represent a quic.Listener as a net.Listener
func ListenAddrCtx(ctx context.Context, addr string, tlsConf *tls.Config, config *quic.Config) (*Listener, error) {
	tlsConf, config = quicconf.DefaultServerConfig(tlsConf, config)
	listener, err := quic.ListenAddr(addr, tlsConf, config)
	if err != nil {
		return nil, err
	}
	return NewListener(ctx, listener), nil
}

var _ net.Listener = &Listener{}

type Listener struct {
	quic.Listener
	ctx    context.Context
	cancel context.CancelFunc

	oneBackgroundWorker sync.Once
	connections         chan net.Conn
}

// Represent a quic.Listener as a net.Listener
func NewListener(ctx context.Context, listener quic.Listener) *Listener {
	ctx, ctxCancel := context.WithCancel(ctx)
	return &Listener{
		Listener:    listener,
		ctx:         ctx,
		cancel:      ctxCancel,
		connections: make(chan net.Conn, 2),
	}
}

func (l *Listener) acceptSessions() {
	var err error
	for {
		var session quic.Session
		session, err = l.Listener.Accept(l.ctx)
		select {
		case <-l.ctx.Done():
			err = l.ctx.Err()
		default:
		}
		if err != nil {
			break
		}
		go l.acceptStreams(session, true)
	}
	log.Printf("[ERROR] QUIC listener: %s", err)
}

func (l *Listener) acceptStreams(session quic.Session, closeSession bool) {
	ctx := ctxext.WithParents(l.ctx, session.Context())
	var err error
	for {
		var stream quic.ReceiveStream
		stream, err = session.AcceptUniStream(ctx)
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:
		}
		if err != nil {
			break
		}
		l.connections <- NewQuicReceiveConn(
			session,
			stream,
			false,
		)
	}
	log.Printf("[ERROR] QUIC listener: %s", err)
	if closeSession {
		var closeCode quic.ApplicationErrorCode = 0
		var closeMessage = ""
		if err != nil {
			closeCode = 1
			closeMessage = err.Error()
		} else if err = ctx.Err(); err != nil {
			closeCode = 1
			closeMessage = err.Error()
		}
		err = session.CloseWithError(closeCode, closeMessage)
		if err != nil {
			log.Printf("[ERROR] QUIC listener: %s", err)
		}
	}
}

func (l *Listener) Accept() (net.Conn, error) {
	l.oneBackgroundWorker.Do(func() {
		go l.acceptSessions()
	})
	select {
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	case connection := <-l.connections:
		return connection, nil
	}
}

func (l *Listener) Close() error {
	select {
	case <-l.ctx.Done():
		return l.ctx.Err()
	default:
		l.cancel()
		return nil
	}
}

func (l *Listener) Addr() net.Addr {
	return l.Listener.Addr()
}
