package libp2p

import (
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
)

var _ p2p.PeerShardResolver = (*unknownPeerShardResolver)(nil)

type unknownPeerShardResolver struct {
}

// GetPeerInfo returns a P2PPeerInfo value holding an unknown peer value
func (upsr *unknownPeerShardResolver) GetPeerInfo(_ core.PeerID) core.P2PPeerInfo {
	return core.P2PPeerInfo{
		PeerType:    core.UnknownPeer,
		PeerSubType: core.RegularPeer,
		ShardID:     0,
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (upsr *unknownPeerShardResolver) IsInterfaceNil() bool {
	return upsr == nil
}
