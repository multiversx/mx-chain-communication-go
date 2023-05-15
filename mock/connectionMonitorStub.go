package mock

import (
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
	p2p "github.com/multiversx/mx-chain-p2p-go"
)

// ConnectionMonitorStub -
type ConnectionMonitorStub struct {
	ListenCalled                        func(netw network.Network, ma multiaddr.Multiaddr)
	ListenCloseCalled                   func(netw network.Network, ma multiaddr.Multiaddr)
	ConnectedCalled                     func(netw network.Network, conn network.Conn)
	DisconnectedCalled                  func(netw network.Network, conn network.Conn)
	IsConnectedToTheNetworkCalled       func(netw network.Network) bool
	SetThresholdMinConnectedPeersCalled func(thresholdMinConnectedPeers int, netw network.Network)
	ThresholdMinConnectedPeersCalled    func() int
	SetPeerDenialEvaluatorCalled        func(handler p2p.PeerDenialEvaluator) error
	PeerDenialEvaluatorCalled           func() p2p.PeerDenialEvaluator
	CloseCalled                         func() error
}

// Listen -
func (cms *ConnectionMonitorStub) Listen(netw network.Network, ma multiaddr.Multiaddr) {
	if cms.ListenCalled != nil {
		cms.ListenCalled(netw, ma)
	}
}

// ListenClose -
func (cms *ConnectionMonitorStub) ListenClose(netw network.Network, ma multiaddr.Multiaddr) {
	if cms.ListenCloseCalled != nil {
		cms.ListenCloseCalled(netw, ma)
	}
}

// Connected -
func (cms *ConnectionMonitorStub) Connected(netw network.Network, conn network.Conn) {
	if cms.ConnectedCalled != nil {
		cms.ConnectedCalled(netw, conn)
	}
}

// Disconnected -
func (cms *ConnectionMonitorStub) Disconnected(netw network.Network, conn network.Conn) {
	if cms.DisconnectedCalled != nil {
		cms.DisconnectedCalled(netw, conn)
	}
}

// IsConnectedToTheNetwork -
func (cms *ConnectionMonitorStub) IsConnectedToTheNetwork(netw network.Network) bool {
	if cms.IsConnectedToTheNetworkCalled != nil {
		return cms.IsConnectedToTheNetworkCalled(netw)
	}

	return false
}

// SetThresholdMinConnectedPeers -
func (cms *ConnectionMonitorStub) SetThresholdMinConnectedPeers(thresholdMinConnectedPeers int, netw network.Network) {
	if cms.SetThresholdMinConnectedPeersCalled != nil {
		cms.SetThresholdMinConnectedPeersCalled(thresholdMinConnectedPeers, netw)
	}
}

// ThresholdMinConnectedPeers -
func (cms *ConnectionMonitorStub) ThresholdMinConnectedPeers() int {
	if cms.ThresholdMinConnectedPeersCalled != nil {
		return cms.ThresholdMinConnectedPeersCalled()
	}

	return 0
}

// SetPeerDenialEvaluator -
func (cms *ConnectionMonitorStub) SetPeerDenialEvaluator(handler p2p.PeerDenialEvaluator) error {
	if cms.SetPeerDenialEvaluatorCalled != nil {
		return cms.SetPeerDenialEvaluatorCalled(handler)
	}
	return nil
}

// PeerDenialEvaluator -
func (cms *ConnectionMonitorStub) PeerDenialEvaluator() p2p.PeerDenialEvaluator {
	if cms.PeerDenialEvaluatorCalled != nil {
		return cms.PeerDenialEvaluatorCalled()
	}
	return nil
}

// Close -
func (cms *ConnectionMonitorStub) Close() error {
	if cms.CloseCalled != nil {
		return cms.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (cms *ConnectionMonitorStub) IsInterfaceNil() bool {
	return cms == nil
}
