package fakenet

import (
	"context"
	"errors"
	"io"
	"net"
	"time"
)

var _ net.Conn = &Conn{}
var _ NetworkPipe = &Conn{}

// Simulated Network (Stream) Connection (point to point)
type Conn struct {
	context       context.Context
	localAddr     net.Addr
	remoteAddr    net.Addr
	reader        io.ReadCloser
	writer        io.WriteCloser
	readDeadline  time.Time
	writeDeadline time.Time
}

func NewConn(ctx context.Context, localAddr, remoteAddr net.Addr, mtu int64) (localConn, remoteConn net.Conn, err error) {
	var localSend NetworkPipe
	var remoteSend NetworkPipe
	if localSend, err = NewStreamPipe(mtu); err != nil {
		return nil, nil, err
	}
	if remoteSend, err = NewStreamPipe(mtu); err != nil {
		return nil, nil, err
	}
	return newConn(ctx, localAddr, remoteAddr, localSend, remoteSend)
}

func newConn(ctx context.Context, localAddr, remoteAddr net.Addr, localSend, remoteSend NetworkPipe, ) (localConn, remoteConn net.Conn, err error) {
	return &Conn{
			context:    ctx,
			localAddr:  localAddr,
			remoteAddr: remoteAddr,
			reader:     remoteSend,
			writer:     localSend,
		}, &Conn{
			context:    ctx,
			localAddr:  remoteAddr,
			remoteAddr: localAddr,
			reader:     localSend,
			writer:     remoteSend,
		}, nil
}

func (c *Conn) MTU() int64 {
	if p, ok := c.reader.(NetworkPipe); ok {
		return p.MTU()
	}
	if p, ok := c.writer.(NetworkPipe); ok {
		return p.MTU()
	}
	return StandardMTU
}

func (c *Conn) Context() context.Context {
	return c.context
}

func (c *Conn) checkDeadline(deadline time.Time) error {
	if !deadline.IsZero() {
		if time.Until(deadline) <= 0 {
			return context.DeadlineExceeded
		}
	}
	return nil
}

func (c *Conn) Read(b []byte) (n int, err error) {
	if err = c.checkDeadline(c.readDeadline); err != nil {
		return 0, err
	}
	select {
	case <-c.context.Done():
		return 0, c.context.Err()
	default:
		n, err = c.reader.Read(b)
		//err = appendErr(err, c.context.Err()) // This has a mutex in it :(
		//err = appendErr(err, c.checkDeadline(c.readDeadline))
		return n, err
	}
}

func (c *Conn) Write(b []byte) (n int, err error) {
	if err = c.checkDeadline(c.writeDeadline); err != nil {
		return 0, err
	}
	select {
	case <-c.context.Done():
		return 0, c.context.Err()
	default:
		n, err = c.writer.Write(b)
		//err = appendErr(err, c.context.Err()) // This has a mutex in it :(
		//err = appendErr(err, c.checkDeadline(c.writeDeadline))
		return n, err
	}
}

func (c *Conn) Close() (err error) {
	err = appendErr(err, c.reader.Close())
	err = appendErr(err, c.writer.Close())
	return err
}

func (c *Conn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

// NB deadlines are best efforts and don't work correctly
func (c *Conn) SetDeadline(t time.Time) (err error) {
	err = appendErr(err, c.SetReadDeadline(t))
	err = appendErr(err, c.SetWriteDeadline(t))
	return err
}

// NB deadlines are best efforts and don't work correctly
func (c *Conn) SetReadDeadline(t time.Time) error {
	if time.Until(t) <= 0 {
		return errors.New("deadline already past")
	}
	c.readDeadline = t
	return nil
}

// NB deadlines are best efforts and don't work correctly
func (c *Conn) SetWriteDeadline(t time.Time) error {
	if time.Until(t) <= 0 {
		return errors.New("deadline already past")
	}
	c.writeDeadline = t
	return nil
}
