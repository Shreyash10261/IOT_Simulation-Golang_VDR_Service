package worker

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/team/vdr/internal/protocols/pjlink"
	"github.com/team/vdr/internal/registry"
)

// DeviceWorker represents the execution runtime of a single spawned virtual device.
type DeviceWorker struct {
	device   *registry.Device
	listener net.Listener
	stopChan chan struct{}
}

// NewDeviceWorker initializes a new worker for the given device.
func NewDeviceWorker(dev *registry.Device) *DeviceWorker {
	return &DeviceWorker{
		device:   dev,
		stopChan: make(chan struct{}),
	}
}

// Start opens the TCP socket for command interception.
func (w *DeviceWorker) Start(ctx context.Context) error {
	// Attempt to bind the device's specific IP to the bridge br0
	err := AssignIPToBridge(w.device.IP, "br0")
	if err != nil {
		log.Printf("[%s] Note: could not assign IP to bridge: %v. Proceeding with socket bind...", w.device.ID, err)
	}

	addr := fmt.Sprintf("%s:%d", w.device.IP, w.device.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("[%s] Warning: failed to bind to %s: %v. Falling back to localhost:%d for testing.", w.device.ID, addr, err, w.device.Port)
		addr = fmt.Sprintf("127.0.0.1:%d", w.device.Port)
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to start TCP listener on localhost fallback: %w", err)
		}
	}

	w.listener = listener
	log.Printf("[%s] Worker TCP Server listening on %s", w.device.ID, addr)

	go func() {
		for {
			conn, err := w.listener.Accept()
			if err != nil {
				select {
				case <-w.stopChan:
					return
				default:
					log.Printf("[%s] Accept error: %v", w.device.ID, err)
					continue
				}
			}
			go w.handleConnection(conn)
		}
	}()

	return nil
}

// Stop terminates the worker TCP listener.
func (w *DeviceWorker) Stop() {
	close(w.stopChan)
	if w.listener != nil {
		w.listener.Close()
	}
	log.Printf("[%s] Worker stopped.", w.device.ID)
}

func (w *DeviceWorker) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Timeout connection after inactivity
	conn.SetDeadline(time.Now().Add(60 * time.Second))

	// Send protocol-specific greeting if applicable
	if w.device.Protocol == "PJLink" {
		_, err := conn.Write([]byte("PJLINK 0\r"))
		if err != nil {
			log.Printf("[%s] Error sending PJLink greeting: %v", w.device.ID, err)
			return
		}
	}

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("[%s] Read error: %v", w.device.ID, err)
			}
			break
		}

		requestBytes := buf[:n]
		var responseBytes []byte

		if w.device.Protocol == "PJLink" {
			pjDev, ok := w.device.ProtocolState.(*pjlink.PJLinkDevice)
			if !ok {
				pjDev = pjlink.NewPJLinkDevice()
				w.device.ProtocolState = pjDev
			}

			// Parse and dispatch PJLink command
			responseBytes = pjDev.Handle(requestBytes)
		} else {
			responseBytes = []byte("ERROR: Unsupported Protocol\n")
		}

		_, err = conn.Write(responseBytes)
		if err != nil {
			log.Printf("[%s] Write error: %v", w.device.ID, err)
			break
		}
	}
}
