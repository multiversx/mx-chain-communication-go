package libp2p

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-p2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
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
	p2p.PeerDiscoverer
	SetSharder(sharder p2p.Sharder) error
}

type p2pSigner interface {
	Sign(payload []byte) ([]byte, error)
	Verify(payload []byte, pid core.PeerID, signature []byte) error
	SignUsingPrivateKey(skBytes []byte, payload []byte) ([]byte, error)
}

// SendableData represents the struct used in data throttler implementation
type SendableData struct {
	Buff  []byte
	Topic string
	Sk    crypto.PrivKey
	ID    peer.ID
}

// ChannelLoadBalancer defines what a load balancer that uses chans should do
type ChannelLoadBalancer interface {
	AddChannel(channel string) error
	RemoveChannel(channel string) error
	GetChannelOrDefault(channel string) chan *SendableData
	CollectOneElementFromChannels() *SendableData
	Close() error
	IsInterfaceNil() bool
}

// PubSub interface defines what a publish/subscribe system should do
type PubSub interface {
	Join(topic string, opts ...pubsub.TopicOpt) (*pubsub.Topic, error)
	ListPeers(topic string) []peer.ID
	RegisterTopicValidator(topic string, val interface{}, opts ...pubsub.ValidatorOpt) error
	UnregisterTopicValidator(topic string) error
	GetTopics() []string
}

// TopicProcessor interface defines what a topic processor can do
type TopicProcessor interface {
	AddTopicProcessor(identifier string, processor p2p.MessageProcessor) error
	RemoveTopicProcessor(identifier string) error
	GetList() ([]string, []p2p.MessageProcessor)
	IsInterfaceNil() bool
}

// PubSubSubscription interface defines what a pubSub subscription can do
type PubSubSubscription interface {
	Topic() string
	Next(ctx context.Context) (*pubsub.Message, error)
	Cancel()
}

// PubSubTopic interface defines what a pubSub topic can do
type PubSubTopic interface {
	Subscribe(opts ...pubsub.SubOpt) (*pubsub.Subscription, error)
	Publish(ctx context.Context, data []byte, opts ...pubsub.PubOpt) error
	Close() error
}

// PeersOnChannel interface defines what a component able to handle peers on a channel should do
type PeersOnChannel interface {
	ConnectedPeersOnChannel(topic string) []core.PeerID
	Close() error
	IsInterfaceNil() bool
}

// ConnectionsMetric is an extension of the libp2p network notifiee able to track connections metrics
type ConnectionsMetric interface {
	network.Notifiee

	ResetNumConnections() uint32
	ResetNumDisconnections() uint32
	IsInterfaceNil() bool
}
