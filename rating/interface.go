package rating

import "github.com/multiversx/mx-chain-core-go/core"

// connectionsProvider defines an entity able to provide connected peers
type connectionsProvider interface {
	ConnectedPeers() []core.PeerID
	IsInterfaceNil() bool
}
