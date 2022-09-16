package libp2p

import (
	"github.com/ElrondNetwork/elrond-go-p2p/common"
	"github.com/libp2p/go-libp2p-core/network"
)

// ConnectionMonitor defines the behavior of a connection monitor
type ConnectionMonitor interface {
	network.Notifiee
	IsConnectedToTheNetwork(netw network.Network) bool
	SetThresholdMinConnectedPeers(thresholdMinConnectedPeers int, netw network.Network)
	ThresholdMinConnectedPeers() int
	Close() error
	IsInterfaceNil() bool
}

// PeerDiscovererWithSharder extends the PeerDiscoverer with the possibility to set the sharder
type PeerDiscovererWithSharder interface {
	common.PeerDiscoverer
	SetSharder(sharder common.Sharder) error
}
