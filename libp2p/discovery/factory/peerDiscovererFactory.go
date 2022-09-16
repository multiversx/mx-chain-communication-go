package factory

import (
	"context"
	"fmt"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-p2p/common"
	"github.com/ElrondNetwork/elrond-go-p2p/config"
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p/discovery"
)

const typeLegacy = "legacy"
const typeOptimized = "optimized"
const defaultSeedersReconnectionInterval = time.Minute * 5

var log = logger.GetOrCreate("p2p/discovery/factory")

// ArgsPeerDiscoverer is the DTO struct used in the NewPeerDiscoverer function
type ArgsPeerDiscoverer struct {
	Context            context.Context
	Host               discovery.ConnectableHost
	Sharder            common.Sharder
	P2pConfig          config.P2PConfig
	ConnectionsWatcher common.ConnectionsWatcher
}

// NewPeerDiscoverer generates an implementation of PeerDiscoverer by parsing the p2pConfig struct
// Errors if config is badly formatted
func NewPeerDiscoverer(args ArgsPeerDiscoverer) (common.PeerDiscoverer, error) {
	if args.P2pConfig.KadDhtPeerDiscovery.Enabled {
		return createKadDhtPeerDiscoverer(args)
	}

	log.Debug("using nil discoverer")
	return discovery.NewNilDiscoverer(), nil
}

func createKadDhtPeerDiscoverer(args ArgsPeerDiscoverer) (common.PeerDiscoverer, error) {
	arg := discovery.ArgKadDht{
		Context:                     args.Context,
		Host:                        args.Host,
		KddSharder:                  args.Sharder,
		PeersRefreshInterval:        time.Second * time.Duration(args.P2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec),
		SeedersReconnectionInterval: defaultSeedersReconnectionInterval,
		ProtocolID:                  args.P2pConfig.KadDhtPeerDiscovery.ProtocolID,
		InitialPeersList:            args.P2pConfig.KadDhtPeerDiscovery.InitialPeerList,
		BucketSize:                  args.P2pConfig.KadDhtPeerDiscovery.BucketSize,
		RoutingTableRefresh:         time.Second * time.Duration(args.P2pConfig.KadDhtPeerDiscovery.RoutingTableRefreshIntervalInSec),
		ConnectionWatcher:           args.ConnectionsWatcher,
	}

	switch args.P2pConfig.Sharding.Type {
	case common.ListsSharder, common.OneListSharder, common.NilListSharder:
		return createKadDhtDiscoverer(args.P2pConfig, arg)
	default:
		return nil, fmt.Errorf("%w unable to select peer discoverer based on "+
			"selected sharder: unknown sharder '%s'", common.ErrInvalidValue, args.P2pConfig.Sharding.Type)
	}
}

func createKadDhtDiscoverer(p2pConfig config.P2PConfig, arg discovery.ArgKadDht) (common.PeerDiscoverer, error) {
	switch p2pConfig.KadDhtPeerDiscovery.Type {
	case typeLegacy:
		log.Debug("using continuous (legacy) kad dht discoverer")
		return discovery.NewContinuousKadDhtDiscoverer(arg)
	case typeOptimized:
		log.Debug("using optimized kad dht discoverer")
		return discovery.NewOptimizedKadDhtDiscoverer(arg)
	default:
		return nil, fmt.Errorf("%w unable to select peer discoverer based on type '%s'",
			common.ErrInvalidValue, p2pConfig.KadDhtPeerDiscovery.Type)
	}
}
