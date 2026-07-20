package network

import (
	"bytes"
	"net"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/team/vdr/internal/registry"
)

func TestPacketDispatcher_ARP(t *testing.T) {
	reg := registry.NewDeviceRegistry()
	dev := &registry.Device{
		ID:         "TestDev",
		IP:         "10.10.1.55",
		MACAddress: []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x55},
		IsOnline:   true,
	}
	reg.RegisterDevice(dev)

	// Create packet dispatcher with mocked write function
	var writtenBytes []byte
	pd := &PacketDispatcher{
		registry: reg,
		writeFn: func(b []byte) (int, error) {
			writtenBytes = b
			return len(b), nil
		},
	}

	// Forge an incoming L2 ARP request for 10.10.1.55
	eth := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
		SourceProtAddress: net.IP{10, 10, 1, 1},
		DstHwAddress:      net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		DstProtAddress:    net.IP{10, 10, 1, 55},
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	_ = gopacket.SerializeLayers(buf, opts, eth, arp)

	// Dispatch the packet
	pd.Dispatch(buf.Bytes())

	if len(writtenBytes) == 0 {
		t.Fatal("No packet injected back by dispatcher")
	}

	// Parse reply
	replyPacket := gopacket.NewPacket(writtenBytes, layers.LayerTypeEthernet, gopacket.Default)
	arpLayer := replyPacket.Layer(layers.LayerTypeARP)
	if arpLayer == nil {
		t.Fatal("Injected packet is not an ARP reply")
	}
	replyArp := arpLayer.(*layers.ARP)
	if replyArp.Operation != layers.ARPReply {
		t.Errorf("Expected ARP Reply, got %v", replyArp.Operation)
	}
	if !bytes.Equal(replyArp.SourceHwAddress, dev.MACAddress) {
		t.Errorf("Expected source MAC %v, got %v", dev.MACAddress, replyArp.SourceHwAddress)
	}
}

func TestPacketDispatcher_IPForwarding(t *testing.T) {
	// Setup Registry
	reg := registry.NewDeviceRegistry()
	dev := &registry.Device{
		ID:         "TestDev",
		IP:         "10.10.1.55",
		MACAddress: []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x55},
		IsOnline:   true,
	}
	reg.RegisterDevice(dev)

	mockInterfaceMAC := []byte{0x00, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE}

	var writtenBytes []byte
	pd := &PacketDispatcher{
		registry:     reg,
		interfaceMAC: mockInterfaceMAC,
		writeFn: func(b []byte) (int, error) {
			writtenBytes = b
			return len(b), nil
		},
	}

	// Construct an IP packet destined for the virtual device's custom MAC
	eth := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
		DstMAC:       dev.MACAddress, // destined for 02:00:00:00:00:55
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip := &layers.IPv4{
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
		SrcIP:    net.IP{10, 10, 1, 1},
		DstIP:    net.IP{10, 10, 1, 55},
	}
	tcp := &layers.TCP{
		SrcPort: 54321,
		DstPort: 4352,
		SYN:     true,
	}
	_ = tcp.SetNetworkLayerForChecksum(ip)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	err := gopacket.SerializeLayers(buf, opts, eth, ip, tcp)
	if err != nil {
		t.Fatalf("Failed to serialize layers: %v", err)
	}

	// Dispatch the packet
	pd.Dispatch(buf.Bytes())

	if len(writtenBytes) == 0 {
		t.Fatal("No packet forwarded/injected back by dispatcher")
	}

	// Verify that the destination MAC was rewritten to our mock interface's MAC address
	rewrittenDstMAC := writtenBytes[0:6]
	if !bytes.Equal(rewrittenDstMAC, mockInterfaceMAC) {
		t.Errorf("Expected rewritten destination MAC to be %v, got %v", mockInterfaceMAC, rewrittenDstMAC)
	}

	// Verify that the rest of the payload remained untouched
	if !bytes.Equal(writtenBytes[6:], buf.Bytes()[6:]) {
		t.Error("Payload or source MAC was incorrectly modified during forwarding")
	}
}
