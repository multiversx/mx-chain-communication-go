package libp2p

import (
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-communication-go/p2p"
)

// pubsubTracer records the received messages that are not propagated further. This is the only observation point:
// pubsub drops byte identical messages before the topic validator runs, and reports ignored ones only to the tracer.
type pubsubTracer struct {
	mut      sync.RWMutex
	debugger p2p.DiscardedMessagesDebugger
}

func newPubsubTracer() *pubsubTracer {
	return &pubsubTracer{}
}

// setDebugger returns false if the provided debugger is unable to record the discarded messages
func (tracer *pubsubTracer) setDebugger(debugger p2p.Debugger) bool {
	if tracer == nil {
		return false
	}

	discardedDebugger, isDiscardedDebugger := debugger.(p2p.DiscardedMessagesDebugger)
	if !isDiscardedDebugger || check.IfNil(discardedDebugger) {
		discardedDebugger = nil
	}

	tracer.mut.Lock()
	tracer.debugger = discardedDebugger
	tracer.mut.Unlock()

	return discardedDebugger != nil
}

func (tracer *pubsubTracer) debuggerFor(msg *pubsub.Message) p2p.DiscardedMessagesDebugger {
	if tracer == nil || msg == nil {
		return nil
	}

	tracer.mut.RLock()
	defer tracer.mut.RUnlock()

	return tracer.debugger
}

// DuplicateMessage is invoked when pubsub drops a byte identical message that was already seen
func (tracer *pubsubTracer) DuplicateMessage(msg *pubsub.Message) {
	debugger := tracer.debuggerFor(msg)
	if debugger == nil {
		return
	}

	debugger.AddDuplicateMessage(msg.GetTopic(), uint64(len(msg.GetData())))
}

// RejectMessage is invoked when a message is rejected or ignored. Only the ignored ones are recorded here, as the
// rejected ones are already counted by the topic validator.
func (tracer *pubsubTracer) RejectMessage(msg *pubsub.Message, reason string) {
	if reason != pubsub.RejectValidationIgnored {
		return
	}

	debugger := tracer.debuggerFor(msg)
	if debugger == nil {
		return
	}

	debugger.AddIgnoredMessage(msg.GetTopic(), uint64(len(msg.GetData())))
}

// OnNewOutboundStream does nothing
func (tracer *pubsubTracer) OnNewOutboundStream(_ peer.ID, _ protocol.ID) {}

// OnClosedOutboundStream does nothing
func (tracer *pubsubTracer) OnClosedOutboundStream(_ peer.ID) {}

// Join does nothing
func (tracer *pubsubTracer) Join(_ string) {}

// Leave does nothing
func (tracer *pubsubTracer) Leave(_ string) {}

// Graft does nothing
func (tracer *pubsubTracer) Graft(_ peer.ID, _ string) {}

// Prune does nothing
func (tracer *pubsubTracer) Prune(_ peer.ID, _ string) {}

// ValidateMessage does nothing
func (tracer *pubsubTracer) ValidateMessage(_ *pubsub.Message) {}

// DeliverMessage does nothing
func (tracer *pubsubTracer) DeliverMessage(_ *pubsub.Message) {}

// ThrottlePeer does nothing
func (tracer *pubsubTracer) ThrottlePeer(_ peer.ID) {}

// RecvRPC does nothing
func (tracer *pubsubTracer) RecvRPC(_ *pubsub.RPC) {}

// SendRPC does nothing
func (tracer *pubsubTracer) SendRPC(_ *pubsub.RPC, _ peer.ID) {}

// DropRPC does nothing
func (tracer *pubsubTracer) DropRPC(_ *pubsub.RPC, _ peer.ID) {}

// UndeliverableMessage does nothing
func (tracer *pubsubTracer) UndeliverableMessage(_ *pubsub.Message) {}
