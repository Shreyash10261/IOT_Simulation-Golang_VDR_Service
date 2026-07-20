package pjlink

// PJLinkDevice represents the public entrypoint for the PJLink protocol simulator.
type PJLinkDevice struct {
	state *ProjectorState
}

// NewPJLinkDevice initializes a new simulated projector.
func NewPJLinkDevice() *PJLinkDevice {
	return &PJLinkDevice{
		state: NewProjectorState(),
	}
}

// State returns the underlying projector state.
func (d *PJLinkDevice) State() *ProjectorState {
	return d.state
}

// Handle is the core public method. It accepts raw TCP bytes from a network socket,
// processes the command against the simulated state, and returns the exact bytes to write back to the socket.
func (d *PJLinkDevice) Handle(requestBytes []byte) []byte {
	// 1. Parse the incoming bytes
	req, err := ParseRequest(requestBytes)
	if err != nil {
		// If parsing fails entirely, check if we at least have a PJLinkError
		if _, ok := err.(*PJLinkError); ok {
			// We can't determine the class/command if parsing completely failed,
			// but if we managed to extract them before failing, we would use them.
			// For simplicity in this handler, a raw parsing failure returns a generic error.
			return FormatGenericError()
		}
		return FormatGenericError()
	}

	// 2. Dispatch to the specific command logic
	resp, err := DispatchCommand(req, d.state)
	if err != nil {
		if pjErr, ok := err.(*PJLinkError); ok {
			return FormatErrorResponse(req.Class, req.Command, pjErr)
		}
		// Fallback for unknown internal errors
		return FormatErrorResponse(req.Class, req.Command, ErrProjectorFailure)
	}

	// 3. Format the successful response
	return FormatResponse(resp)
}
