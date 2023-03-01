package p2p

import (
	"context"
	"encoding/hex"
	"io"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
)

// MessageProcessor is the interface used to describe what a receive message processor should do
// All implementations that will be called from Messenger implementation will need to satisfy this interface
// If the function returns a non nil value, the received message will not be propagated to its connected peers
type MessageProcessor interface {
	ProcessReceivedMessage(message MessageP2P, fromConnectedPeer core.PeerID) error
	IsInterfaceNil() bool
}

// PeerDiscoverer defines the behaviour of a peer discovery mechanism
type PeerDiscoverer interface {
	Bootstrap() error
	Name() string
	IsInterfaceNil() bool
}

// Reconnecter defines the behaviour of a network reconnection mechanism
type Reconnecter interface {
	ReconnectToNetwork(ctx context.Context)
	IsInterfaceNil() bool
}

// MessageHandler defines the behaviour of a component able to send and process messages
type MessageHandler interface {
	io.Closer

	CreateTopic(name string, createChannelForTopic bool) error
	HasTopic(name string) bool
	RegisterMessageProcessor(topic string, identifier string, handler MessageProcessor) error
	UnregisterAllMessageProcessors() error
	UnregisterMessageProcessor(topic string, identifier string) error
	Broadcast(topic string, buff []byte)
	BroadcastOnChannel(channel string, topic string, buff []byte)
	BroadcastUsingPrivateKey(topic string, buff []byte, pid core.PeerID, skBytes []byte)
	BroadcastOnChannelUsingPrivateKey(channel string, topic string, buff []byte, pid core.PeerID, skBytes []byte)
	SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error
	UnJoinAllTopics() error
	IsInterfaceNil() bool
}

// ConnectionsHandler defines the behaviour of a component able to handle connections
type ConnectionsHandler interface {
	io.Closer

	Bootstrap() error
	Peers() []core.PeerID
	Addresses() []string
	ConnectToPeer(address string) error
	IsConnected(peerID core.PeerID) bool
	ConnectedPeers() []core.PeerID
	ConnectedAddresses() []string
	PeerAddresses(pid core.PeerID) []string
	ConnectedPeersOnTopic(topic string) []core.PeerID
	ConnectedFullHistoryPeersOnTopic(topic string) []core.PeerID
	SetPeerShardResolver(peerShardResolver PeerShardResolver) error
	GetConnectedPeersInfo() *ConnectedPeersInfo
	WaitForConnections(maxWaitingTime time.Duration, minNumOfPeers uint32)
	IsConnectedToTheNetwork() bool
	ThresholdMinConnectedPeers() int
	SetThresholdMinConnectedPeers(minConnectedPeers int) error
	SetPeerDenialEvaluator(handler PeerDenialEvaluator) error
	IsInterfaceNil() bool
}

// Messenger is the main interface used for communication with other peers
type Messenger interface {
	MessageHandler
	ConnectionsHandler

	ID() core.PeerID

	Port() int
	Sign(payload []byte) ([]byte, error)
	Verify(payload []byte, pid core.PeerID, signature []byte) error
	SignUsingPrivateKey(skBytes []byte, payload []byte) ([]byte, error)
	AddPeerTopicNotifier(notifier PeerTopicNotifier) error
	IsInterfaceNil() bool
}

// MessageP2P defines what a p2p message can do (should return)
type MessageP2P interface {
	From() []byte
	Data() []byte
	Payload() []byte
	SeqNo() []byte
	Topic() string
	Signature() []byte
	Key() []byte
	Peer() core.PeerID
	Timestamp() int64
	Type() MessageType
	IsInterfaceNil() bool
}

// DirectSender defines a component that can send direct messages to connected peers
type DirectSender interface {
	NextSequenceNumber() []byte
	Send(topic string, buff []byte, peer core.PeerID) error
	RegisterDirectMessageProcessor(handler MessageProcessor) error
	IsInterfaceNil() bool
}

// PeerDiscoveryFactory defines the factory for peer discoverer implementation
type PeerDiscoveryFactory interface {
	CreatePeerDiscoverer() (PeerDiscoverer, error)
	IsInterfaceNil() bool
}

// MessageOriginatorPid will output the message peer id in a pretty format
// If it can, it will display the last displayLastPidChars (12) characters from the pid
func MessageOriginatorPid(msg MessageP2P) string {
	return PeerIdToShortString(msg.Peer())
}

// PeerIdToShortString trims the first displayLastPidChars characters of the provided peer ID after
// converting the peer ID to string using the Pretty functionality
func PeerIdToShortString(pid core.PeerID) string {
	prettyPid := pid.Pretty()
	lenPrettyPid := len(prettyPid)

	if lenPrettyPid > displayLastPidChars {
		return "..." + prettyPid[lenPrettyPid-displayLastPidChars:]
	}

	return prettyPid
}

// MessageOriginatorSeq will output the sequence number as hex
func MessageOriginatorSeq(msg MessageP2P) string {
	return hex.EncodeToString(msg.SeqNo())
}

// PeerShardResolver is able to resolve the link between the provided PeerID and the shardID
type PeerShardResolver interface {
	GetPeerInfo(pid core.PeerID) core.P2PPeerInfo
	IsInterfaceNil() bool
}

// ConnectedPeersInfo represents the DTO structure used to output the metrics for connected peers
type ConnectedPeersInfo struct {
	SelfShardID              uint32
	UnknownPeers             []string
	Seeders                  []string
	IntraShardValidators     map[uint32][]string
	IntraShardObservers      map[uint32][]string
	CrossShardValidators     map[uint32][]string
	CrossShardObservers      map[uint32][]string
	FullHistoryObservers     map[uint32][]string
	NumValidatorsOnShard     map[uint32]int
	NumObserversOnShard      map[uint32]int
	NumPreferredPeersOnShard map[uint32]int
	NumIntraShardValidators  int
	NumIntraShardObservers   int
	NumCrossShardValidators  int
	NumCrossShardObservers   int
	NumFullHistoryObservers  int
}

// NetworkShardingCollector defines the updating methods used by the network sharding component
// The interface assures that the collected data will be used by the p2p network sharding components
type NetworkShardingCollector interface {
	UpdatePeerIDInfo(pid core.PeerID, pk []byte, shardID uint32)
	IsInterfaceNil() bool
}

// SignerVerifier is used in higher level protocol authentication of 2 peers after the basic p2p connection has been made
type SignerVerifier interface {
	Sign(payload []byte) ([]byte, error)
	Verify(payload []byte, pid core.PeerID, signature []byte) error
	IsInterfaceNil() bool
}

// Marshaller defines the 2 basic operations: serialize (marshal) and deserialize (unmarshal)
type Marshaller interface {
	Marshal(obj interface{}) ([]byte, error)
	Unmarshal(obj interface{}, buff []byte) error
	IsInterfaceNil() bool
}

// PreferredPeersHolderHandler defines the behavior of a component able to handle preferred peers operations
type PreferredPeersHolderHandler interface {
	PutConnectionAddress(peerID core.PeerID, address string)
	PutShardID(peerID core.PeerID, shardID uint32)
	Get() map[uint32][]core.PeerID
	Contains(peerID core.PeerID) bool
	Remove(peerID core.PeerID)
	Clear()
	IsInterfaceNil() bool
}

// PeerCounts represents the DTO structure used to output the count metrics for connected peers
type PeerCounts struct {
	UnknownPeers    int
	IntraShardPeers int
	CrossShardPeers int
}

// Sharder defines the eviction computing process of unwanted peers
type Sharder interface {
	SetSeeders(addresses []string)
	IsSeeder(pid core.PeerID) bool
	SetPeerShardResolver(psp PeerShardResolver) error
	IsInterfaceNil() bool
}

// PeerDenialEvaluator defines the behavior of a component that is able to decide if a peer ID is black listed or not
// TODO move antiflooding inside network messenger
type PeerDenialEvaluator interface {
	IsDenied(pid core.PeerID) bool
	UpsertPeerID(pid core.PeerID, duration time.Duration) error
	IsInterfaceNil() bool
}

// Debugger represent a p2p debugger able to print p2p statistics (messages received/sent per topic)
type Debugger interface {
	AddIncomingMessage(topic string, size uint64, isRejected bool)
	AddOutgoingMessage(topic string, size uint64, isRejected bool)
	Close() error
	IsInterfaceNil() bool
}

// SyncTimer represent an entity able to tell the current time
type SyncTimer interface {
	CurrentTime() time.Time
	IsInterfaceNil() bool
}

// ConnectionsWatcher represent an entity able to watch new connections
type ConnectionsWatcher interface {
	NewKnownConnection(pid core.PeerID, connection string)
	Close() error
	IsInterfaceNil() bool
}

// PeersRatingHandler represent an entity able to handle peers ratings
type PeersRatingHandler interface {
	AddPeers(pids []core.PeerID)
	IncreaseRating(pid core.PeerID)
	DecreaseRating(pid core.PeerID)
	GetTopRatedPeersFromList(peers []core.PeerID, minNumOfPeersExpected int) []core.PeerID
	Close() error
	IsInterfaceNil() bool
}

// PeerTopicNotifier represent an entity able to handle new notifications on a new peer on a topic
type PeerTopicNotifier interface {
	NewPeerFound(pid core.PeerID, topic string)
	IsInterfaceNil() bool
}
