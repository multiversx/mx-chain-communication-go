package mock

// DebuggerStub -
type DebuggerStub struct {
	AddIncomingMessageCalled func(topic string, size uint64, isRejected bool)
	AddOutgoingMessageCalled func(topic string, size uint64, isRejected bool)
	CloseCalled              func() error
}

// AddIncomingMessage -
func (stub *DebuggerStub) AddIncomingMessage(topic string, size uint64, isRejected bool) {
	if stub.AddIncomingMessageCalled != nil {
		stub.AddIncomingMessageCalled(topic, size, isRejected)
	}
}

// AddOutgoingMessage -
func (stub *DebuggerStub) AddOutgoingMessage(topic string, size uint64, isRejected bool) {
	if stub.AddOutgoingMessageCalled != nil {
		stub.AddOutgoingMessageCalled(topic, size, isRejected)
	}
}

// Close -
func (stub *DebuggerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (stub *DebuggerStub) IsInterfaceNil() bool {
	return stub == nil
}
