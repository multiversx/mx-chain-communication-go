package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	p2p "github.com/multiversx/mx-chain-p2p-go"
)

// DirectSenderStub -
type DirectSenderStub struct {
	NextSequenceNumberCalled             func() []byte
	SendCalled                           func(topic string, buff []byte, peer core.PeerID) error
	RegisterDirectMessageProcessorCalled func(handler p2p.MessageProcessor) error
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

// RegisterDirectMessageProcessor -
func (stub *DirectSenderStub) RegisterDirectMessageProcessor(handler p2p.MessageProcessor) error {
	if stub.RegisterDirectMessageProcessorCalled != nil {
		return stub.RegisterDirectMessageProcessorCalled(handler)
	}
	return nil
}

// IsInterfaceNil -
func (stub *DirectSenderStub) IsInterfaceNil() bool {
	return stub == nil
}
