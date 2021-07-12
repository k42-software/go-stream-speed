package udptcp

import (
	"context"
	"fmt"
	"gvisor.dev/gvisor/pkg/sentry/socket/netstack"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"log"
	"net"
)

type VirtualNetwork struct {
	stack     *stack.Stack
	localAddr *tcpip.FullAddress
	cancel    context.CancelFunc
}

func fatalErr(context string, err interface{}) {
	if err != nil {
		switch e := err.(type) {
		case error:
			err = fmt.Errorf("%s %w", context, e)
		case tcpip.Error:
			err = fmt.Errorf("%s %s", context, e.String())
		default:
			err = fmt.Errorf("%s %v", context, err)
		}
		log.Fatalf("[ERROR] %s", err.(error))
	}
}

// Attach a pure golang user mode network stack to a net.PacketConn
func NewVirtualNetwork(packetConn net.PacketConn, mtu int) (vn *VirtualNetwork) {

	var (
		err      error
		tcpipErr tcpip.Error
	)
	_ = err

	if mtu == 0 {
		mtu = defaultMtu
	}

	localAddr, _ := Flexible(packetConn.LocalAddr())
	vn = &VirtualNetwork{
		stack: stack.New(stack.Options{
			NetworkProtocols: []stack.NetworkProtocolFactory{
				ipv4.NewProtocolWithOptions(ipv4.Options{
					AllowExternalLoopbackTraffic: true,
				}),
			},
			TransportProtocols: []stack.TransportProtocolFactory{
				tcp.NewProtocol,
				udp.NewProtocol,
			},
			HandleLocal: false,
			Stats:       netstack.Metrics,
		}),
		localAddr: localAddr.Virtual(),
	}

	// Use the cubic congestion control protocol
	congestionControlOption := tcpip.CongestionControlOption("cubic")
	vn.stack.SetTransportProtocolOption(tcp.ProtocolNumber, &congestionControlOption)

	// Only allow the use of our real local port number
	vn.stack.PortManager.SetPortRange(localAddr.Port, localAddr.Port)

	//go func() {
	//	t := time.NewTicker(time.Second)
	//	for {
	//		<-t.C
	//		spew.Dump(netstack.Metrics.IP)
	//		//spew.Dump(netstack.Metrics.TCP)
	//	}
	//}()

	udpLinkEP := NewEndpoint(packetConn, localAddr.Virtual(), uint32(mtu))

	var linkEP stack.LinkEndpoint = udpLinkEP

	//// PCAP
	//filename := strings.ReplaceAll("net-"+packetConn.LocalAddr().String()+".pcap", ":", "-")
	//pcapFH, err := os.Create(filename)
	//fatalErr("creating pcap file", err)
	//linkEP, err := sniffer.NewWithWriter(udpLinkEP, pcapFH, defaultMtu*2)
	//fatalErr("configuring pcap sniffer", err)

	//linkEP = sniffer.New(linkEP) // log packets to terminal

	// NIC
	tcpipErr = vn.stack.CreateNIC(1, linkEP)
	fatalErr("create virtual network interface", tcpipErr)

	// IP
	tcpipErr = vn.stack.AddAddress(1, ipv4.ProtocolNumber, localAddr.Addr)
	fatalErr("add ip address to network interface", tcpipErr)
	vn.stack.SetRouteTable([]tcpip.Route{
		{Destination: header.IPv4EmptySubnet, NIC: 1}, // Default Route
	})

	vn.cancel = func() {
		//_ = pcapFH.Close()
		udpLinkEP.Close()
		vn.stack.Close()
	}

	return vn
}

func (vn *VirtualNetwork) Close() error {
	vn.cancel()
	return nil
}

// Dial a virtual TCP session to a real UDP address.
func (vn *VirtualNetwork) Dial(ctx context.Context, address *net.UDPAddr) (*gonet.TCPConn, error) {
	//realAddr, err := ParseAddress(address)
	realAddr, err := Flexible(address)
	if err != nil {
		return nil, err
	}
	return gonet.DialContextTCP(
		ctx,
		vn.stack,
		*realAddr.Virtual(),
		ipv4.ProtocolNumber,
	)
}

// Listen for new virtual TCP sessions arriving over the virtual network.
func (vn *VirtualNetwork) Listen() (*gonet.TCPListener, error) {
	return gonet.ListenTCP(
		vn.stack,
		*vn.localAddr,
		ipv4.ProtocolNumber,
	)
}
