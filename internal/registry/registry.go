package registry

import (
	"sync"
)

// Device represents a simulated Projector or IoT device.
type Device struct {
	IP         string
	MACAddress []byte
	IsOnline   bool
	// PJLink specific properties can be added here later (Task 175)
}

// DeviceRegistry holds all simulated devices in memory with O(1) lookup.
type DeviceRegistry struct {
	mu      sync.RWMutex
	devices map[string]*Device
}

// NewDeviceRegistry initializes a new empty registry.
func NewDeviceRegistry() *DeviceRegistry {
	return &DeviceRegistry{
		devices: make(map[string]*Device),
	}
}

// RegisterDevice adds a new simulated device to the registry.
func (r *DeviceRegistry) RegisterDevice(dev *Device) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices[dev.IP] = dev
}

// GetDeviceByIP performs an O(1) lookup for a device by its IP address.
func (r *DeviceRegistry) GetDeviceByIP(ip string) (*Device, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	dev, exists := r.devices[ip]
	return dev, exists
}
