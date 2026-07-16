package pjlink

import "fmt"

// PJLinkError represents a protocol-level error that must be returned to the client.
type PJLinkError struct {
	Code    string
	Message string
}

func (e *PJLinkError) Error() string {
	return fmt.Sprintf("PJLink Error %s: %s", e.Code, e.Message)
}

var (
	// ErrUndefinedCommand (ERR1) is returned when the command is not recognized (e.g. %1FAKE).
	ErrUndefinedCommand = &PJLinkError{Code: "ERR1", Message: "Undefined command"}

	// ErrOutOfParameter (ERR2) is returned when the parameter is out of bounds (e.g. %1POWR 9).
	ErrOutOfParameter = &PJLinkError{Code: "ERR2", Message: "Out of parameter"}

	// ErrUnavailableTime (ERR3) is returned when the projector is busy (e.g. warming up/cooling down).
	ErrUnavailableTime = &PJLinkError{Code: "ERR3", Message: "Unavailable time"}

	// ErrProjectorFailure (ERR4) is returned when the hardware is in an error state (e.g. Lamp dead).
	ErrProjectorFailure = &PJLinkError{Code: "ERR4", Message: "Projector/Display failure"}
)
