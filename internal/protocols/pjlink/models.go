package pjlink

// CommandClass represents the PJLink protocol class (e.g., 1 or 2).
type CommandClass string

const (
	Class1 CommandClass = "1"
	Class2 CommandClass = "2"
)

// Request represents a parsed PJLink request from the TCP socket.
type Request struct {
	Class     CommandClass // Typically '1'
	Command   string       // e.g., "POWR"
	Parameter string       // e.g., "1", "0", or "?" for status query
}

// Response represents a structured PJLink response before formatting to bytes.
type Response struct {
	Class   CommandClass
	Command string
	Payload string // e.g., "OK", "ERR1", or "1"
}
