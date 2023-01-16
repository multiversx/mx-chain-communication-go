package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// DirectSenderStub -
type DirectSenderStub struct {
	NextSequenceNumberCalled     func() []byte
	SendCalled                   func(topic string, buff []byte, peer core.PeerID) error
	RegisterMessageHandlerCalled func(handler func(msg *pubsub.Message, fromConnectedPeer core.PeerID) error) error
}

// NextSequenceNumber -
func (stub *DirectSenderStub) NextSequenceNumber() []byte {
	if stub.NextSequenceNumberCalled != nil {
		return stub.NextSequenceNumberCalled()
	}
	return nil
}

// Send -
func (stub *DirectSenderStub) Send(topic string, buff []byte, peer core.PeerID) error {
	if stub.SendCalled != nil {
		return stub.SendCalled(topic, buff, peer)
	}
	return nil
}

// RegisterMessageHandler -
func (stub *DirectSenderStub) RegisterMessageHandler(handler func(msg *pubsub.Message, fromConnectedPeer core.PeerID) error) error {
	if stub.RegisterMessageHandlerCalled != nil {
		return stub.RegisterMessageHandlerCalled(handler)
	}
	return nil
}

// IsInterfaceNil -
func (stub *DirectSenderStub) IsInterfaceNil() bool {
	return stub == nil
}
