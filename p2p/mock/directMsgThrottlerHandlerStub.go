package mock

import "github.com/multiversx/mx-chain-core-go/core"

// DirectMsgThrottlerHandlerStub -
type DirectMsgThrottlerHandlerStub struct {
	CanProcessCalled      func(pid core.PeerID) bool
	StartProcessingCalled func(pid core.PeerID)
	EndProcessingCalled   func(pid core.PeerID)
}

// CanProcess -
func (stub *DirectMsgThrottlerHandlerStub) CanProcess(pid core.PeerID) bool {
	if stub.CanProcessCalled != nil {
		return stub.CanProcessCalled(pid)
	}
	return true
}

// StartProcessing -
func (stub *DirectMsgThrottlerHandlerStub) StartProcessing(pid core.PeerID) {
	if stub.StartProcessingCalled != nil {
		stub.StartProcessingCalled(pid)
	}
}

// EndProcessing -
func (stub *DirectMsgThrottlerHandlerStub) EndProcessing(pid core.PeerID) {
	if stub.EndProcessingCalled != nil {
		stub.EndProcessingCalled(pid)
	}
}

// IsInterfaceNil -
func (stub *DirectMsgThrottlerHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
