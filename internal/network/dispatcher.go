package network

import (
	"bytes"
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/team/vdr/internal/registry"
)

// PacketDispatcher handles raw network packets and routes them to the correct logic
type PacketDispatcher struct {
	vni          *VirtualNetworkInterface
	registry     *registry.DeviceRegistry
	writeFn      func([]byte) (int, error) // for mocking and virtual interface decoupling
	interfaceMAC net.HardwareAddr          // cached MAC address of the host interface
}

// NewPacketDispatcher creates a new packet dispatcher
func NewPacketDispatcher(vni *VirtualNetworkInterface, reg *registry.DeviceRegistry) *PacketDispatcher {
	var writeFn func([]byte) (int, error)
	var interfaceMAC net.HardwareAddr
	if vni != nil {
		if vni.ifce != nil {
			writeFn = vni.ifce.Write
		}
		// Retrieve and cache details of the host interface to get its MAC
		ifce, err := net.InterfaceByName(vni.name)
		if err == nil {
			interfaceMAC = ifce.HardwareAddr
		} else {
			log.Printf("Warning: could not retrieve hardware address for interface %s: %v", vni.name, err)
		}
	}
	return &PacketDispatcher{
		vni:          vni,
		registry:     reg,
		writeFn:      writeFn,
		interfaceMAC: interfaceMAC,
	}
}

// Dispatch processes a raw ethernet frame received from the TAP interface.
// If the packet is destined for a simulated device, it either handles ARP queries
// or rewrites the destination MAC address to the local interface MAC and re-injects
// it so the host kernel's TCP stack can process it.
func (pd *PacketDispatcher) Dispatch(rawBytes []byte) {
	if len(rawBytes) < 14 {
		return // Too short to be an Ethernet frame
	}

	// Extract destination MAC and EtherType
	dstMAC := rawBytes[0:6]
	etherType := (uint16(rawBytes[12]) << 8) | uint16(rawBytes[13])

	// Check if the payload is an ARP packet
	if etherType == 0x0806 {
		packet := gopacket.NewPacket(rawBytes, layers.LayerTypeEthernet, gopacket.Default)
		ethLayer := packet.Layer(layers.LayerTypeEthernet)
		arpLayer := packet.Layer(layers.LayerTypeARP)
		if ethLayer != nil && arpLayer != nil {
			eth, _ := ethLayer.(*layers.Ethernet)
			arp, _ := arpLayer.(*layers.ARP)
			pd.handleARP(eth, arp)
		}
		return
	}

	// For other network packets (e.g., IPv4/IPv6), check if destination MAC matches any simulated device
	if pd.registry == nil {
		return
	}
	devices := pd.registry.GetAllDevices()
	var matchedDevice *registry.Device
	for _, dev := range devices {
		if bytes.Equal(dstMAC, dev.MACAddress) {
			matchedDevice = dev
			break
		}
	}

	if matchedDevice != nil {
		if pd.writeFn == nil || len(pd.interfaceMAC) == 0 {
			return
		}

		// Rewrite destination MAC directly in raw bytes to match the local interface
		copy(rawBytes[0:6], pd.interfaceMAC)

		// Inject back to TAP interface so the kernel routes it locally
		_, err := pd.writeFn(rawBytes)
		if err != nil {
			log.Printf("Error injecting IP packet back to TAP interface: %v", err)
		}
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

	// Inject the forged raw bytes back into the TAP interface using writeFn
	if pd.writeFn != nil {
		_, err = pd.writeFn(buf.Bytes())
		if err != nil {
			log.Printf("Error writing ARP reply to TAP interface: %v", err)
		}
	}
}
