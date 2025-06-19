package libp2p

import (
	"context"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/config"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// argsPubSubsHolder is the DTO used to create a new instance of pubSubsHolder
type argsPubSubsHolder struct {
	Log                 p2p.Logger
	Host                ConnectableHost
	MainNetwork         p2p.NetworkType
	MessageSigning      bool
	P2pConfig           config.P2PConfig
	NetworkTopicsHolder NetworkTopicsHolder
}

type pubSubsHolder struct {
	pubSubsMap          map[p2p.NetworkType]PubSub
	mut                 sync.RWMutex
	log                 p2p.Logger
	p2pHost             ConnectableHost
	networkTopicsHolder NetworkTopicsHolder
	ctx                 context.Context
	cancelFunc          func()
}

// newPubSubsHolder returns a new instance of pubSubsHolder
func newPubSubsHolder(args argsPubSubsHolder) (*pubSubsHolder, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	holder := &pubSubsHolder{
		log:                 args.Log,
		p2pHost:             args.Host,
		networkTopicsHolder: args.NetworkTopicsHolder,
		ctx:                 ctx,
		cancelFunc:          cancelFunc,
	}

	messageSigningCfg := messageSigningConfig(args.MessageSigning)
	mainPubSub, err := holder.createMainPubSub(messageSigningCfg)
	if err != nil {
		return nil, err
	}

	pubSubs, err := holder.createPubSubsForSubNetworks(messageSigningCfg, args.P2pConfig)
	if err != nil {
		return nil, err
	}

	pubSubs[args.MainNetwork] = mainPubSub

	holder.pubSubsMap = pubSubs

	return holder, nil
}

func checkArgs(args argsPubSubsHolder) error {
	if check.IfNil(args.Log) {
		return p2p.ErrNilLogger
	}
	if check.IfNil(args.Log) {
		return p2p.ErrNilHost
	}
	if check.IfNil(args.NetworkTopicsHolder) {
		return p2p.ErrNilNetworkTopicsHolder
	}

	return nil
}

// GetPubSub returns the pubSub instance that holds the provided topic
func (holder *pubSubsHolder) GetPubSub(topic string) (PubSub, bool) {
	network := holder.networkTopicsHolder.GetNetworkTypeForTopic(topic)

	holder.mut.RLock()
	defer holder.mut.RUnlock()

	pubSub, found := holder.pubSubsMap[network]
	return pubSub, found
}

// Close closes the internal context
func (holder *pubSubsHolder) Close() error {
	holder.cancelFunc()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *pubSubsHolder) IsInterfaceNil() bool {
	return holder == nil
}

func (holder *pubSubsHolder) createMainPubSub(
	messageSigning messageSigningConfig,
) (PubSub, error) {
	optsPS := make([]pubsub.Option, 0)
	if messageSigning == withoutMessageSigning {
		holder.log.Warn("signature verification is turned off in network messenger instance. NOT recommended in production environment")
		optsPS = append(optsPS, pubsub.WithMessageSignaturePolicy(noSignPolicy))
	}

	optsPS = append(optsPS, pubsub.WithMaxMessageSize(pubSubMaxMessageSize))

	return pubsub.NewGossipSub(holder.ctx, holder.p2pHost, optsPS...)
}

func (holder *pubSubsHolder) createPubSubsForSubNetworks(
	messageSigning messageSigningConfig,
	p2pConfig config.P2PConfig,
) (map[p2p.NetworkType]PubSub, error) {
	subNetworksConfig := p2pConfig.SubNetworks

	pubSubs := make(map[p2p.NetworkType]PubSub)
	for _, networkCfg := range subNetworksConfig.Networks {
		pubSub, err := holder.createPubSubForSubNetwork(messageSigning, networkCfg)
		if err != nil {
			return nil, err
		}

		networkType := p2p.NetworkType(networkCfg.Name)
		pubSubs[networkType] = pubSub
	}

	return pubSubs, nil
}

func (holder *pubSubsHolder) createPubSubForSubNetwork(
	messageSigning messageSigningConfig,
	subNetworkConfig config.SubNetworkConfig,
) (PubSub, error) {
	optsPS := make([]pubsub.Option, 0)
	if messageSigning == withoutMessageSigning {
		holder.log.Warn("signature verification is turned off in network messenger instance. NOT recommended in production environment")
		optsPS = append(optsPS, pubsub.WithMessageSignaturePolicy(noSignPolicy))
	}

	optsPS = append(optsPS, pubsub.WithMaxMessageSize(pubSubMaxMessageSize))

	pubSubConfig := subNetworkConfig.PubSub
	gossipSubParams := pubsub.DefaultGossipSubParams()
	holder.log.Debug("pubsub instance running with the custom parameters",
		"subNetwork", subNetworkConfig.Name,
		"D", pubSubConfig.OptimalPeersNum,
		"Dhi", pubSubConfig.MaximumPeersNum,
		"Dlo", pubSubConfig.MinimumPeersNum)

	gossipSubParams.D = pubSubConfig.OptimalPeersNum
	gossipSubParams.Dhi = pubSubConfig.MaximumPeersNum
	gossipSubParams.Dlo = pubSubConfig.MinimumPeersNum

	optsPS = append(optsPS, pubsub.WithGossipSubParams(gossipSubParams))

	// when same host is reused, different protocol ids are needed for secondary instances
	protocolIDs := make([]protocol.ID, 0, len(subNetworkConfig.ProtocolIDs))
	for _, protocolIDStr := range subNetworkConfig.ProtocolIDs {
		protocolId := protocol.ID(protocolIDStr)
		protocolIDs = append(protocolIDs, protocolId)
	}

	optsPS = append(optsPS, pubsub.WithGossipSubProtocols(protocolIDs, pubsub.GossipSubDefaultFeatures))

	return pubsub.NewGossipSub(holder.ctx, holder.p2pHost, optsPS...)
}
