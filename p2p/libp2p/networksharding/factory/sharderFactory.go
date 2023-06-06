package factory

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/config"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p/networksharding"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const loggerName = "p2p/networksharding/factory"

var log = logger.GetOrCreate(loggerName)

// ArgsSharderFactory represents the argument for the sharder factory
type ArgsSharderFactory struct {
	PeerShardResolver    p2p.PeerShardResolver
	Pid                  peer.ID
	P2pConfig            config.P2PConfig
	PreferredPeersHolder p2p.PreferredPeersHolderHandler
}

// NewSharder creates new Sharder instances
func NewSharder(arg ArgsSharderFactory) (p2p.Sharder, error) {
	shardingType := arg.P2pConfig.Sharding.Type
	if len(arg.P2pConfig.Logger.Prefix) > 0 {
		log = logger.GetOrCreate(arg.P2pConfig.Logger.Prefix + loggerName)
	}
	switch shardingType {
	case p2p.ListsSharder:
		return listSharder(arg)
	case p2p.OneListSharder:
		return oneListSharder(arg)
	case p2p.NilListSharder:
		return nilListSharder()
	default:
		return nil, fmt.Errorf("%w when selecting sharder: unknown %s value", p2p.ErrInvalidValue, shardingType)
	}
}

func listSharder(arg ArgsSharderFactory) (p2p.Sharder, error) {

	log.Debug("using lists sharder",
		"MaxConnectionCount", arg.P2pConfig.Sharding.TargetPeerCount,
		"MaxIntraShardValidators", arg.P2pConfig.Sharding.MaxIntraShardValidators,
		"MaxCrossShardValidators", arg.P2pConfig.Sharding.MaxCrossShardValidators,
		"MaxIntraShardObservers", arg.P2pConfig.Sharding.MaxIntraShardObservers,
		"MaxCrossShardObservers", arg.P2pConfig.Sharding.MaxCrossShardObservers,
		"MaxSeeders", arg.P2pConfig.Sharding.MaxSeeders,
	)
	argListsSharder := networksharding.ArgListsSharder{
		PeerResolver:         arg.PeerShardResolver,
		SelfPeerId:           arg.Pid,
		P2pConfig:            arg.P2pConfig,
		PreferredPeersHolder: arg.PreferredPeersHolder,
	}
	return networksharding.NewListsSharder(argListsSharder)
}

func oneListSharder(arg ArgsSharderFactory) (p2p.Sharder, error) {
	log.Debug("using one list sharder",
		"MaxConnectionCount", arg.P2pConfig.Sharding.TargetPeerCount,
	)
	return networksharding.NewOneListSharder(
		arg.Pid,
		int(arg.P2pConfig.Sharding.TargetPeerCount),
	)
}

func nilListSharder() (p2p.Sharder, error) {
	log.Debug("using nil list sharder")
	return networksharding.NewNilListSharder(), nil
}
