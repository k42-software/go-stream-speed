package udptcp

import (
	"context"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/header/parse"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"io"
	"log"
	"math"
	"net"
)

const queueSize = 4096
const defaultMtu = 1500

var _ stack.LinkEndpoint = (*Endpoint)(nil)
var _ stack.GSOEndpoint = (*Endpoint)(nil)

// Endpoint is link layer endpoint that stores outbound packets in a channel
// and allows injection of inbound packets.
type Endpoint struct {
	ctx        context.Context
	cancel     context.CancelFunc
	packetConn net.PacketConn
	channelEP  *channel.Endpoint
	localAddr  *tcpip.FullAddress // TODO not used? remove me?
}

func NewEndpoint(packetConn net.PacketConn, localAddr *tcpip.FullAddress, mtu uint32) *Endpoint {
	return NewEndpointContext(context.Background(), packetConn, localAddr, mtu)
}

func NewEndpointContext(ctx context.Context, packetConn net.PacketConn, localAddr *tcpip.FullAddress, mtu uint32) *Endpoint {
	epCtx, cancel := context.WithCancel(ctx)
	if mtu == 0 {
		mtu = defaultMtu
	}
	e := &Endpoint{
		ctx:        epCtx,
		cancel:     cancel,
		packetConn: packetConn,
		channelEP: channel.New(
			queueSize,
			mtu,
			"",
		),
		localAddr: localAddr,
	}
	go e.backgroundSender()
	go e.backgroundReceiver()
	return e
}

func (e *Endpoint) Close() {
	e.cancel()
	e.channelEP.Close()
	if closer, ok := e.packetConn.(io.Closer); ok {
		_ = closer.Close()
	}
}

// Context returns the value initialized during construction.
func (e *Endpoint) Context() context.Context {
	return e.ctx
}

func (e *Endpoint) readPacketHeaders(pkt *stack.PacketBuffer) (localAddr, remoteAddr *FlexiAddr) {

	//Read the Network Level Header (IPv4)
	if ok := parse.IPv4(pkt); !ok {
		log.Printf("[ERROR] Unable to parse packet as IPv4")
		return
	}
	netHeader := header.IPv4(pkt.NetworkHeader().View())
	fragmentOffset := netHeader.FragmentOffset()
	moreFragments := netHeader.Flags()&header.IPv4FlagMoreFragments == header.IPv4FlagMoreFragments
	//src := netHeader.SourceAddress()
	//dst := netHeader.DestinationAddress()
	//transProto := netHeader.Protocol()
	//size := netHeader.TotalLength() - uint16(netHeader.HeaderLength())
	//id := uint32(netHeader.ID())

	if fragmentOffset > 0 || moreFragments {
		log.Printf("[ERROR] Fragmented packets are not supported")
		return
	}

	srcPort := uint16(0)
	dstPort := uint16(0)
	switch tcpip.TransportProtocolNumber(netHeader.Protocol()) {
	case header.UDPProtocolNumber:
		if ok := parse.UDP(pkt); !ok {
			log.Printf("[ERROR] Unable to parse packet as UDP")
			return
		}

		udp := header.UDP(pkt.TransportHeader().View())
		srcPort = udp.SourcePort()
		dstPort = udp.DestinationPort()

	case header.TCPProtocolNumber:
		if ok := parse.TCP(pkt); !ok {
			log.Printf("[ERROR] Unable to parse packet as TCP")
			return
		}

		tcp := header.TCP(pkt.TransportHeader().View())
		offset := int(tcp.DataOffset())
		if offset < header.TCPMinimumSize {
			log.Printf("[ERROR] invalid packet: tcp data offset too small %d", offset)
			return
		}
		if size := pkt.Data().Size() + len(tcp); offset > size && !moreFragments {
			log.Printf("[ERROR] invalid packet: tcp data offset %d larger than tcp packet length %d", offset, size)
			return
		}

		srcPort = tcp.SourcePort()
		dstPort = tcp.DestinationPort()
	}

	//log.Printf(
	//	"[DEBUG] src=%s:%v dst=%s:%v frag=%v",
	//	netHeader.SourceAddress(), srcPort,
	//	netHeader.DestinationAddress(), dstPort,
	//	moreFragments,
	//)

	localAddr = &FlexiAddr{
		NIC:  1,
		Addr: netHeader.SourceAddress(),
		Port: srcPort,
	}
	remoteAddr = &FlexiAddr{
		NIC:  1,
		Addr: netHeader.DestinationAddress(),
		Port: dstPort,
	}

	return
}

func (e *Endpoint) addIPHeader(
	srcAddr, dstAddr tcpip.Address,
	pkt *stack.PacketBuffer,
	params stack.NetworkHeaderParams,
	options header.IPv4OptionsSerializer,
) tcpip.Error {
	hdrLen := header.IPv4MinimumSize
	var optLen int
	if options != nil {
		optLen = int(options.Length())
	}
	hdrLen += optLen
	if hdrLen > header.IPv4MaximumHeaderSize {
		return &tcpip.ErrMessageTooLong{}
	}
	ipH := header.IPv4(pkt.NetworkHeader().Push(hdrLen))
	length := pkt.Size()
	if length > math.MaxUint16 {
		return &tcpip.ErrMessageTooLong{}
	}
	ipH.Encode(&header.IPv4Fields{
		TotalLength: uint16(length),
		Flags:       header.IPv4FlagDontFragment,
		TTL:         params.TTL,
		TOS:         params.TOS,
		Protocol:    uint8(params.Protocol),
		SrcAddr:     srcAddr,
		DstAddr:     dstAddr,
		Options:     options,
	})
	ipH.SetChecksum(^ipH.CalculateChecksum())
	pkt.NetworkProtocolNumber = ipv4.ProtocolNumber
	return nil
}

func (e *Endpoint) backgroundSender() {
	mtu := e.MTU()
SendLoop:
	for {
		mtu = e.MTU()
		packet, ok := e.channelEP.ReadContext(e.ctx)
		if !ok {
			select {
			case <-e.ctx.Done():
				log.Printf("[DEBUG] context error: %s", e.ctx.Err())
			default:
				log.Printf("[DEBUG] unknown network stack endpoint read error")
			}
			break SendLoop
		}

		if packet.Pkt.Size() > int(mtu) {
			log.Printf("[ERROR] packet is bigger than MTU: %d > %d",
				packet.Pkt.Size(), mtu,
			)
			continue SendLoop
		}

		// Get a view of the raw packet (without the non-existent Link Level Header)
		vv := buffer.NewVectorisedView(packet.Pkt.Size(), packet.Pkt.Views())
		vv.TrimFront(len(packet.Pkt.LinkHeader().View()))

		// Identify the correct remote address
		srcAddr, dstAddr := e.readPacketHeaders(stack.NewPacketBuffer(stack.PacketBufferOptions{Data: vv}))
		if dstAddr == nil {
			log.Printf("[ERROR] send: no remote address")
			continue SendLoop
		}
		_ = srcAddr
		//log.Printf("[ERROR] SENDING: %s => %s", srcAddr.String(), dstAddr.String())

		// Send the packet over UDP
		if _, err := e.packetConn.WriteTo(vv.ToView(), dstAddr.UDP()); err != nil {
			log.Printf("[ERROR] send: write: %s", err)
			continue SendLoop
		}
	}
	log.Printf("[DEBUG] background sender stopped")
}

func (e *Endpoint) backgroundReceiver() {
	mtu := e.MTU()
WriteLoop:
	for {
		select {
		case <-e.ctx.Done():
			log.Printf("[DEBUG] context error: %s", e.ctx.Err())
			break WriteLoop
		default:
		}
		mtu = e.MTU()
		buf := make([]byte, mtu)
		n, sendingAddr, err := e.packetConn.ReadFrom(buf)
		if err != nil {
			log.Printf("[ERROR] receive: %s", err)
			continue WriteLoop
		}
		_ = sendingAddr

		// Create a packet buffer from the received data
		vv := buffer.NewVectorisedView(n, []buffer.View{buffer.NewViewFromBytes(buf[:n])})
		pkb := stack.NewPacketBuffer(stack.PacketBufferOptions{Data: vv})

		//// Identify the embedded addresses
		//srcAddr, dstAddr := e.readPacketHeaders(stack.NewPacketBuffer(stack.PacketBufferOptions{Data: vv}))
		//log.Printf("[ERROR] RECEIVED: %s: %s => %s", sendingAddr, srcAddr.String(), dstAddr.String())

		//// Add a Network Level Header
		//remoteAddr, err := e.addresses.LookupVirtual(sendingAddr)
		//if err != nil {
		//	log.Printf("[ERROR] receive: %s", err)
		//	continue
		//}
		//log.Printf(
		//	"[DEBUG] RECEIVED: sender=%s src=%s dst=%s",
		//	sendingAddr, (*FlexiAddr)(remoteAddr).String(), (*FlexiAddr)(e.localAddr).String(),
		//)
		//e.addIPHeader(
		//	remoteAddr.Addr,
		//	e.localAddr.Addr,
		//	pkb,
		//	stack.NetworkHeaderParams{
		//		Protocol: header.TCPProtocolNumber,
		//		TTL:      ipv4.DefaultTTL,
		//	},
		//	nil,
		//)

		// Inject the packet into the virtual network stack
		e.channelEP.InjectInbound(ipv4.ProtocolNumber, pkb)
	}
	log.Printf("[DEBUG] background receiver stopped")
}

// WritePacket stores outbound packets into the channel.
func (e *Endpoint) WritePacket(r stack.RouteInfo, protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) tcpip.Error {
	return e.channelEP.WritePacket(r, protocol, pkt)
}

// WritePackets stores outbound packets into the channel.
func (e *Endpoint) WritePackets(r stack.RouteInfo, pkts stack.PacketBufferList, protocol tcpip.NetworkProtocolNumber) (int, tcpip.Error) {
	return e.channelEP.WritePackets(r, pkts, protocol)
}

// Attach saves the stack network-layer dispatcher for use later when packets
// are injected.
func (e *Endpoint) Attach(dispatcher stack.NetworkDispatcher) {
	e.channelEP.Attach(dispatcher)
}

// IsAttached implements stack.LinkEndpoint.IsAttached.
func (e *Endpoint) IsAttached() bool {
	return e.channelEP.IsAttached()
}

// MTU implements stack.LinkEndpoint.MTU. It returns the value initialized
// during construction.
func (e *Endpoint) MTU() uint32 {
	return e.channelEP.MTU()
}

// LinkAddress returns the link address of this endpoint.
func (e *Endpoint) LinkAddress() tcpip.LinkAddress {
	return e.channelEP.LinkAddress()
}

// Capabilities implements stack.LinkEndpoint.Capabilities.
func (e *Endpoint) Capabilities() stack.LinkEndpointCapabilities {
	return stack.CapabilityNone
}

// GSOMaxSize implements stack.GSOEndpoint.
func (*Endpoint) GSOMaxSize() uint32 {
	return 1 << 15
}

// SupportedGSO implements stack.GSOEndpoint.
func (e *Endpoint) SupportedGSO() stack.SupportedGSO {
	return stack.GSONotSupported
}

// MaxHeaderLength returns the maximum size of the link layer header. Given it
// doesn't have a header, it just returns 0.
func (*Endpoint) MaxHeaderLength() uint16 {
	return 0
}

// ARPHardwareType implements stack.LinkEndpoint.ARPHardwareType.
func (*Endpoint) ARPHardwareType() header.ARPHardwareType {
	return header.ARPHardwareNone
}

// Wait implements stack.LinkEndpoint.Wait.
func (*Endpoint) Wait() {
}

// AddHeader implements stack.LinkEndpoint.AddHeader.
func (*Endpoint) AddHeader(tcpip.LinkAddress, tcpip.LinkAddress, tcpip.NetworkProtocolNumber, *stack.PacketBuffer) {
}
