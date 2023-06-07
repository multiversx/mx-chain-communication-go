package factory

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/config"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p/networksharding"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// ArgsSharderFactory represents the argument for the sharder factory
type ArgsSharderFactory struct {
	PeerShardResolver    p2p.PeerShardResolver
	Pid                  peer.ID
	P2pConfig            config.P2PConfig
	PreferredPeersHolder p2p.PreferredPeersHolderHandler
	NodeOperationMode    p2p.NodeOperation
	Logger               p2p.Logger
}

// NewSharder creates new Sharder instances
func NewSharder(arg ArgsSharderFactory) (p2p.Sharder, error) {
	if check.IfNil(arg.Logger) {
		return nil, p2p.ErrNilLogger
	}

	shardingType := arg.P2pConfig.Sharding.Type
	switch shardingType {
	case p2p.ListsSharder:
		return listSharder(arg)
	case p2p.OneListSharder:
		return oneListSharder(arg)
	case p2p.NilListSharder:
		return nilListSharder(arg.Logger)
	default:
		return nil, fmt.Errorf("%w when selecting sharder: unknown %s value", p2p.ErrInvalidValue, shardingType)
	}
}

func listSharder(arg ArgsSharderFactory) (p2p.Sharder, error) {
	switch arg.NodeOperationMode {
	case p2p.NormalOperation, p2p.FullArchiveMode:
	default:
		return nil, fmt.Errorf("%w unknown node operation mode %s", p2p.ErrInvalidValue, arg.NodeOperationMode)
	}

	arg.Logger.Debug("using lists sharder",
		"MaxConnectionCount", arg.P2pConfig.Sharding.TargetPeerCount,
		"MaxIntraShardValidators", arg.P2pConfig.Sharding.MaxIntraShardValidators,
		"MaxCrossShardValidators", arg.P2pConfig.Sharding.MaxCrossShardValidators,
		"MaxIntraShardObservers", arg.P2pConfig.Sharding.MaxIntraShardObservers,
		"MaxCrossShardObservers", arg.P2pConfig.Sharding.MaxCrossShardObservers,
		"MaxSeeders", arg.P2pConfig.Sharding.MaxSeeders,
		"node operation", arg.NodeOperationMode,
	)
	argListsSharder := networksharding.ArgListsSharder{
		PeerResolver:         arg.PeerShardResolver,
		SelfPeerId:           arg.Pid,
		P2pConfig:            arg.P2pConfig,
		PreferredPeersHolder: arg.PreferredPeersHolder,
		NodeOperationMode:    arg.NodeOperationMode,
		Logger:               arg.Logger,
	}
	return networksharding.NewListsSharder(argListsSharder)
}

func oneListSharder(arg ArgsSharderFactory) (p2p.Sharder, error) {
	arg.Logger.Debug("using one list sharder",
		"MaxConnectionCount", arg.P2pConfig.Sharding.TargetPeerCount,
	)
	return networksharding.NewOneListSharder(
		arg.Pid,
		int(arg.P2pConfig.Sharding.TargetPeerCount),
	)
}

func nilListSharder(log p2p.Logger) (p2p.Sharder, error) {
	log.Debug("using nil list sharder")
	return networksharding.NewNilListSharder(), nil
}
