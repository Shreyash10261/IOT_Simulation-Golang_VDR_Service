package main

import (
	"fmt"
	"log"
	"time"

	"github.com/team/vdr/internal/network"
)

func main() {
	fmt.Println("Virtual Device Runtime (VDR) Engine Started...")
	log.Println("Ready to consume JSON profiles.")

	// Scaffold TUN Interface
	tun, err := network.NewTUNInterface("tun0")
	if err != nil {
		log.Printf("Failed to create TUN interface (requires elevated privileges): %v", err)
		log.Println("VDR continuing without network packet interception...")
	} else {
		// Run the interface listener in the background
		go tun.Listen()
	}

	// Prevent main from exiting immediately
	for {
		time.Sleep(time.Hour)
	}
}
