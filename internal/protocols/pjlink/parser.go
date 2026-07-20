package pjlink

import (
	"strings"
)

// ParseRequest takes raw bytes from the TCP socket and returns a structured Request.
// The PJLink Class 1 specification mandates the format: "%1COMMAND PARAMETER\r"
//
// Note (Sprint 2): This implementation assumes Authentication Disabled mode.
// MD5 Challenge-Response authentication headers are not currently supported.
func ParseRequest(raw []byte) (*Request, error) {
	// 1. Strip trailing carriage return or newlines
	msg := strings.TrimRight(string(raw), "\r\n")

	// 2. Validate prefix "%1" (Class 1 PJLink) or "%2" (Class 2)
	if len(msg) < 6 || !strings.HasPrefix(msg, "%") {
		return nil, ErrUndefinedCommand
	}

	classStr := msg[1:2]
	class := CommandClass(classStr)

	if class != Class1 && class != Class2 {
		return nil, ErrUndefinedCommand
	}

	// 3. Extract Command (4 characters)
	// Example: "%1POWR 1" -> msg[2:6] is "POWR"
	command := msg[2:6]

	// 4. Validate the space separator
	if msg[6] != ' ' {
		return nil, ErrUndefinedCommand
	}

	// 5. Extract Parameter
	// Example: "%1POWR 1" -> msg[7:] is "1"
	parameter := msg[7:]

	return &Request{
		Class:     class,
		Command:   command,
		Parameter: parameter,
	}, nil
}
