package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-communication-go/p2p"
)

// MessageHandlerStub -
type MessageHandlerStub struct {
	CreateTopicCalled                       func(networkType p2p.NetworkType, name string, createChannelForTopic bool) error
	HasTopicCalled                          func(name string) bool
	RegisterMessageProcessorCalled          func(topic string, identifier string, handler p2p.MessageProcessor) error
	UnregisterAllMessageProcessorsCalled    func() error
	UnregisterMessageProcessorCalled        func(topic string, identifier string) error
	BroadcastCalled                         func(topic string, buff []byte)
	BroadcastOnChannelCalled                func(channel string, topic string, buff []byte)
	BroadcastUsingPrivateKeyCalled          func(topic string, buff []byte, pid core.PeerID, skBytes []byte)
	BroadcastOnChannelUsingPrivateKeyCalled func(channel string, topic string, buff []byte, pid core.PeerID, skBytes []byte)
	SendToConnectedPeerCalled               func(topic string, buff []byte, peerID core.PeerID) error
	UnJoinAllTopicsCalled                   func() error
	ProcessReceivedMessageCalled            func(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) ([]byte, error)
	SetDebuggerCalled                       func(debugger p2p.Debugger) error
	CloseCalled                             func() error
}

// CreateTopic -
func (stub *MessageHandlerStub) CreateTopic(networkType p2p.NetworkType, name string, createChannelForTopic bool) error {
	if stub.CreateTopicCalled != nil {
		return stub.CreateTopicCalled(networkType, name, createChannelForTopic)
	}
	return nil
}

// HasTopic -
func (stub *MessageHandlerStub) HasTopic(name string) bool {
	if stub.HasTopicCalled != nil {
		return stub.HasTopicCalled(name)
	}
	return false
}

// RegisterMessageProcessor -
func (stub *MessageHandlerStub) RegisterMessageProcessor(topic string, identifier string, handler p2p.MessageProcessor) error {
	if stub.RegisterMessageProcessorCalled != nil {
		return stub.RegisterMessageProcessorCalled(topic, identifier, handler)
	}
	return nil
}

// UnregisterAllMessageProcessors -
func (stub *MessageHandlerStub) UnregisterAllMessageProcessors() error {
	if stub.UnregisterAllMessageProcessorsCalled != nil {
		return stub.UnregisterAllMessageProcessorsCalled()
	}
	return nil
}

// UnregisterMessageProcessor -
func (stub *MessageHandlerStub) UnregisterMessageProcessor(topic string, identifier string) error {
	if stub.UnregisterMessageProcessorCalled != nil {
		return stub.UnregisterMessageProcessorCalled(topic, identifier)
	}
	return nil
}

// Broadcast -
func (stub *MessageHandlerStub) Broadcast(topic string, buff []byte) {
	if stub.BroadcastCalled != nil {
		stub.BroadcastCalled(topic, buff)
	}
}

// BroadcastOnChannel -
func (stub *MessageHandlerStub) BroadcastOnChannel(channel string, topic string, buff []byte) {
	if stub.BroadcastOnChannelCalled != nil {
		stub.BroadcastOnChannelCalled(channel, topic, buff)
	}
}

// BroadcastUsingPrivateKey -
func (stub *MessageHandlerStub) BroadcastUsingPrivateKey(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	if stub.BroadcastUsingPrivateKeyCalled != nil {
		stub.BroadcastUsingPrivateKeyCalled(topic, buff, pid, skBytes)
	}
}

// BroadcastOnChannelUsingPrivateKey -
func (stub *MessageHandlerStub) BroadcastOnChannelUsingPrivateKey(channel string, topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	if stub.BroadcastOnChannelUsingPrivateKeyCalled != nil {
		stub.BroadcastOnChannelUsingPrivateKeyCalled(channel, topic, buff, pid, skBytes)
	}
}

// SendToConnectedPeer -
func (stub *MessageHandlerStub) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	if stub.SendToConnectedPeerCalled != nil {
		return stub.SendToConnectedPeerCalled(topic, buff, peerID)
	}
	return nil
}

// UnJoinAllTopics -
func (stub *MessageHandlerStub) UnJoinAllTopics() error {
	if stub.UnJoinAllTopicsCalled != nil {
		return stub.UnJoinAllTopicsCalled()
	}
	return nil
}

// ProcessReceivedMessage -
func (stub *MessageHandlerStub) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) ([]byte, error) {
	if stub.ProcessReceivedMessageCalled != nil {
		return stub.ProcessReceivedMessageCalled(message, fromConnectedPeer, source)
	}
	return []byte{}, nil
}

// SetDebugger -
func (stub *MessageHandlerStub) SetDebugger(debugger p2p.Debugger) error {
	if stub.SetDebuggerCalled != nil {
		return stub.SetDebuggerCalled(debugger)
	}
	return nil
}

// Close -
func (stub *MessageHandlerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (stub *MessageHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
