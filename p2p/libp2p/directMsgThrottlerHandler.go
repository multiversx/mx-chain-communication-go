package libp2p

import (
	"sync"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreThrottler "github.com/multiversx/mx-chain-core-go/core/throttler"

	"github.com/multiversx/mx-chain-communication-go/p2p"
)

var _ DirectMsgThrottlerHandler = (*directMsgThrottlerHandler)(nil)

// ArgDirectMsgThrottlerHandler is the DTO used in the NewDirectMsgThrottlerHandler constructor
type ArgDirectMsgThrottlerHandler struct {
	MaxGoroutinesPerPeer int32
	Network              network.Network
	Logger               p2p.Logger
}

type directMsgThrottlerHandler struct {
	maxGoroutinesPerPeer int32
	throttlers           map[core.PeerID]core.Throttler
	mutThrottlers        sync.RWMutex
	log                  p2p.Logger
}

// NewDirectMsgThrottlerHandler creates a new instance of directMsgThrottlerHandler and registers it as a network Notifiee
func NewDirectMsgThrottlerHandler(args ArgDirectMsgThrottlerHandler) (*directMsgThrottlerHandler, error) {
	if args.MaxGoroutinesPerPeer <= 0 {
		return nil, p2p.ErrInvalidValue
	}
	if check.IfNilReflect(args.Network) {
		return nil, p2p.ErrNilNetwork
	}
	if check.IfNil(args.Logger) {
		return nil, p2p.ErrNilLogger
	}

	handler := &directMsgThrottlerHandler{
		maxGoroutinesPerPeer: args.MaxGoroutinesPerPeer,
		throttlers:           make(map[core.PeerID]core.Throttler),
		log:                  args.Logger,
	}

	args.Network.Notify(handler)

	return handler, nil
}

// CanProcess returns true if the peer is under the per-peer goroutine limit
func (handler *directMsgThrottlerHandler) CanProcess(pid core.PeerID) bool {
	handler.mutThrottlers.RLock()
	throttler, exists := handler.throttlers[pid]
	handler.mutThrottlers.RUnlock()
	if exists {
		return throttler.CanProcess()
	}

	handler.mutThrottlers.Lock()
	defer handler.mutThrottlers.Unlock()

	throttler, exists = handler.throttlers[pid]
	if !exists {
		var err error
		throttler, err = coreThrottler.NewNumGoRoutinesThrottler(handler.maxGoroutinesPerPeer)
		if err != nil {
			handler.log.Warn("could not create throttler for peer", "pid", pid.Pretty(), "error", err)
			return false
		}
		handler.throttlers[pid] = throttler
	}

	return throttler.CanProcess()
}

// StartProcessing marks the start of processing a message from the given peer
func (handler *directMsgThrottlerHandler) StartProcessing(pid core.PeerID) {
	handler.mutThrottlers.RLock()
	throttler, exists := handler.throttlers[pid]
	handler.mutThrottlers.RUnlock()
	if exists {
		throttler.StartProcessing()
		return
	}

	handler.mutThrottlers.Lock()
	defer handler.mutThrottlers.Unlock()

	throttler, exists = handler.throttlers[pid]
	if !exists {
		var err error
		throttler, err = coreThrottler.NewNumGoRoutinesThrottler(handler.maxGoroutinesPerPeer)
		if err != nil {
			handler.log.Warn("could not create throttler for peer", "pid", pid.Pretty(), "error", err)
			return
		}
		handler.throttlers[pid] = throttler
	}

	throttler.StartProcessing()
}

// EndProcessing marks the end of processing a message from the given peer
func (handler *directMsgThrottlerHandler) EndProcessing(pid core.PeerID) {
	handler.mutThrottlers.RLock()
	throttler, exists := handler.throttlers[pid]
	handler.mutThrottlers.RUnlock()
	if exists {
		throttler.EndProcessing()
		return
	}

	handler.mutThrottlers.Lock()
	defer handler.mutThrottlers.Unlock()

	throttler, exists = handler.throttlers[pid]
	if !exists {
		var err error
		throttler, err = coreThrottler.NewNumGoRoutinesThrottler(handler.maxGoroutinesPerPeer)
		if err != nil {
			handler.log.Warn("could not create throttler for peer", "pid", pid.Pretty(), "error", err)
			return
		}
		handler.throttlers[pid] = throttler
	}

	throttler.EndProcessing()
}

// Listen is called when network starts listening on an addr
func (handler *directMsgThrottlerHandler) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (handler *directMsgThrottlerHandler) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when a connection opened
func (handler *directMsgThrottlerHandler) Connected(network.Network, network.Conn) {}

// Disconnected is called when a connection closed; it removes the peer's throttler from the map
func (handler *directMsgThrottlerHandler) Disconnected(_ network.Network, conn network.Conn) {
	if conn == nil {
		return
	}

	pid := core.PeerID(conn.RemotePeer())

	handler.mutThrottlers.Lock()
	delete(handler.throttlers, pid)
	handler.mutThrottlers.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *directMsgThrottlerHandler) IsInterfaceNil() bool {
	return handler == nil
}
