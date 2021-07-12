package udptcp

import (
	"context"
	"fmt"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"net"
)

// Attach a pure golang user mode network stack to a UDP socket.
func Bind(localAddr *net.UDPAddr) (*VirtualNetwork, error) {
	packetConn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("udp dial: %w", err)
	}
	return NewVirtualNetwork(packetConn, defaultMtu), nil
}

// Listen for TCP over a UDP socket (using a pure golang user mode network stack).
func Listen(localAddr *net.UDPAddr) (listener *gonet.TCPListener, err error) {
	// FIXME: that this is leaky - not currently suitable for long running processes.
	var virtualNetwork *VirtualNetwork
	virtualNetwork, err = Bind(localAddr)
	if err != nil {
		return nil, err
	}
	listener, err = virtualNetwork.Listen()
	if err != nil {
		return nil, err
	}
	return listener, nil
}

// Dial TCP over a UDP socket (using a pure golang user mode network stack).
func Dial(localAddr, remoteAddr *net.UDPAddr) (conn *gonet.TCPConn, err error) {
	// FIXME: that this is leaky - not currently suitable for long running processes.
	var virtualNetwork *VirtualNetwork
	virtualNetwork, err = Bind(localAddr)
	if err != nil {
		return nil, err
	}
	conn, err = virtualNetwork.Dial(context.Background(), remoteAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
