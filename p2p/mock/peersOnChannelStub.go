package mock

import "github.com/multiversx/mx-chain-core-go/core"

// PeersOnChannelStub -
type PeersOnChannelStub struct {
	ConnectedPeersOnChannelCalled func(topic string) []core.PeerID
	CloseCalled                   func() error
}

// ConnectedPeersOnChannel -
func (stub *PeersOnChannelStub) ConnectedPeersOnChannel(topic string) []core.PeerID {
	if stub.ConnectedPeersOnChannelCalled != nil {
		return stub.ConnectedPeersOnChannelCalled(topic)
	}
	return nil
}

// Close -
func (stub *PeersOnChannelStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (stub *PeersOnChannelStub) IsInterfaceNil() bool {
	return stub == nil
}
