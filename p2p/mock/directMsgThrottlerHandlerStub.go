package mock

import "github.com/multiversx/mx-chain-core-go/core"

// DirectMsgThrottlerHandlerStub -
type DirectMsgThrottlerHandlerStub struct {
	TryStartProcessingCalled func(pid core.PeerID) bool
	EndProcessingCalled      func(pid core.PeerID)
}

// TryStartProcessing -
func (stub *DirectMsgThrottlerHandlerStub) TryStartProcessing(pid core.PeerID) bool {
	if stub.TryStartProcessingCalled != nil {
		return stub.TryStartProcessingCalled(pid)
	}
	return true
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
