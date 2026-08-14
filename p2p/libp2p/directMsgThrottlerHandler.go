package libp2p

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-communication-go/p2p"
)

var _ DirectMsgThrottlerHandler = (*directMsgThrottlerHandler)(nil)

// ArgDirectMsgThrottlerHandler is the DTO used in the NewDirectMsgThrottlerHandler constructor
type ArgDirectMsgThrottlerHandler struct {
	MaxGoroutinesPerPeer int32
}

type directMsgThrottlerHandler struct {
	maxGoroutinesPerPeer int32
	activeForPeer        map[core.PeerID]int32
	mutThrottlers        sync.Mutex
}

// NewDirectMsgThrottlerHandler creates a new instance of directMsgThrottlerHandler
func NewDirectMsgThrottlerHandler(args ArgDirectMsgThrottlerHandler) (*directMsgThrottlerHandler, error) {
	if args.MaxGoroutinesPerPeer <= 0 {
		return nil, p2p.ErrInvalidValue
	}

	handler := &directMsgThrottlerHandler{
		maxGoroutinesPerPeer: args.MaxGoroutinesPerPeer,
		activeForPeer:        make(map[core.PeerID]int32),
	}

	return handler, nil
}

// TryStartProcessing atomically checks and reserves processing capacity for the peer
func (handler *directMsgThrottlerHandler) TryStartProcessing(pid core.PeerID) bool {
	handler.mutThrottlers.Lock()
	defer handler.mutThrottlers.Unlock()

	active := handler.activeForPeer[pid]
	if active >= handler.maxGoroutinesPerPeer {
		return false
	}

	handler.activeForPeer[pid] = active + 1
	return true
}

// EndProcessing marks the end of processing a message from the given peer
func (handler *directMsgThrottlerHandler) EndProcessing(pid core.PeerID) {
	handler.mutThrottlers.Lock()
	defer handler.mutThrottlers.Unlock()

	active := handler.activeForPeer[pid]
	if active <= 1 {
		delete(handler.activeForPeer, pid)
		return
	}

	handler.activeForPeer[pid] = active - 1
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *directMsgThrottlerHandler) IsInterfaceNil() bool {
	return handler == nil
}
