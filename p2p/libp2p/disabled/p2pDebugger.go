package disabled

type p2pDebugger struct {
}

// NewP2PDebugger returns a new disabled p2p debugger
func NewP2PDebugger() *p2pDebugger {
	return &p2pDebugger{}
}

// AddIncomingMessage does nothing as it is disabled
func (debugger *p2pDebugger) AddIncomingMessage(_ string, _ uint64, _ bool) {
}

// AddOutgoingMessage does nothing as it is disabled
func (debugger *p2pDebugger) AddOutgoingMessage(_ string, _ uint64, _ bool) {
}

// Close returns nil as it is disabled
func (debugger *p2pDebugger) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (debugger *p2pDebugger) IsInterfaceNil() bool {
	return debugger == nil
}
