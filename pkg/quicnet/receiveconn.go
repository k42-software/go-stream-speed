package quicnet

import (
	"github.com/lucas-clemente/quic-go"
	"net"
	"time"
)

var _ net.Conn = &quicReceiveConn{}

// Represent a quic.ReceiveStream as a net.Conn
type quicReceiveConn struct {
	quic.ReceiveStream
	session      quic.Session
	closeSession bool
}

// Represent a quic.ReceiveStream as a net.Conn
func NewQuicReceiveConn(session quic.Session, receiveStream quic.ReceiveStream, closeSession bool) net.Conn {
	return &quicReceiveConn{
		ReceiveStream:   receiveStream,
		session:      session,
		closeSession: closeSession,
	}
}

func (q *quicReceiveConn) LocalAddr() net.Addr {
	return q.session.LocalAddr()
}

func (q *quicReceiveConn) RemoteAddr() net.Addr {
	return q.session.RemoteAddr()
}

func (q *quicReceiveConn) Close() (err error) {
	if q.closeSession {
		err = q.session.CloseWithError(0, "")
	}
	return err
}

func (q *quicReceiveConn) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (q *quicReceiveConn) SetDeadline(t time.Time) error {
	return nil
}

func (q *quicReceiveConn) SetWriteDeadline(t time.Time) error {
	return nil
}
