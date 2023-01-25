package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-p2p-go"
)

// MessageProcessorStub -
type MessageProcessorStub struct {
	ProcessMessageCalled func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
}

// ProcessReceivedMessage -
func (mps *MessageProcessorStub) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	if mps.ProcessMessageCalled != nil {
		return mps.ProcessMessageCalled(message, fromConnectedPeer)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mps *MessageProcessorStub) IsInterfaceNil() bool {
	return mps == nil
}
