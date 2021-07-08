package quicnet

import (
	"github.com/k42-software/go-multierror/v2"
	"github.com/lucas-clemente/quic-go"
	"net"
)

var _ net.Conn = &quicConn{}

// Represent a quic.Stream as a net.Conn
type quicConn struct {
	quic.Stream
	session      quic.Session
	closeSession bool
}

// Represent a quic.Stream as a net.Conn
func NewQuicConn(session quic.Session, stream quic.Stream, closeSession bool) net.Conn {
	return &quicConn{
		Stream:       stream,
		session:      session,
		closeSession: closeSession,
	}
}

func (q *quicConn) LocalAddr() net.Addr {
	return q.session.LocalAddr()
}

func (q *quicConn) RemoteAddr() net.Addr {
	return q.session.RemoteAddr()
}

func (q *quicConn) Close() (err error) {
	err = q.Stream.Close()
	if q.closeSession {
		err = multierror.Append(err, q.session.CloseWithError(0, ""))
	}
	return err
}
