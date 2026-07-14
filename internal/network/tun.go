package network

import (
	"log"

	"github.com/songgao/water"
)

// VirtualNetworkInterface encapsulates the OS-level TUN/TAP connection
type VirtualNetworkInterface struct {
	ifce *water.Interface
	name string
}

// NewTUNInterface creates and returns a new TUN interface using the water library.
// Note: This requires appropriate OS privileges (NET_ADMIN / sudo) to execute successfully.
func NewTUNInterface(ifceName string) (*VirtualNetworkInterface, error) {
	// We use the default configuration, but specify TUN instead of TAP
	config := water.Config{
		DeviceType: water.TUN,
	}
	
	// Different OSes require different configurations, water handles this cleanly
	// For Linux: config.Name = ifceName
	
	ifce, err := water.New(config)
	if err != nil {
		return nil, err
	}

	return &VirtualNetworkInterface{
		ifce: ifce,
		name: ifce.Name(),
	}, nil
}

// Name returns the actual interface name assigned by the OS (e.g., "tun0")
func (vni *VirtualNetworkInterface) Name() string {
	return vni.name
}

// Listen starts a blocking loop to continuously read raw packets from the interface
func (vni *VirtualNetworkInterface) Listen() {
	log.Printf("Starting virtual interface listener on %s...\n", vni.name)

	// Standard MTU size buffer
	packet := make([]byte, 1500)

	for {
		n, err := vni.ifce.Read(packet)
		if err != nil {
			log.Printf("Error reading from interface %s: %v\n", vni.name, err)
			continue
		}

		// At this point, `packet[:n]` contains the raw IP packet (Layer 3).
		// In the next task, we will send this packet to the PacketDispatcher to check
		// if it's an ICMP ping targeting one of our registered devices.
		
		log.Printf("Received %d bytes on %s\n", n, vni.name)
		
		// TODO: Forward packet to Dispatcher here
	}
}
