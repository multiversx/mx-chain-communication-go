package mock

import "github.com/multiversx/mx-chain-core-go/core"

// ConnectionsProviderStub -
type ConnectionsProviderStub struct {
	ConnectedPeersCalled func() []core.PeerID
}

// ConnectedPeers -
func (stub *ConnectionsProviderStub) ConnectedPeers() []core.PeerID {
	if stub.ConnectedPeersCalled != nil {
		return stub.ConnectedPeersCalled()
	}
	return make([]core.PeerID, 0)
}

// IsInterfaceNil -
func (stub *ConnectionsProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
