package worker

import (
	"encoding/json"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/team/vdr/internal/models"
	"github.com/team/vdr/internal/registry"
)

func TestWorkerAndTelemetry(t *testing.T) {
	// Create a temporary configuration file
	tempFile, err := os.CreateTemp("", "devices_test_*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testConfigs := []DeviceConfig{
		{
			ID:           "TestProj01",
			IP:           "127.0.0.1",
			MAC:          "02:00:00:00:00:99",
			Port:         14352, // Using high port for local testing
			Protocol:     "PJLink",
			Manufacturer: "TestVendor",
			Model:        "TestModel",
			Telemetry: []models.TelemetryField{
				{
					FieldName: "lamp_hours",
					DataType:  "int",
					Unit:      "hours",
				},
				{
					FieldName: "temperature",
					DataType:  "float",
					Unit:      "celsius",
				},
			},
		},
	}

	bytes, _ := json.Marshal(testConfigs)
	_, _ = tempFile.Write(bytes)
	_ = tempFile.Close()

	// Initialize Registry and Manager
	reg := registry.NewDeviceRegistry()
	mgr := NewWorkerManager(reg)

	// Load and spawn
	err = mgr.LoadAndSpawn(tempFile.Name())
	if err != nil {
		t.Fatalf("failed to LoadAndSpawn: %v", err)
	}
	defer mgr.StopAll()

	// 1. Verify device registration
	dev, exists := reg.GetDeviceByID("TestProj01")
	if !exists {
		t.Fatal("device was not registered in registry")
	}

	// 2. Connect to the TCP listener and test PJLink exchange
	conn, err := net.DialTimeout("tcp", "127.0.0.1:14352", 2*time.Second)
	if err != nil {
		t.Fatalf("failed to connect to device TCP server: %v", err)
	}
	defer conn.Close()

	// Read greeting
	greetingBuf := make([]byte, 10)
	n, err := conn.Read(greetingBuf)
	if err != nil {
		t.Fatalf("failed to read greeting: %v", err)
	}
	greeting := string(greetingBuf[:n])
	if greeting != "PJLINK 0\r" {
		t.Errorf("unexpected greeting: %q", greeting)
	}

	// Send POWR query
	_, err = conn.Write([]byte("%1POWR ?\r"))
	if err != nil {
		t.Fatalf("failed to write command: %v", err)
	}

	// Read response
	respBuf := make([]byte, 1024)
	n, err = conn.Read(respBuf)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	resp := string(respBuf[:n])
	if !strings.HasPrefix(resp, "%1POWR=") {
		t.Errorf("unexpected command response: %q", resp)
	}

	// 3. Test telemetry generation values (wait for at least one simulation tick)
	t.Log("Waiting for Telemetry Engine tick...")
	time.Sleep(6 * time.Second)

	telemetry := dev.GetTelemetry()
	t.Logf("Generated telemetry values: %v", telemetry)

	lampHours, ok := telemetry["lamp_hours"]
	if !ok {
		t.Error("lamp_hours telemetry field was not populated")
	} else {
		switch v := lampHours.(type) {
		case int:
			if v < 100 {
				t.Errorf("expected lamp_hours to start at >= 100, got %d", v)
			}
		case float64:
			if v < 100.0 {
				t.Errorf("expected lamp_hours to start at >= 100.0, got %f", v)
			}
		default:
			t.Errorf("unexpected type for lamp_hours: %T", lampHours)
		}
	}

	temp, ok := telemetry["temperature"]
	if !ok {
		t.Error("temperature telemetry field was not populated")
	} else {
		switch v := temp.(type) {
		case float64:
			if v < 30.0 || v > 80.0 {
				t.Errorf("temperature value %f out of expected range [30, 80]", v)
			}
		default:
			t.Errorf("unexpected type for temperature: %T", temp)
		}
	}
}
