package rating

import "github.com/multiversx/mx-chain-core-go/core"

// ConnectionsProvider defines an entity able to provide connected peers
type ConnectionsProvider interface {
	ConnectedPeers() []core.PeerID
	IsInterfaceNil() bool
}
