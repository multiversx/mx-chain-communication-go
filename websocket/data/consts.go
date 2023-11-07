package data

const (
	// ClosedConnectionMessage is the message that is received when try to send a message over a closed WebSocket connection
	ClosedConnectionMessage = "use of closed network connection"
	// ResetByPeerConnectionMessage is the message that is received from server is case of a connection with problems
	ResetByPeerConnectionMessage = "connection reset by peer"
	// BrokenPipeMessage is the error message that is received in case of connection with problems
	BrokenPipeMessage = "broken pipe"
)
