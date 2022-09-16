package factory

import (
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-p2p/common"
	"github.com/ElrondNetwork/elrond-go-p2p/config"
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p/networksharding"
	"github.com/libp2p/go-libp2p-core/peer"
)

var log = logger.GetOrCreate("p2p/networksharding/factory")

// ArgsSharderFactory represents the argument for the sharder factory
type ArgsSharderFactory struct {
	PeerShardResolver    common.PeerShardResolver
	Pid                  peer.ID
	P2pConfig            config.P2PConfig
	PreferredPeersHolder common.PreferredPeersHolderHandler
	NodeOperationMode    common.NodeOperation
}

// NewSharder creates new Sharder instances
func NewSharder(arg ArgsSharderFactory) (common.Sharder, error) {
	shardingType := arg.P2pConfig.Sharding.Type
	switch shardingType {
	case common.ListsSharder:
		return listSharder(arg)
	case common.OneListSharder:
		return oneListSharder(arg)
	case common.NilListSharder:
		return nilListSharder()
	default:
		return nil, fmt.Errorf("%w when selecting sharder: unknown %s value", common.ErrInvalidValue, shardingType)
	}
}

func listSharder(arg ArgsSharderFactory) (common.Sharder, error) {
	switch arg.NodeOperationMode {
	case common.NormalOperation, common.FullArchiveMode:
	default:
		return nil, fmt.Errorf("%w unknown node operation mode %s", common.ErrInvalidValue, arg.NodeOperationMode)
	}

	log.Debug("using lists sharder",
		"MaxConnectionCount", arg.P2pConfig.Sharding.TargetPeerCount,
		"MaxIntraShardValidators", arg.P2pConfig.Sharding.MaxIntraShardValidators,
		"MaxCrossShardValidators", arg.P2pConfig.Sharding.MaxCrossShardValidators,
		"MaxIntraShardObservers", arg.P2pConfig.Sharding.MaxIntraShardObservers,
		"MaxCrossShardObservers", arg.P2pConfig.Sharding.MaxCrossShardObservers,
		"MaxFullHistoryObservers", arg.P2pConfig.Sharding.AdditionalConnections.MaxFullHistoryObservers,
		"MaxSeeders", arg.P2pConfig.Sharding.MaxSeeders,
		"node operation", arg.NodeOperationMode,
	)
	argListsSharder := networksharding.ArgListsSharder{
		PeerResolver:         arg.PeerShardResolver,
		SelfPeerId:           arg.Pid,
		P2pConfig:            arg.P2pConfig,
		PreferredPeersHolder: arg.PreferredPeersHolder,
		NodeOperationMode:    arg.NodeOperationMode,
	}
	return networksharding.NewListsSharder(argListsSharder)
}

func oneListSharder(arg ArgsSharderFactory) (common.Sharder, error) {
	log.Debug("using one list sharder",
		"MaxConnectionCount", arg.P2pConfig.Sharding.TargetPeerCount,
	)
	return networksharding.NewOneListSharder(
		arg.Pid,
		int(arg.P2pConfig.Sharding.TargetPeerCount),
	)
}

func nilListSharder() (common.Sharder, error) {
	log.Debug("using nil list sharder")
	return networksharding.NewNilListSharder(), nil
}
