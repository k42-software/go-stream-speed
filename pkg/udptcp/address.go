package udptcp

import (
	"errors"
	"fmt"
	"gvisor.dev/gvisor/pkg/tcpip"
	"net"
	"strconv"
)

var _ net.Addr = &FlexiAddr{}

type FlexiAddr tcpip.FullAddress

func (f *FlexiAddr) Network() string {
	return "virtual"
}

func (f *FlexiAddr) String() string {
	return net.JoinHostPort(f.Addr.String(), strconv.Itoa(int(f.Port)))
}

func (f *FlexiAddr) Virtual() *tcpip.FullAddress {
	return (*tcpip.FullAddress)(f)
}

func (f *FlexiAddr) TCP() *net.TCPAddr {
	return &net.TCPAddr{IP: net.IP(f.Addr), Port: int(f.Port)}
}

func (f *FlexiAddr) UDP() *net.UDPAddr {
	return &net.UDPAddr{IP: net.IP(f.Addr), Port: int(f.Port)}
}

func ParseAddress(address string) (addr *FlexiAddr, err error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, errors.New("invalid IP")
	}
	ip = ip.To4()
	if ip == nil {
		return nil, errors.New("not IPv4")
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}
	return &FlexiAddr{
		NIC:  1,
		Addr: tcpip.Address(ip),
		Port: uint16(portInt),
	}, nil
}

func Flexible(address net.Addr) (addr *FlexiAddr, err error) {
	var ip net.IP
	var portInt int
	switch address := address.(type) {
	case *FlexiAddr:
		return address, nil
	case *net.TCPAddr:
		ip = address.IP
		portInt = address.Port
	case *net.UDPAddr:
		ip = address.IP
		portInt = address.Port
	default:
		return nil, fmt.Errorf("unsupported address: %T: %v", address, address)
	}
	return &FlexiAddr{
		NIC:  1,
		Addr: tcpip.Address(ip),
		Port: uint16(portInt),
	}, nil
}
