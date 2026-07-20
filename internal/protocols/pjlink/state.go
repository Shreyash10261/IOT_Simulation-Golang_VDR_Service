package pjlink

import (
	"sync"
)

// PowerState defines the standard PJLink power states.
type PowerState string

const (
	PowerOff     PowerState = "0"
	PowerOn      PowerState = "1"
	PowerCooling PowerState = "2"
	PowerWarming PowerState = "3"
)

// ProjectorState holds the thread-safe simulated state of the hardware.
type ProjectorState struct {
	mu         sync.RWMutex
	powerState PowerState
	lampHours  int

	// Future extension points for Sprint 3
	// inputState string
	// hasErrors  bool
}

// NewProjectorState initializes a new projector, defaulting to powered off.
func NewProjectorState() *ProjectorState {
	return &ProjectorState{
		powerState: PowerOff,
		lampHours:  0,
	}
}

// GetPower safely retrieves the current power state.
func (s *ProjectorState) GetPower() PowerState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.powerState
}

// SetPower safely updates the power state.
func (s *ProjectorState) SetPower(state PowerState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.powerState = state
}

// GetLampHours safely retrieves the current lamp hours.
func (s *ProjectorState) GetLampHours() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lampHours
}

// SetLampHours safely updates the lamp hours.
func (s *ProjectorState) SetLampHours(hours int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lampHours = hours
}
