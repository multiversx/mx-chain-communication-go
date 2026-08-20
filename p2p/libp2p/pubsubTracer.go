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
// controlWithoutTopic labels the control messages that carry no topic id, as iwant and idontwant only hold message ids
const controlWithoutTopic = "[control without topic]"

type pubsubTracer struct {
	mut         sync.RWMutex
	debugger    p2p.DiscardedMessagesDebugger
	rpcDebugger p2p.RPCDebugger
}

func newPubsubTracer() *pubsubTracer {
	return &pubsubTracer{}
}

// setDebugger returns false if the provided debugger is unable to record the extra p2p statistics
func (tracer *pubsubTracer) setDebugger(debugger p2p.Debugger) bool {
	if tracer == nil {
		return false
	}

	discardedDebugger, isDiscardedDebugger := debugger.(p2p.DiscardedMessagesDebugger)
	if !isDiscardedDebugger || check.IfNil(discardedDebugger) {
		discardedDebugger = nil
	}

	rpcDebugger, isRPCDebugger := debugger.(p2p.RPCDebugger)
	if !isRPCDebugger || check.IfNil(rpcDebugger) {
		rpcDebugger = nil
	}

	tracer.mut.Lock()
	tracer.debugger = discardedDebugger
	tracer.rpcDebugger = rpcDebugger
	tracer.mut.Unlock()

	return discardedDebugger != nil && rpcDebugger != nil
}

func (tracer *pubsubTracer) recordingRPCDebugger() p2p.RPCDebugger {
	if tracer == nil {
		return nil
	}

	tracer.mut.RLock()
	rpcDebugger := tracer.rpcDebugger
	tracer.mut.RUnlock()

	if rpcDebugger == nil || !rpcDebugger.IsRecording() {
		return nil
	}

	return rpcDebugger
}

// recordRPC covers the published and the control messages only, so the totals stay below the bytes the operating
// system reports: the subscriptions and the transport framing are not accounted.
func (tracer *pubsubTracer) recordRPC(rpc *pubsub.RPC, isIncoming bool) {
	if rpc == nil {
		return
	}

	rpcDebugger := tracer.recordingRPCDebugger()
	if rpcDebugger == nil {
		return
	}

	for _, msg := range rpc.Publish {
		rpcDebugger.AddRPCPublishedMessage(msg.GetTopic(), uint64(msg.Size()), isIncoming)
	}

	if rpc.Control == nil {
		return
	}

	for _, ihave := range rpc.Control.Ihave {
		rpcDebugger.AddRPCControlMessage(ihave.GetTopicID(), uint64(ihave.Size()), isIncoming)
	}
	for _, graft := range rpc.Control.Graft {
		rpcDebugger.AddRPCControlMessage(graft.GetTopicID(), uint64(graft.Size()), isIncoming)
	}
	for _, prune := range rpc.Control.Prune {
		rpcDebugger.AddRPCControlMessage(prune.GetTopicID(), uint64(prune.Size()), isIncoming)
	}
	for _, iwant := range rpc.Control.Iwant {
		rpcDebugger.AddRPCControlMessage(controlWithoutTopic, uint64(iwant.Size()), isIncoming)
	}
	for _, idontwant := range rpc.Control.Idontwant {
		rpcDebugger.AddRPCControlMessage(controlWithoutTopic, uint64(idontwant.Size()), isIncoming)
	}
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

// RecvRPC is invoked for every incoming RPC
func (tracer *pubsubTracer) RecvRPC(rpc *pubsub.RPC) {
	tracer.recordRPC(rpc, true)
}

// SendRPC is invoked for every outgoing RPC, once per destination peer
func (tracer *pubsubTracer) SendRPC(rpc *pubsub.RPC, _ peer.ID) {
	tracer.recordRPC(rpc, false)
}

// DropRPC does nothing
func (tracer *pubsubTracer) DropRPC(_ *pubsub.RPC, _ peer.ID) {}

// UndeliverableMessage does nothing
func (tracer *pubsubTracer) UndeliverableMessage(_ *pubsub.Message) {}
