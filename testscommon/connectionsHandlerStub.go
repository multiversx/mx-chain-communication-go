package testscommon

import (
	"time"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
)

// ConnectionsHandlerStub -
type ConnectionsHandlerStub struct {
	BootstrapCalled                     func() error
	PeersCalled                         func() []core.PeerID
	AddressesCalled                     func() []string
	ConnectToPeerCalled                 func(address string) error
	IsConnectedCalled                   func(peerID core.PeerID) bool
	ConnectedPeersCalled                func() []core.PeerID
	ConnectedAddressesCalled            func() []string
	PeerAddressesCalled                 func(pid core.PeerID) []string
	ConnectedPeersOnTopicCalled         func(topic string) []core.PeerID
	SetPeerShardResolverCalled          func(peerShardResolver p2p.PeerShardResolver) error
	GetConnectedPeersInfoCalled         func() *p2p.ConnectedPeersInfo
	WaitForConnectionsCalled            func(maxWaitingTime time.Duration, minNumOfPeers uint32)
	IsConnectedToTheNetworkCalled       func() bool
	ThresholdMinConnectedPeersCalled    func() int
	SetThresholdMinConnectedPeersCalled func(minConnectedPeers int) error
	SetPeerDenialEvaluatorCalled        func(handler p2p.PeerDenialEvaluator) error
	CloseCalled                         func() error
}

// Bootstrap -
func (stub *ConnectionsHandlerStub) Bootstrap() error {
	if stub.BootstrapCalled != nil {
		return stub.BootstrapCalled()
	}
	return nil
}

// Peers -
func (stub *ConnectionsHandlerStub) Peers() []core.PeerID {
	if stub.PeersCalled != nil {
		return stub.PeersCalled()
	}
	return nil
}

// Addresses -
func (stub *ConnectionsHandlerStub) Addresses() []string {
	if stub.AddressesCalled != nil {
		return stub.AddressesCalled()
	}
	return nil
}

// ConnectToPeer -
func (stub *ConnectionsHandlerStub) ConnectToPeer(address string) error {
	if stub.ConnectToPeerCalled != nil {
		return stub.ConnectToPeerCalled(address)
	}
	return nil
}

// IsConnected -
func (stub *ConnectionsHandlerStub) IsConnected(peerID core.PeerID) bool {
	if stub.IsConnectedCalled != nil {
		return stub.IsConnectedCalled(peerID)
	}
	return false
}

// ConnectedPeers -
func (stub *ConnectionsHandlerStub) ConnectedPeers() []core.PeerID {
	if stub.ConnectedPeersCalled != nil {
		return stub.ConnectedPeersCalled()
	}
	return nil
}

// ConnectedAddresses -
func (stub *ConnectionsHandlerStub) ConnectedAddresses() []string {
	if stub.ConnectedAddressesCalled != nil {
		return stub.ConnectedAddressesCalled()
	}
	return nil
}

// PeerAddresses -
func (stub *ConnectionsHandlerStub) PeerAddresses(pid core.PeerID) []string {
	if stub.PeerAddressesCalled != nil {
		return stub.PeerAddressesCalled(pid)
	}
	return nil
}

// ConnectedPeersOnTopic -
func (stub *ConnectionsHandlerStub) ConnectedPeersOnTopic(topic string) []core.PeerID {
	if stub.ConnectedPeersOnTopicCalled != nil {
		return stub.ConnectedPeersOnTopicCalled(topic)
	}
	return nil
}

// SetPeerShardResolver -
func (stub *ConnectionsHandlerStub) SetPeerShardResolver(peerShardResolver p2p.PeerShardResolver) error {
	if stub.SetPeerShardResolverCalled != nil {
		return stub.SetPeerShardResolverCalled(peerShardResolver)
	}
	return nil
}

// GetConnectedPeersInfo -
func (stub *ConnectionsHandlerStub) GetConnectedPeersInfo() *p2p.ConnectedPeersInfo {
	if stub.GetConnectedPeersInfoCalled != nil {
		return stub.GetConnectedPeersInfoCalled()
	}
	return nil
}

// WaitForConnections -
func (stub *ConnectionsHandlerStub) WaitForConnections(maxWaitingTime time.Duration, minNumOfPeers uint32) {
	if stub.WaitForConnectionsCalled != nil {
		stub.WaitForConnectionsCalled(maxWaitingTime, minNumOfPeers)
	}
}

// IsConnectedToTheNetwork -
func (stub *ConnectionsHandlerStub) IsConnectedToTheNetwork() bool {
	if stub.IsConnectedToTheNetworkCalled != nil {
		return stub.IsConnectedToTheNetworkCalled()
	}
	return false
}

// ThresholdMinConnectedPeers -
func (stub *ConnectionsHandlerStub) ThresholdMinConnectedPeers() int {
	if stub.ThresholdMinConnectedPeersCalled != nil {
		return stub.ThresholdMinConnectedPeersCalled()
	}
	return 0
}

// SetThresholdMinConnectedPeers -
func (stub *ConnectionsHandlerStub) SetThresholdMinConnectedPeers(minConnectedPeers int) error {
	if stub.SetThresholdMinConnectedPeersCalled != nil {
		return stub.SetThresholdMinConnectedPeersCalled(minConnectedPeers)
	}
	return nil
}

// SetPeerDenialEvaluator -
func (stub *ConnectionsHandlerStub) SetPeerDenialEvaluator(handler p2p.PeerDenialEvaluator) error {
	if stub.SetPeerDenialEvaluatorCalled != nil {
		return stub.SetPeerDenialEvaluatorCalled(handler)
	}
	return nil
}

// Close -
func (stub *ConnectionsHandlerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (stub *ConnectionsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
