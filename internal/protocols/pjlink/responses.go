package pjlink

import (
	"fmt"
)

// FormatResponse converts a successful Response object into the exact PJLink byte specification.
func FormatResponse(resp *Response) []byte {
	// Standard response format: "%[Class][Command]=[Payload]\r"
	// Example: "%1POWR=OK\r" or "%1POWR=1\r"
	str := fmt.Sprintf("%%%s%s=%s\r", resp.Class, resp.Command, resp.Payload)
	return []byte(str)
}

// FormatErrorResponse converts a PJLinkError into the exact byte specification.
func FormatErrorResponse(class CommandClass, command string, err *PJLinkError) []byte {
	// Error response format: "%[Class][Command]=[ErrorCode]\r"
	// Example: "%1POWR=ERR1\r"
	str := fmt.Sprintf("%%%s%s=%s\r", class, command, err.Code)
	return []byte(str)
}

// FormatGenericError handles edge cases where parsing failed entirely (e.g., malformed prefix).
func FormatGenericError() []byte {
	// If we can't even extract the command, return a generic Class 1 undefined command error
	return []byte("%1ERR=ERR1\r")
}
