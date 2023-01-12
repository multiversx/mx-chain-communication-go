package libp2p

import (
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
}

// TopicsHandler interface defines what a component able to handle topics should do
type TopicsHandler interface {
	GetTopic(topic string) *pubsub.Topic
	HasTopic(topic string) bool
	AddTopic(topic string, pubSubTopic *pubsub.Topic)
	RemoveTopic(topic string)
	GetAllTopics() map[string]*pubsub.Topic
	GetTopicProcessors(topic string) *topicProcessors
	AddNewTopicProcessors(topic string) *topicProcessors
	RemoveTopicProcessors(topic string)
	GetAllTopicsProcessors() map[string]*topicProcessors
	GetSubscription(topic string) *pubsub.Subscription
	AddSubscription(topic string, sub *pubsub.Subscription)
	IsInterfaceNil() bool
}

// IDProvider interface defines a component able to provide its own peer ID
type IDProvider interface {
	ID() peer.ID
	IsInterfaceNil() bool
}
