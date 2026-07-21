package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/team/vdr/internal/models"
	"github.com/team/vdr/internal/registry"
	"github.com/team/vdr/internal/worker"
)

func TestAPIServerFlow(t *testing.T) {
	// Initialize registry and manager
	reg := registry.NewDeviceRegistry()
	mgr := worker.NewWorkerManager(reg)
	defer mgr.StopAll()

	// Initialize API server
	apiServer := NewAPIServer(mgr, reg)
	mux := http.NewServeMux()
	apiServer.RegisterRoutes(mux)

	// 1. Initially, GET /devices should return empty array
	req, _ := http.NewRequest(http.MethodGet, "/devices", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var devices []DeviceResponse
	if err := json.NewDecoder(rr.Body).Decode(&devices); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}

	// 2. POST /devices/spawn to add a device
	spawnPayload := worker.DeviceConfig{
		ID:           "DynamicProj01",
		IP:           "127.0.0.1",
		MAC:          "02:00:00:00:00:aa",
		Port:         14360,
		Protocol:     "PJLink",
		Manufacturer: "DynamicVendor",
		Model:        "DynamicModel",
		Telemetry: []models.TelemetryField{
			{
				FieldName: "lamp_hours",
				DataType:  "int",
				Unit:      "hours",
			},
		},
	}
	bodyBytes, _ := json.Marshal(spawnPayload)
	req, _ = http.NewRequest(http.MethodPost, "/devices/spawn", bytes.NewBuffer(bodyBytes))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var resp Response
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if !resp.Success {
		t.Errorf("expected success: true, got %v", resp.Success)
	}

	// 3. GET /devices should now return 1 device
	req, _ = http.NewRequest(http.MethodGet, "/devices", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	_ = json.NewDecoder(rr.Body).Decode(&devices)
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0].ID != "DynamicProj01" || devices[0].IP != "127.0.0.1" || devices[0].Port != 14360 {
		t.Errorf("device details mismatch: %+v", devices[0])
	}

	// 4. POST /devices/spawn with duplicate ID should fail
	req, _ = http.NewRequest(http.MethodPost, "/devices/spawn", bytes.NewBuffer(bodyBytes))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected status 409 conflict, got %d", rr.Code)
	}

	// 5. DELETE /devices/kill using query param
	req, _ = http.NewRequest(http.MethodDelete, "/devices/kill?id=DynamicProj01", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	// 6. GET /devices should be empty again
	req, _ = http.NewRequest(http.MethodGet, "/devices", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	_ = json.NewDecoder(rr.Body).Decode(&devices)
	if len(devices) != 0 {
		t.Errorf("expected 0 devices after kill, got %d", len(devices))
	}

	// 7. DELETE /devices/kill of non-existing device should return 404
	req, _ = http.NewRequest(http.MethodDelete, "/devices/kill?id=DynamicProj01", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}

	// 8. DELETE /devices/kill using request body
	// Spawn first again
	req, _ = http.NewRequest(http.MethodPost, "/devices/spawn", bytes.NewBuffer(bodyBytes))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("failed to re-spawn device: %d", rr.Code)
	}

	killBody := struct {
		ID string `json:"id"`
	}{ID: "DynamicProj01"}
	killBodyBytes, _ := json.Marshal(killBody)

	req, _ = http.NewRequest(http.MethodDelete, "/devices/kill", bytes.NewBuffer(killBodyBytes))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 for body-based kill, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}
