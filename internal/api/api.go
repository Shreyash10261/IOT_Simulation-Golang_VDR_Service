package api

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/team/vdr/internal/registry"
	"github.com/team/vdr/internal/worker"
)

// APIServer handles HTTP requests to dynamically manage virtual devices.
type APIServer struct {
	mgr *worker.WorkerManager
	reg *registry.DeviceRegistry
}

// NewAPIServer creates a new instance of APIServer.
func NewAPIServer(mgr *worker.WorkerManager, reg *registry.DeviceRegistry) *APIServer {
	return &APIServer{
		mgr: mgr,
		reg: reg,
	}
}

// RegisterRoutes registers the API routes to the given ServeMux.
func (s *APIServer) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/devices/spawn", s.handleSpawn)
	mux.HandleFunc("/devices/kill", s.handleKill)
	mux.HandleFunc("/devices", s.handleDevices)
}

// Response represents a standard JSON API response structure.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// DeviceResponse represents a JSON-serializable view of a device.
type DeviceResponse struct {
	ID           string                 `json:"id"`
	IP           string                 `json:"ip"`
	MAC          string                 `json:"mac"`
	Port         int                    `json:"port"`
	Protocol     string                 `json:"protocol"`
	Manufacturer string                 `json:"manufacturer"`
	Model        string                 `json:"model"`
	IsOnline     bool                   `json:"is_online"`
	Telemetry    map[string]interface{} `json:"telemetry"`
}

func (s *APIServer) handleSpawn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Success: false, Error: "Method not allowed"})
		return
	}

	var cfg worker.DeviceConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Error: "Invalid JSON request body: " + err.Error()})
		return
	}

	if cfg.ID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Error: "Missing required field: id"})
		return
	}
	if cfg.IP == "" {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Error: "Missing required field: ip"})
		return
	}
	if cfg.Port <= 0 {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Error: "Invalid or missing field: port"})
		return
	}

	if err := s.mgr.Spawn(cfg); err != nil {
		// If error is due to address or ID conflict, return conflict or bad request
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "in use") {
			status = http.StatusConflict
		}
		writeJSON(w, status, Response{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, Response{Success: true, Message: "Device spawned successfully"})
}

func (s *APIServer) handleKill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Success: false, Error: "Method not allowed"})
		return
	}

	var deviceID string

	// 1. Try to read from query param ?id=...
	deviceID = r.URL.Query().Get("id")

	// 2. Fall back to JSON request body if query param is empty
	if deviceID == "" && r.Body != nil {
		var body struct {
			ID string `json:"id"`
		}
		// Decode, but ignore error in case request body is empty/non-JSON and we fallback
		_ = json.NewDecoder(r.Body).Decode(&body)
		deviceID = body.ID
	}

	if deviceID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Error: "Missing device id in query parameter or request body"})
		return
	}

	if err := s.mgr.Kill(deviceID); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not running") {
			status = http.StatusNotFound
		}
		writeJSON(w, status, Response{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, Response{Success: true, Message: "Device killed successfully"})
}

func (s *APIServer) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Success: false, Error: "Method not allowed"})
		return
	}

	devices := s.reg.GetAllDevices()
	resp := make([]DeviceResponse, 0, len(devices))

	for _, d := range devices {
		macStr := net.HardwareAddr(d.MACAddress).String()
		if len(d.MACAddress) == 0 {
			macStr = ""
		}
		resp = append(resp, DeviceResponse{
			ID:           d.ID,
			IP:           d.IP,
			MAC:          macStr,
			Port:         d.Port,
			Protocol:     d.Protocol,
			Manufacturer: d.Manufacturer,
			Model:        d.Model,
			IsOnline:     d.IsOnline,
			Telemetry:    d.GetTelemetry(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
