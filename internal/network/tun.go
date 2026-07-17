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
	// We specify TAP to ensure we receive Layer 2 Ethernet frames (needed for ARP)
	config := water.Config{
		DeviceType: water.TAP,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: ifceName,
		},
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
func (vni *VirtualNetworkInterface) Listen(dispatcher *PacketDispatcher) {
	log.Printf("Starting virtual interface listener on %s...\n", vni.name)

	// Standard MTU size buffer
	packet := make([]byte, 1500)

	for {
		n, err := vni.ifce.Read(packet)
		if err != nil {
			log.Printf("Error reading from interface %s: %v\n", vni.name, err)
			continue
		}

		// Forward the raw packet slice to the dispatcher
		dispatcher.Dispatch(packet[:n])
	}
}
