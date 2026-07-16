package pjlink

// HandlePOWR processes the PJLink Power command ("POWR")
func HandlePOWR(req *Request, state *ProjectorState) (*Response, error) {
	// Status Query
	if req.Parameter == "?" {
		return &Response{
			Class:   req.Class,
			Command: req.Command,
			Payload: string(state.GetPower()),
		}, nil
	}

	// Turn OFF
	if req.Parameter == "0" {
		// Example of ErrUnavailableTime:
		// If the projector is already cooling down (state "2"), it can't accept a new power command.
		if state.GetPower() == PowerCooling {
			return nil, ErrUnavailableTime
		}
		
		state.SetPower(PowerOff)
		return &Response{
			Class:   req.Class,
			Command: req.Command,
			Payload: "OK",
		}, nil
	}

	// Turn ON
	if req.Parameter == "1" {
		// Example of ErrUnavailableTime:
		// If the projector is currently warming up (state "3"), it shouldn't accept a new command until done.
		if state.GetPower() == PowerWarming {
			return nil, ErrUnavailableTime
		}

		state.SetPower(PowerOn)
		return &Response{
			Class:   req.Class,
			Command: req.Command,
			Payload: "OK",
		}, nil
	}

	// If the parameter is anything else (e.g., "9"), it violates the PJLink spec for POWR
	return nil, ErrOutOfParameter
}
