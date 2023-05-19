package networksharding

import (
	"math/big"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiversx/mx-chain-communication-go/p2p"
)

const (
	// MinAllowedConnectedPeersListSharder -
	MinAllowedConnectedPeersListSharder = minAllowedConnectedPeersListSharder
	// MinAllowedValidators -
	MinAllowedValidators = minAllowedValidators
	// MinAllowedObservers -
	MinAllowedObservers = minAllowedObservers
	// MinAllowedConnectedPeersOneSharder -
	MinAllowedConnectedPeersOneSharder = minAllowedConnectedPeersOneSharder
)

// GetMaxPeerCount -
func (ls *listsSharder) GetMaxPeerCount() int {
	return ls.maxPeerCount
}

// GetMaxIntraShardValidators -
func (ls *listsSharder) GetMaxIntraShardValidators() int {
	return ls.maxIntraShardValidators
}

// GetMaxCrossShardValidators -
func (ls *listsSharder) GetMaxCrossShardValidators() int {
	return ls.maxCrossShardValidators
}

// GetMaxIntraShardObservers -
func (ls *listsSharder) GetMaxIntraShardObservers() int {
	return ls.maxIntraShardObservers
}

// GetMaxCrossShardObservers -
func (ls *listsSharder) GetMaxCrossShardObservers() int {
	return ls.maxCrossShardObservers
}

// GetMaxSeeders -
func (ls *listsSharder) GetMaxSeeders() int {
	return ls.maxSeeders
}

// GetMaxFullHistoryObservers -
func (ls *listsSharder) GetMaxFullHistoryObservers() int {
	return ls.maxFullHistoryObservers
}

// GetMaxUnknown -
func (ls *listsSharder) GetMaxUnknown() int {
	return ls.maxUnknown
}

// ComputeDistanceByCountingBits -
func ComputeDistanceByCountingBits(src peer.ID, dest peer.ID) *big.Int {
	return computeDistanceByCountingBits(src, dest)
}

// ComputeDistanceLog2Based -
func ComputeDistanceLog2Based(src peer.ID, dest peer.ID) *big.Int {
	return computeDistanceLog2Based(src, dest)
}

// GetPeerShardResolver -
func (ls *listsSharder) GetPeerShardResolver() p2p.PeerShardResolver {
	return ls.peerShardResolver
}
