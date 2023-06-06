package mock

import (
	"time"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
)

// MessengerStub -
type MessengerStub struct {
	ConnectedFullHistoryPeersOnTopicCalled  func(topic string) []core.PeerID
	IDCalled                                func() core.PeerID
	CloseCalled                             func() error
	CreateTopicCalled                       func(name string, createChannelForTopic bool) error
	HasTopicCalled                          func(name string) bool
	HasTopicValidatorCalled                 func(name string) bool
	BroadcastOnChannelCalled                func(channel string, topic string, buff []byte)
	BroadcastCalled                         func(topic string, buff []byte)
	RegisterMessageProcessorCalled          func(topic string, identifier string, handler p2p.MessageProcessor) error
	BootstrapCalled                         func() error
	PeerAddressesCalled                     func(pid core.PeerID) []string
	IsConnectedToTheNetworkCalled           func() bool
	PeersCalled                             func() []core.PeerID
	AddressesCalled                         func() []string
	ConnectToPeerCalled                     func(address string) error
	IsConnectedCalled                       func(peerID core.PeerID) bool
	ConnectedPeersCalled                    func() []core.PeerID
	ConnectedAddressesCalled                func() []string
	ConnectedPeersOnTopicCalled             func(topic string) []core.PeerID
	UnregisterAllMessageProcessorsCalled    func() error
	UnregisterMessageProcessorCalled        func(topic string, identifier string) error
	SendToConnectedPeerCalled               func(topic string, buff []byte, peerID core.PeerID) error
	ThresholdMinConnectedPeersCalled        func() int
	SetThresholdMinConnectedPeersCalled     func(minConnectedPeers int) error
	SetPeerShardResolverCalled              func(peerShardResolver p2p.PeerShardResolver) error
	SetPeerDenialEvaluatorCalled            func(handler p2p.PeerDenialEvaluator) error
	GetConnectedPeersInfoCalled             func() *p2p.ConnectedPeersInfo
	UnJoinAllTopicsCalled                   func() error
	PortCalled                              func() int
	WaitForConnectionsCalled                func(maxWaitingTime time.Duration, minNumOfPeers uint32)
	SignCalled                              func(payload []byte) ([]byte, error)
	VerifyCalled                            func(payload []byte, pid core.PeerID, signature []byte) error
	AddPeerTopicNotifierCalled              func(notifier p2p.PeerTopicNotifier) error
	BroadcastUsingPrivateKeyCalled          func(topic string, buff []byte, pid core.PeerID, skBytes []byte)
	BroadcastOnChannelUsingPrivateKeyCalled func(channel string, topic string, buff []byte, pid core.PeerID, skBytes []byte)
	SignUsingPrivateKeyCalled               func(skBytes []byte, payload []byte) ([]byte, error)
	TypeCalled                              func() p2p.NetworkMessengerType
}

// ConnectedFullHistoryPeersOnTopic -
func (stub *MessengerStub) ConnectedFullHistoryPeersOnTopic(topic string) []core.PeerID {
	if stub.ConnectedFullHistoryPeersOnTopicCalled != nil {
		return stub.ConnectedFullHistoryPeersOnTopicCalled(topic)
	}

	return make([]core.PeerID, 0)
}

// ID -
func (stub *MessengerStub) ID() core.PeerID {
	if stub.IDCalled != nil {
		return stub.IDCalled()
	}

	return "peer ID"
}

// RegisterMessageProcessor -
func (stub *MessengerStub) RegisterMessageProcessor(topic string, identifier string, handler p2p.MessageProcessor) error {
	if stub.RegisterMessageProcessorCalled != nil {
		return stub.RegisterMessageProcessorCalled(topic, identifier, handler)
	}

	return nil
}

// Broadcast -
func (stub *MessengerStub) Broadcast(topic string, buff []byte) {
	if stub.BroadcastCalled != nil {
		stub.BroadcastCalled(topic, buff)
	}
}

// Close -
func (stub *MessengerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// CreateTopic -
func (stub *MessengerStub) CreateTopic(name string, createChannelForTopic bool) error {
	if stub.CreateTopicCalled != nil {
		return stub.CreateTopicCalled(name, createChannelForTopic)
	}

	return nil
}

// HasTopic -
func (stub *MessengerStub) HasTopic(name string) bool {
	if stub.HasTopicCalled != nil {
		return stub.HasTopicCalled(name)
	}

	return false
}

// HasTopicValidator -
func (stub *MessengerStub) HasTopicValidator(name string) bool {
	if stub.HasTopicValidatorCalled != nil {
		return stub.HasTopicValidatorCalled(name)
	}

	return false
}

// BroadcastOnChannel -
func (stub *MessengerStub) BroadcastOnChannel(channel string, topic string, buff []byte) {
	if stub.BroadcastOnChannelCalled != nil {
		stub.BroadcastOnChannelCalled(channel, topic, buff)
	}
}

// Bootstrap -
func (stub *MessengerStub) Bootstrap() error {
	if stub.BootstrapCalled != nil {
		return stub.BootstrapCalled()
	}

	return nil
}

// PeerAddresses -
func (stub *MessengerStub) PeerAddresses(pid core.PeerID) []string {
	if stub.PeerAddressesCalled != nil {
		return stub.PeerAddressesCalled(pid)
	}

	return make([]string, 0)
}

// IsConnectedToTheNetwork -
func (stub *MessengerStub) IsConnectedToTheNetwork() bool {
	if stub.IsConnectedToTheNetworkCalled != nil {
		return stub.IsConnectedToTheNetworkCalled()
	}

	return true
}

// Peers -
func (stub *MessengerStub) Peers() []core.PeerID {
	if stub.PeersCalled != nil {
		return stub.PeersCalled()
	}

	return make([]core.PeerID, 0)
}

// Addresses -
func (stub *MessengerStub) Addresses() []string {
	if stub.AddressesCalled != nil {
		return stub.AddressesCalled()
	}

	return make([]string, 0)
}

// ConnectToPeer -
func (stub *MessengerStub) ConnectToPeer(address string) error {
	if stub.ConnectToPeerCalled != nil {
		return stub.ConnectToPeerCalled(address)
	}

	return nil
}

// IsConnected -
func (stub *MessengerStub) IsConnected(peerID core.PeerID) bool {
	if stub.IsConnectedCalled != nil {
		return stub.IsConnectedCalled(peerID)
	}

	return false
}

// ConnectedPeers -
func (stub *MessengerStub) ConnectedPeers() []core.PeerID {
	if stub.ConnectedPeersCalled != nil {
		return stub.ConnectedPeersCalled()
	}

	return make([]core.PeerID, 0)
}

// ConnectedAddresses -
func (stub *MessengerStub) ConnectedAddresses() []string {
	if stub.ConnectedAddressesCalled != nil {
		return stub.ConnectedAddressesCalled()
	}

	return make([]string, 0)
}

// ConnectedPeersOnTopic -
func (stub *MessengerStub) ConnectedPeersOnTopic(topic string) []core.PeerID {
	if stub.ConnectedPeersOnTopicCalled != nil {
		return stub.ConnectedPeersOnTopicCalled(topic)
	}

	return make([]core.PeerID, 0)
}

// UnregisterAllMessageProcessors -
func (stub *MessengerStub) UnregisterAllMessageProcessors() error {
	if stub.UnregisterAllMessageProcessorsCalled != nil {
		return stub.UnregisterAllMessageProcessorsCalled()
	}

	return nil
}

// UnregisterMessageProcessor -
func (stub *MessengerStub) UnregisterMessageProcessor(topic string, identifier string) error {
	if stub.UnregisterMessageProcessorCalled != nil {
		return stub.UnregisterMessageProcessorCalled(topic, identifier)
	}

	return nil
}

// SendToConnectedPeer -
func (stub *MessengerStub) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	if stub.SendToConnectedPeerCalled != nil {
		return stub.SendToConnectedPeerCalled(topic, buff, peerID)
	}

	return nil
}

// ThresholdMinConnectedPeers -
func (stub *MessengerStub) ThresholdMinConnectedPeers() int {
	if stub.ThresholdMinConnectedPeersCalled != nil {
		return stub.ThresholdMinConnectedPeersCalled()
	}

	return 0
}

// SetThresholdMinConnectedPeers -
func (stub *MessengerStub) SetThresholdMinConnectedPeers(minConnectedPeers int) error {
	if stub.SetThresholdMinConnectedPeersCalled != nil {
		return stub.SetThresholdMinConnectedPeersCalled(minConnectedPeers)
	}

	return nil
}

// SetPeerShardResolver -
func (stub *MessengerStub) SetPeerShardResolver(peerShardResolver p2p.PeerShardResolver) error {
	if stub.SetPeerShardResolverCalled != nil {
		return stub.SetPeerShardResolverCalled(peerShardResolver)
	}

	return nil
}

// SetPeerDenialEvaluator -
func (stub *MessengerStub) SetPeerDenialEvaluator(handler p2p.PeerDenialEvaluator) error {
	if stub.SetPeerDenialEvaluatorCalled != nil {
		return stub.SetPeerDenialEvaluatorCalled(handler)
	}

	return nil
}

// GetConnectedPeersInfo -
func (stub *MessengerStub) GetConnectedPeersInfo() *p2p.ConnectedPeersInfo {
	if stub.GetConnectedPeersInfoCalled != nil {
		return stub.GetConnectedPeersInfoCalled()
	}

	return nil
}

// UnJoinAllTopics -
func (stub *MessengerStub) UnJoinAllTopics() error {
	if stub.UnJoinAllTopicsCalled != nil {
		return stub.UnJoinAllTopicsCalled()
	}

	return nil
}

// Port -
func (stub *MessengerStub) Port() int {
	if stub.PortCalled != nil {
		return stub.PortCalled()
	}

	return 0
}

// WaitForConnections -
func (stub *MessengerStub) WaitForConnections(maxWaitingTime time.Duration, minNumOfPeers uint32) {
	if stub.WaitForConnectionsCalled != nil {
		stub.WaitForConnectionsCalled(maxWaitingTime, minNumOfPeers)
	}
}

// Sign -
func (stub *MessengerStub) Sign(payload []byte) ([]byte, error) {
	if stub.SignCalled != nil {
		return stub.SignCalled(payload)
	}

	return make([]byte, 0), nil
}

// Verify -
func (stub *MessengerStub) Verify(payload []byte, pid core.PeerID, signature []byte) error {
	if stub.VerifyCalled != nil {
		return stub.VerifyCalled(payload, pid, signature)
	}

	return nil
}

// AddPeerTopicNotifier -
func (stub *MessengerStub) AddPeerTopicNotifier(notifier p2p.PeerTopicNotifier) error {
	if stub.AddPeerTopicNotifierCalled != nil {
		return stub.AddPeerTopicNotifierCalled(notifier)
	}

	return nil
}

// BroadcastUsingPrivateKey -
func (stub *MessengerStub) BroadcastUsingPrivateKey(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	if stub.BroadcastUsingPrivateKeyCalled != nil {
		stub.BroadcastUsingPrivateKeyCalled(topic, buff, pid, skBytes)
	}
}

// BroadcastOnChannelUsingPrivateKey -
func (stub *MessengerStub) BroadcastOnChannelUsingPrivateKey(channel string, topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	if stub.BroadcastOnChannelUsingPrivateKeyCalled != nil {
		stub.BroadcastOnChannelUsingPrivateKeyCalled(channel, topic, buff, pid, skBytes)
	}
}

// SignUsingPrivateKey -
func (stub *MessengerStub) SignUsingPrivateKey(skBytes []byte, payload []byte) ([]byte, error) {
	if stub.SignUsingPrivateKeyCalled != nil {
		return stub.SignUsingPrivateKeyCalled(skBytes, payload)
	}

	return make([]byte, 0), nil
}

// Type -
func (stub *MessengerStub) Type() p2p.NetworkMessengerType {
	if stub.TypeCalled != nil {
		return stub.TypeCalled()
	}

	return p2p.RegularMessenger
}

// IsInterfaceNil -
func (stub *MessengerStub) IsInterfaceNil() bool {
	return stub == nil
}
