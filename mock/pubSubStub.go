package mock

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

// PubSubStub -
type PubSubStub struct {
	JoinCalled                     func(topic string, opts ...pubsub.TopicOpt) (*pubsub.Topic, error)
	ListPeersCalled                func(topic string) []peer.ID
	RegisterTopicValidatorCalled   func(topic string, val interface{}, opts ...pubsub.ValidatorOpt) error
	UnregisterTopicValidatorCalled func(topic string) error
}

// Join -
func (stub *PubSubStub) Join(topic string, opts ...pubsub.TopicOpt) (*pubsub.Topic, error) {
	if stub.JoinCalled != nil {
		return stub.JoinCalled(topic, opts...)
	}
	return nil, nil
}

// ListPeers -
func (stub *PubSubStub) ListPeers(topic string) []peer.ID {
	if stub.ListPeersCalled != nil {
		return stub.ListPeersCalled(topic)
	}
	return nil
}

// RegisterTopicValidator -
func (stub *PubSubStub) RegisterTopicValidator(topic string, val interface{}, opts ...pubsub.ValidatorOpt) error {
	if stub.RegisterTopicValidatorCalled != nil {
		return stub.RegisterTopicValidatorCalled(topic, val, opts...)
	}
	return nil
}

// UnregisterTopicValidator -
func (stub *PubSubStub) UnregisterTopicValidator(topic string) error {
	if stub.UnregisterTopicValidatorCalled != nil {
		return stub.UnregisterTopicValidatorCalled(topic)
	}
	return nil
}
