package mock

import p2p "github.com/ElrondNetwork/elrond-go-p2p"

// ConnectionMonitorWrapperStub -
type ConnectionMonitorWrapperStub struct {
	CheckConnectionsBlockingCalled func()
	SetPeerDenialEvaluatorCalled   func(handler p2p.PeerDenialEvaluator) error
	PeerDenialEvaluatorCalled      func() p2p.PeerDenialEvaluator
}

// CheckConnectionsBlocking -
func (stub *ConnectionMonitorWrapperStub) CheckConnectionsBlocking() {
	if stub.CheckConnectionsBlockingCalled != nil {
		stub.CheckConnectionsBlockingCalled()
	}
}

// SetPeerDenialEvaluator -
func (stub *ConnectionMonitorWrapperStub) SetPeerDenialEvaluator(handler p2p.PeerDenialEvaluator) error {
	if stub.SetPeerDenialEvaluatorCalled != nil {
		return stub.SetPeerDenialEvaluatorCalled(handler)
	}
	return nil
}

// PeerDenialEvaluator -
func (stub *ConnectionMonitorWrapperStub) PeerDenialEvaluator() p2p.PeerDenialEvaluator {
	if stub.PeerDenialEvaluatorCalled != nil {
		return stub.PeerDenialEvaluatorCalled()
	}
	return nil
}

// IsInterfaceNil -
func (stub *ConnectionMonitorWrapperStub) IsInterfaceNil() bool {
	return stub == nil
}
