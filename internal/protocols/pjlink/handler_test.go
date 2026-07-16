package pjlink

import (
	"bytes"
	"testing"
)

func TestPJLinkDevice_HandlePOWR(t *testing.T) {
	device := NewPJLinkDevice()

	tests := []struct {
		name           string
		requestBytes   []byte
		expectedBytes  []byte
		setupStateFunc func(s *ProjectorState)
	}{
		{
			name:          "Status Query - Default Off",
			requestBytes:  []byte("%1POWR ?\r"),
			expectedBytes: []byte("%1POWR=0\r"),
		},
		{
			name:          "Turn On",
			requestBytes:  []byte("%1POWR 1\r"),
			expectedBytes: []byte("%1POWR=OK\r"),
		},
		{
			name:          "Status Query - Now On",
			requestBytes:  []byte("%1POWR ?\r"),
			expectedBytes: []byte("%1POWR=1\r"),
			setupStateFunc: func(s *ProjectorState) {
				s.SetPower(PowerOn)
			},
		},
		{
			name:          "Turn Off",
			requestBytes:  []byte("%1POWR 0\r"),
			expectedBytes: []byte("%1POWR=OK\r"),
		},
		{
			name:          "Error - Out of Parameter",
			requestBytes:  []byte("%1POWR 9\r"),
			expectedBytes: []byte("%1POWR=ERR2\r"),
		},
		{
			name:          "Error - Unavailable Time (Cooling)",
			requestBytes:  []byte("%1POWR 0\r"),
			expectedBytes: []byte("%1POWR=ERR3\r"),
			setupStateFunc: func(s *ProjectorState) {
				s.SetPower(PowerCooling)
			},
		},
		{
			name:          "Error - Undefined Command",
			requestBytes:  []byte("%1FAKE 1\r"),
			expectedBytes: []byte("%1FAKE=ERR1\r"),
		},
		{
			name:          "Error - Malformed Prefix",
			requestBytes:  []byte("INVALID_BYTES\r"),
			expectedBytes: []byte("%1ERR=ERR1\r"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup specific state if required by the test
			if tt.setupStateFunc != nil {
				tt.setupStateFunc(device.state)
			}

			// Execute the public handler
			actualBytes := device.Handle(tt.requestBytes)

			// Assert
			if !bytes.Equal(actualBytes, tt.expectedBytes) {
				t.Errorf("Handle() = %q, want %q", actualBytes, tt.expectedBytes)
			}
		})
	}
}
