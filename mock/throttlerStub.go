package mock

// ThrottlerStub -
type ThrottlerStub struct {
	CanProcessCalled      func() bool
	StartProcessingCalled func()
	EndProcessingCalled   func()
}

// CanProcess -
func (stub *ThrottlerStub) CanProcess() bool {
	if stub.CanProcessCalled != nil {
		return stub.CanProcessCalled()
	}
	return false
}

// StartProcessing -
func (stub *ThrottlerStub) StartProcessing() {
	if stub.StartProcessingCalled != nil {
		stub.StartProcessingCalled()
	}
}

// EndProcessing -
func (stub *ThrottlerStub) EndProcessing() {
	if stub.EndProcessingCalled != nil {
		stub.EndProcessingCalled()
	}
}

// IsInterfaceNil -
func (stub *ThrottlerStub) IsInterfaceNil() bool {
	return stub == nil
}
