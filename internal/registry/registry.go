package registry

import (
	"sync"

	"github.com/team/vdr/internal/models"
)

// Device represents a simulated Projector or IoT device.
type Device struct {
	ID              string
	IP              string
	MACAddress      []byte
	Port            int
	Protocol        string
	Manufacturer    string
	Model           string
	IsOnline        bool
	TelemetryFields []models.TelemetryField
	ProtocolState   interface{} // stores stateful protocol handlers, e.g. *pjlink.PJLinkDevice

	mu        sync.RWMutex
	Telemetry map[string]interface{}
}

// UpdateTelemetry updates a telemetry field in a thread-safe manner.
func (d *Device) UpdateTelemetry(field string, value interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.Telemetry == nil {
		d.Telemetry = make(map[string]interface{})
	}
	d.Telemetry[field] = value
}

// GetTelemetry returns a copy of the telemetry map.
func (d *Device) GetTelemetry() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()
	res := make(map[string]interface{}, len(d.Telemetry))
	for k, v := range d.Telemetry {
		res[k] = v
	}
	return res
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

// GetDeviceByID performs a lookup for a device by its ID.
func (r *DeviceRegistry) GetDeviceByID(id string) (*Device, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, dev := range r.devices {
		if dev.ID == id {
			return dev, true
		}
	}
	return nil, false
}

// GetAllDevices returns a slice of all registered devices.
func (r *DeviceRegistry) GetAllDevices() []*Device {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*Device, 0, len(r.devices))
	for _, dev := range r.devices {
		list = append(list, dev)
	}
	return list
}
