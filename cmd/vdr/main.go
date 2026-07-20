package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/team/vdr/internal/network"
	"github.com/team/vdr/internal/registry"
	"github.com/team/vdr/internal/worker"
)

func main() {
	fmt.Println("Virtual Device Runtime (VDR) Engine Started...")
	log.Println("Ready to consume JSON profiles.")

	// 1. Initialize the Device Registry
	reg := registry.NewDeviceRegistry()

	// 2. Initialize and start the Worker Manager
	mgr := worker.NewWorkerManager(reg)

	configPath := os.Getenv("DEVICES_CONFIG")
	if configPath == "" {
		configPath = "devices.json"
	}

	log.Printf("Loading device configurations from %s...", configPath)
	if err := mgr.LoadAndSpawn(configPath); err != nil {
		log.Printf("Failed to load and spawn virtual devices: %v", err)
		log.Println("VDR continuing with empty registry...")
	}

	// 3. Scaffold TAP Interface
	var tun *network.VirtualNetworkInterface
	var err error
	tun, err = network.NewTUNInterface("tap0")
	if err != nil {
		log.Printf("Failed to create TAP interface (requires elevated privileges): %v", err)
		log.Println("VDR continuing without network packet interception...")
	} else {
		// 4. Initialize the Packet Dispatcher
		dispatcher := network.NewPacketDispatcher(tun, reg)

		// 5. Run the interface listener in the background
		go tun.Listen(dispatcher)
	}

	// Set up signal handling for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-stop

	log.Println("Shutting down VDR Engine...")
	mgr.StopAll()
	if tun != nil {
		if err := tun.Close(); err != nil {
			log.Printf("Error closing TAP interface: %v", err)
		}
	}
	log.Println("VDR Engine stopped cleanly.")
}
