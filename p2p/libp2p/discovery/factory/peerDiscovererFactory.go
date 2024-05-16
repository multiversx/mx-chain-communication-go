package factory

import (
	"context"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/config"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p/discovery"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const typeLegacy = "legacy"
const typeOptimized = "optimized"
const defaultSeedersReconnectionInterval = time.Minute * 5

// ArgsPeerDiscoverer is the DTO struct used in the NewPeerDiscoverer function
type ArgsPeerDiscoverer struct {
	Context            context.Context
	Host               discovery.ConnectableHost
	Sharder            p2p.Sharder
	P2pConfig          config.P2PConfig
	ConnectionsWatcher p2p.ConnectionsWatcher
	NetworkType        p2p.NetworkType
	Logger             p2p.Logger
}

// NewPeerDiscoverers generates an array of implementations of PeerDiscoverer by parsing the p2pConfig struct
// Errors if config is badly formatted
func NewPeerDiscoverers(args ArgsPeerDiscoverer) ([]p2p.PeerDiscoverer, error) {
	if check.IfNil(args.Logger) {
		return nil, p2p.ErrNilLogger
	}

	if args.P2pConfig.KadDhtPeerDiscovery.Enabled {
		if len(args.P2pConfig.KadDhtPeerDiscovery.ProtocolIDs) == 0 {
			return nil, fmt.Errorf("%w: KadDhtPeerDiscovery.Enabled is enabled but no protocol ID was provided", p2p.ErrInvalidConfig)
		}

		protocolIDs := args.P2pConfig.KadDhtPeerDiscovery.ProtocolIDs
		instances := make([]p2p.PeerDiscoverer, 0, len(protocolIDs))
		for i := 0; i < len(protocolIDs); i++ {
			instance, err := createKadDhtPeerDiscoverer(args, protocolIDs[i])
			if err != nil {
				return nil, err
			}

			instances = append(instances, instance)
		}

		return instances, nil
	}

	args.Logger.Debug("using nil discoverer")
	return []p2p.PeerDiscoverer{discovery.NewNilDiscoverer()}, nil
}

func createKadDhtPeerDiscoverer(args ArgsPeerDiscoverer, protocolID string) (p2p.PeerDiscoverer, error) {
	arg := discovery.ArgKadDht{
		Context:                     args.Context,
		Host:                        args.Host,
		KddSharder:                  args.Sharder,
		PeersRefreshInterval:        time.Second * time.Duration(args.P2pConfig.KadDhtPeerDiscovery.RefreshIntervalInSec),
		SeedersReconnectionInterval: defaultSeedersReconnectionInterval,
		ProtocolID:                  protocolID,
		InitialPeersList:            args.P2pConfig.KadDhtPeerDiscovery.InitialPeerList,
		BucketSize:                  args.P2pConfig.KadDhtPeerDiscovery.BucketSize,
		RoutingTableRefresh:         time.Second * time.Duration(args.P2pConfig.KadDhtPeerDiscovery.RoutingTableRefreshIntervalInSec),
		ConnectionWatcher:           args.ConnectionsWatcher,
		NetworkType:                 args.NetworkType,
		Logger:                      args.Logger,
	}

	switch args.P2pConfig.Sharding.Type {
	case p2p.ListsSharder, p2p.OneListSharder, p2p.NilListSharder:
		return createKadDhtDiscoverer(args.P2pConfig, arg)
	default:
		return nil, fmt.Errorf("%w unable to select peer discoverer based on "+
			"selected sharder: unknown sharder '%s'", p2p.ErrInvalidValue, args.P2pConfig.Sharding.Type)
	}
}

func createKadDhtDiscoverer(p2pConfig config.P2PConfig, arg discovery.ArgKadDht) (p2p.PeerDiscoverer, error) {
	switch p2pConfig.KadDhtPeerDiscovery.Type {
	case typeLegacy:
		arg.Logger.Debug("using continuous (legacy) kad dht discoverer")
		return discovery.NewContinuousKadDhtDiscoverer(arg)
	case typeOptimized:
		arg.Logger.Debug("using optimized kad dht discoverer")
		return discovery.NewOptimizedKadDhtDiscoverer(arg)
	default:
		return nil, fmt.Errorf("%w unable to select peer discoverer based on type '%s'",
			p2p.ErrInvalidValue, p2pConfig.KadDhtPeerDiscovery.Type)
	}
}
