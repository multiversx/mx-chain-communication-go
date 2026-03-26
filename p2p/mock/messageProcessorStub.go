package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-communication-go/p2p"
)

// MessageProcessorStub -
type MessageProcessorStub struct {
	ProcessMessageCalled func(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) ([]byte, bool, error)
}

// ProcessReceivedMessage -
func (mps *MessageProcessorStub) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) ([]byte, bool, error) {
	if mps.ProcessMessageCalled != nil {
		return mps.ProcessMessageCalled(message, fromConnectedPeer, source)
	}

	return []byte{}, true, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mps *MessageProcessorStub) IsInterfaceNil() bool {
	return mps == nil
}
