package fakenet

import (
	"net"
)

const network = "fake"

var _ net.Addr = &FakeAddress{}

type FakeAddress struct {
	str string
}

func ResolveFakeAddr(network, address string) (*FakeAddress, error) {
	_ = network
	return &FakeAddress{address}, nil
}

func NewAddr(host, port string) (addr net.Addr) {
	addr, _ = ResolveFakeAddr(network, net.JoinHostPort(host, port))
	return addr
}

func (addr *FakeAddress) Network() string {
	return network
}

func (addr *FakeAddress) String() string {
	return addr.str
}
