package mock

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-communication-go/p2p"
)

// MessageProcessorMock -
type MessageProcessorMock struct {
	mut      sync.RWMutex
	messages map[core.PeerID]int
}

// NewMessageProcessorMock -
func NewMessageProcessorMock() *MessageProcessorMock {
	return &MessageProcessorMock{
		messages: make(map[core.PeerID]int),
	}
}

// ProcessReceivedMessage -
func (processor *MessageProcessorMock) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID, _ p2p.MessageHandler) ([]byte, error) {
	processor.mut.Lock()
	defer processor.mut.Unlock()

	fmt.Printf("got message from %s\n", message.Peer().Pretty())
	processor.messages[message.Peer()]++

	return nil, nil
}

// GetMessages -
func (processor *MessageProcessorMock) GetMessages() map[core.PeerID]int {
	processor.mut.RLock()
	defer processor.mut.RUnlock()

	result := make(map[core.PeerID]int)
	for key, val := range processor.messages {
		result[key] = val
	}

	return result
}

// IsInterfaceNil -
func (processor *MessageProcessorMock) IsInterfaceNil() bool {
	return processor == nil
}
