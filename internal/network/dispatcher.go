package network

import (
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/team/vdr/internal/registry"
)

// PacketDispatcher handles raw network packets and routes them to the correct logic
type PacketDispatcher struct {
	vni      *VirtualNetworkInterface
	registry *registry.DeviceRegistry
}

// NewPacketDispatcher creates a new packet dispatcher
func NewPacketDispatcher(vni *VirtualNetworkInterface, reg *registry.DeviceRegistry) *PacketDispatcher {
	return &PacketDispatcher{
		vni:      vni,
		registry: reg,
	}
}

// Dispatch processes a raw ethernet frame received from the TAP interface
func (pd *PacketDispatcher) Dispatch(rawBytes []byte) {
	// Parse the raw bytes as an Ethernet frame
	packet := gopacket.NewPacket(rawBytes, layers.LayerTypeEthernet, gopacket.Default)
	
	// Check if it's an Ethernet frame
	ethLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		return // Not an ethernet frame
	}
	eth, _ := ethLayer.(*layers.Ethernet)

	// Check if the payload is an ARP packet
	arpLayer := packet.Layer(layers.LayerTypeARP)
	if arpLayer != nil {
		arp, _ := arpLayer.(*layers.ARP)
		pd.handleARP(eth, arp)
	}
}

// handleARP processes ARP requests and generates fake ARP replies for our simulated devices
func (pd *PacketDispatcher) handleARP(eth *layers.Ethernet, arp *layers.ARP) {
	// We only care about ARP Requests (Who has IP X?)
	if arp.Operation != layers.ARPRequest {
		return
	}

	targetIP := net.IP(arp.DstProtAddress)
	
	// Check if the requested IP exists in our Device Registry
	device, exists := pd.registry.GetDeviceByIP(targetIP.String())
	if !exists {
		// The ARP request is for an IP we don't simulate, ignore it
		return
	}

	log.Printf("[ARP Intercept] Request for %s. Forging reply with MAC: %v", targetIP.String(), device.MACAddress)

	// Forge the ARP Reply
	arpReply := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPReply,
		SourceHwAddress:   device.MACAddress,     // Our simulated device MAC
		SourceProtAddress: arp.DstProtAddress,    // Our simulated device IP
		DstHwAddress:      arp.SourceHwAddress,   // The MAC of the original requester
		DstProtAddress:    arp.SourceProtAddress, // The IP of the original requester
	}

	// Forge the Ethernet header for the reply
	ethReply := &layers.Ethernet{
		SrcMAC:       device.MACAddress,
		DstMAC:       eth.SrcMAC,
		EthernetType: layers.EthernetTypeARP,
	}

	// Serialize the packet back into raw bytes
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	err := gopacket.SerializeLayers(buf, opts, ethReply, arpReply)
	if err != nil {
		log.Printf("Error serializing ARP reply: %v", err)
		return
	}

	// Inject the forged raw bytes back into the TAP interface
	_, err = pd.vni.ifce.Write(buf.Bytes())
	if err != nil {
		log.Printf("Error writing ARP reply to TAP interface: %v", err)
	}
}
