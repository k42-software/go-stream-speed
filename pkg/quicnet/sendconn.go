package quicnet

import (
	"github.com/k42-software/go-multierror/v2"
	"github.com/lucas-clemente/quic-go"
	"net"
	"time"
)

var _ net.Conn = &quicSendConn{}

// Represent a quic.SendStream as a net.Conn
type quicSendConn struct {
	quic.SendStream
	session      quic.Session
	closeSession bool
}

// Represent a quic.SendStream as a net.Conn
func NewQuicSendConn(session quic.Session, sendStream quic.SendStream, closeSession bool) net.Conn {
	return &quicSendConn{
		SendStream:   sendStream,
		session:      session,
		closeSession: closeSession,
	}
}

func (q *quicSendConn) LocalAddr() net.Addr {
	return q.session.LocalAddr()
}

func (q *quicSendConn) RemoteAddr() net.Addr {
	return q.session.RemoteAddr()
}

func (q *quicSendConn) Close() (err error) {
	err = q.SendStream.Close()
	if q.closeSession {
		err = multierror.Append(err, q.session.CloseWithError(0, ""))
	}
	return err
}

func (q *quicSendConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (q *quicSendConn) SetDeadline(t time.Time) error {
	return nil
}

func (q *quicSendConn) SetReadDeadline(t time.Time) error {
	return nil
}
