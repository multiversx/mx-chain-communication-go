package pubSub

import (
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	p2pMocks "github.com/multiversx/mx-chain-communication-go/p2p/mock"
)

// PubSubsHolderMock -
type PubSubsHolderMock struct {
	GetPubSubCalled func(topic string) (libp2p.PubSub, bool)
	CloseCalled     func() error
}

// GetPubSub -
func (mock *PubSubsHolderMock) GetPubSub(topic string) (libp2p.PubSub, bool) {
	if mock.GetPubSubCalled != nil {
		return mock.GetPubSubCalled(topic)
	}
	return &p2pMocks.PubSubStub{}, true
}

// Close -
func (mock *PubSubsHolderMock) Close() error {
	if mock.CloseCalled != nil {
		return mock.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (mock *PubSubsHolderMock) IsInterfaceNil() bool {
	return mock == nil
}
