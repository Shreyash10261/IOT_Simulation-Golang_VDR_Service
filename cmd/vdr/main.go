package main

import (
	"fmt"
	"log"
	"time"

	"github.com/team/vdr/internal/network"
	"github.com/team/vdr/internal/registry"
)

func main() {
	fmt.Println("Virtual Device Runtime (VDR) Engine Started...")
	log.Println("Ready to consume JSON profiles.")

	// 1. Initialize the Device Registry (Task 162)
	reg := registry.NewDeviceRegistry()
	
	// Create a test Projector and add it to the registry
	testProjector := &registry.Device{
		IP:         "10.10.1.55",
		MACAddress: []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x55}, // Fake MAC
		IsOnline:   true,
	}
	reg.RegisterDevice(testProjector)
	log.Printf("Registered Test Projector at IP: %s, MAC: %x", testProjector.IP, testProjector.MACAddress)

	// 2. Scaffold TAP Interface
	tun, err := network.NewTUNInterface("tap0")
	if err != nil {
		log.Printf("Failed to create TAP interface (requires elevated privileges): %v", err)
		log.Println("VDR continuing without network packet interception...")
	} else {
		// 3. Initialize the Packet Dispatcher (Task 163)
		dispatcher := network.NewPacketDispatcher(tun, reg)
		
		// 4. Run the interface listener in the background
		go tun.Listen(dispatcher)
	}

	// Prevent main from exiting immediately
	for {
		time.Sleep(time.Hour)
	}
}
