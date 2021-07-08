package fakenet

import (
	"context"
	"net"
)

var _ net.PacketConn = &PacketConn{}

// Simulated Network Packet Connection
type PacketConn struct {
	net.Conn
}

func NewPacketConn(ctx context.Context, localAddr, remoteAddr net.Addr, mtu int64) (localPacketConn, remotePacketConn net.PacketConn, err error) {
	var localConn net.Conn
	var remoteConn net.Conn
	{
		var localSend NetworkPipe
		var remoteSend NetworkPipe
		if localSend, err = NewPacketPipe(mtu); err != nil {
			return nil, nil, err
		}
		if remoteSend, err = NewPacketPipe(mtu); err != nil {
			return nil, nil, err
		}
		localConn, remoteConn, err = newConn(ctx, localAddr, remoteAddr, localSend, remoteSend)
	}
	return &PacketConn{localConn}, &PacketConn{remoteConn}, err
}

func (c *PacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, err = c.Read(p)
	return n, c.RemoteAddr(), err
}

func (c *PacketConn) WriteTo(p []byte, _ net.Addr) (n int, err error) {
	return c.Write(p)
}
