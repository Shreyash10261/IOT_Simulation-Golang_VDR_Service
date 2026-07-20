package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/team/vdr/internal/models"
	"github.com/team/vdr/internal/protocols/pjlink"
	"github.com/team/vdr/internal/registry"
)

// DeviceConfig represents the configuration layout of a single simulated device in JSON.
type DeviceConfig struct {
	ID           string                  `json:"id"`
	IP           string                  `json:"ip"`
	MAC          string                  `json:"mac"`
	Port         int                     `json:"port"`
	Protocol     string                  `json:"protocol"`
	Manufacturer string                  `json:"manufacturer"`
	Model        string                  `json:"model"`
	Telemetry    []models.TelemetryField `json:"telemetry"`
}

// WorkerManager manages all virtual device workers and the background telemetry scheduler.
type WorkerManager struct {
	registry *registry.DeviceRegistry
	workers  map[string]*DeviceWorker
	mu       sync.Mutex
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewWorkerManager initializes a new WorkerManager.
func NewWorkerManager(reg *registry.DeviceRegistry) *WorkerManager {
	return &WorkerManager{
		registry: reg,
		workers:  make(map[string]*DeviceWorker),
		stopChan: make(chan struct{}),
	}
}

// LoadAndSpawn reads device configurations from JSON, registers devices, and boots their worker servers.
func (m *WorkerManager) LoadAndSpawn(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read device config file: %w", err)
	}

	var configs []DeviceConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to parse device config: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cfg := range configs {
		macAddr, err := net.ParseMAC(cfg.MAC)
		if err != nil {
			log.Printf("Warning: invalid MAC address %s for device %s, generating default", cfg.MAC, cfg.ID)
			macAddr = []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x01}
		}

		dev := &registry.Device{
			ID:              cfg.ID,
			IP:              cfg.IP,
			MACAddress:      macAddr,
			Port:            cfg.Port,
			Protocol:        cfg.Protocol,
			Manufacturer:    cfg.Manufacturer,
			Model:           cfg.Model,
			IsOnline:        true,
			TelemetryFields: cfg.Telemetry,
			Telemetry:       make(map[string]interface{}),
		}

		if cfg.Protocol == "PJLink" {
			dev.ProtocolState = pjlink.NewPJLinkDevice()
		}

		// Register in central DeviceRegistry
		m.registry.RegisterDevice(dev)

		// Create worker
		worker := NewDeviceWorker(dev)
		m.workers[cfg.ID] = worker

		// Start worker listener
		ctx := context.Background()
		if err := worker.Start(ctx); err != nil {
			log.Printf("Error starting worker for device %s: %v", cfg.ID, err)
		} else {
			log.Printf("Successfully spawned virtual device %s (%s) listening on %s:%d", cfg.ID, cfg.Protocol, cfg.IP, cfg.Port)
		}
	}

	// Start Telemetry Engine loop
	m.wg.Add(1)
	go m.startTelemetryEngine()

	return nil
}

// StopAll halts all TCP listeners and stops the telemetry simulation loop.
func (m *WorkerManager) StopAll() {
	close(m.stopChan)

	m.mu.Lock()
	for _, w := range m.workers {
		w.Stop()
	}
	m.mu.Unlock()

	m.wg.Wait()
}

func (m *WorkerManager) startTelemetryEngine() {
	defer m.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("Telemetry Engine Simulator loop started...")

	for {
		select {
		case <-m.stopChan:
			log.Println("Telemetry Engine Simulator loop stopped.")
			return
		case <-ticker.C:
			devices := m.registry.GetAllDevices()
			for _, dev := range devices {
				for _, field := range dev.TelemetryFields {
					currentVal := dev.Telemetry[field.FieldName]
					newVal := GenerateTelemetryValue(field, currentVal)
					dev.UpdateTelemetry(field.FieldName, newVal)

					// Sync with PJLink state if the field is lamp_hours
					if field.FieldName == "lamp_hours" && dev.Protocol == "PJLink" {
						if pjDev, ok := dev.ProtocolState.(*pjlink.PJLinkDevice); ok {
							var hours int
							switch h := newVal.(type) {
							case int:
								hours = h
							case float64:
								hours = int(h)
							}
							pjDev.State().SetLampHours(hours)
						}
					}
				}
				log.Printf("[Telemetry Engine] Device: %s, Telemetry: %v", dev.ID, dev.GetTelemetry())
			}
		}
	}
}
