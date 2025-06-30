package config

// P2PConfig will hold all the P2P settings
type P2PConfig struct {
	Node                NodeConfig
	KadDhtPeerDiscovery KadDhtPeerDiscoveryConfig
	Sharding            ShardingConfig
	SubNetworks         SubNetworksConfig
}

// NodeConfig will hold basic p2p settings
type NodeConfig struct {
	Port                            string
	MaximumExpectedPeerCount        uint64
	ThresholdMinConnectedPeers      uint32
	MinNumPeersToWaitForOnBootstrap uint32
	Transports                      TransportConfig
	ResourceLimiter                 ResourceLimiterConfig
}

// TransportConfig specifies the supported protocols by the node
type TransportConfig struct {
	TCP                 TCPProtocolConfig
	QUICAddress         string
	WebSocketAddress    string
	WebTransportAddress string
}

// TCPProtocolConfig specifies the TCP protocol config
type TCPProtocolConfig struct {
	ListenAddress    string
	PreventPortReuse bool
}

// ResourceLimiterConfig specifies the resource limiter configuration
type ResourceLimiterConfig struct {
	Type                   string
	ManualSystemMemoryInMB int64
	ManualMaximumFD        int
	Ipv4ConnLimit          []ConnLimitConfig
	Ipv6ConnLimit          []ConnLimitConfig
}

// ConnLimitConfig specifies the limit that will be set for an ip on libp2p connection limiter
type ConnLimitConfig struct {
	PrefixLength int
	ConnCount    int
}

// KadDhtPeerDiscoveryConfig will hold the kad-dht discovery config settings
type KadDhtPeerDiscoveryConfig struct {
	Enabled                          bool
	Type                             string
	RefreshIntervalInSec             uint32
	ProtocolIDs                      []string
	InitialPeerList                  []string
	BucketSize                       uint32
	RoutingTableRefreshIntervalInSec uint32
}

// ShardingConfig will hold the network sharding config settings
type ShardingConfig struct {
	TargetPeerCount         uint32
	MaxIntraShardValidators uint32
	MaxCrossShardValidators uint32
	MaxIntraShardObservers  uint32
	MaxCrossShardObservers  uint32
	MaxSeeders              uint32
	Type                    string
}

// SubNetworksConfig holds an array of subnetworks configs
// all subnetworks will use the same host
type SubNetworksConfig struct {
	Networks []SubNetworkConfig
}

// SubNetworkConfig holds a subnetwork config
type SubNetworkConfig struct {
	Name        string
	PubSub      PubSubConfig
	ProtocolIDs []string
}

// PubSubConfig holds the configuration for PubSub
type PubSubConfig struct {
	OptimalPeersNum int
	MinimumPeersNum int
	MaximumPeersNum int
}
