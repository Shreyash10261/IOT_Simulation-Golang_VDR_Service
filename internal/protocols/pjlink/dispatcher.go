package pjlink

// DispatchCommand routes the parsed request to the appropriate command handler.
func DispatchCommand(req *Request, state *ProjectorState) (*Response, error) {
	switch req.Command {
	case "POWR":
		return HandlePOWR(req, state)
	case "LAMP":
		return HandleLAMP(req, state)

	// Future extension points for Sprint 3
	/*
		case "INPT":
			return HandleINPT(req, state)
		case "AVMT":
			return HandleAVMT(req, state)
		case "ERST":
			return HandleERST(req, state)
		case "NAME":
			return HandleNAME(req, state)
	*/

	default:
		// If the command is not in our switch statement, it is undefined
		return nil, ErrUndefinedCommand
	}
}
