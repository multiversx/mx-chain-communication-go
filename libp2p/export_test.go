package libp2p

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiversx/mx-chain-core-go/core"
	p2p "github.com/multiversx/mx-chain-p2p-go"
	"github.com/multiversx/mx-chain-storage-go/types"
	"github.com/whyrusleeping/timecache"
)

var MaxSendBuffSize = maxSendBuffSize
var BroadcastGoRoutines = broadcastGoRoutines
var PubsubTimeCacheDuration = pubsubTimeCacheDuration
var AcceptMessagesInAdvanceDuration = acceptMessagesInAdvanceDuration
var SequenceNumberSize = sequenceNumberSize

const CurrentTopicMessageVersion = currentTopicMessageVersion
const PollWaitForConnectionsInterval = pollWaitForConnectionsInterval

// SetHost -
func (handler *connectionsHandler) SetHost(newHost ConnectableHost) {
	handler.p2pHost = newHost
}

// SetHost -
func (netMes *networkMessenger) SetHost(newHost ConnectableHost) {
	netMes.p2pHost = newHost
	netMes.ConnectionsHandler.(*connectionsHandler).SetHost(newHost)
}

// SetLoadBalancer -
func (handler *messagesHandler) SetLoadBalancer(outgoingCLB ChannelLoadBalancer) {
	handler.outgoingCLB = outgoingCLB
}

// SetLoadBalancer -
func (netMes *networkMessenger) SetLoadBalancer(outgoingCLB ChannelLoadBalancer) {
	netMes.MessageHandler.(*messagesHandler).SetLoadBalancer(outgoingCLB)
}

// SetPeerDiscoverer -
func (handler *connectionsHandler) SetPeerDiscoverer(discoverer p2p.PeerDiscoverer) {
	handler.peerDiscoverer = discoverer
}

// SetPeerDiscoverer -
func (netMes *networkMessenger) SetPeerDiscoverer(discoverer p2p.PeerDiscoverer) {
	netMes.ConnectionsHandler.(*connectionsHandler).SetPeerDiscoverer(discoverer)
}

// PubsubCallback -
func (handler *messagesHandler) PubsubCallback(msgProc p2p.MessageProcessor, topic string) func(ctx context.Context, pid peer.ID, message *pubsub.Message) bool {
	topicProcs := newTopicProcessors()
	_ = topicProcs.AddTopicProcessor("identifier", msgProc)

	return handler.pubsubCallback(topicProcs, topic)
}

// PubsubCallback -
func (netMes *networkMessenger) PubsubCallback(handler p2p.MessageProcessor, topic string) func(ctx context.Context, pid peer.ID, message *pubsub.Message) bool {
	return netMes.MessageHandler.(*messagesHandler).PubsubCallback(handler, topic)
}

// ValidMessageByTimestamp -
func (handler *messagesHandler) ValidMessageByTimestamp(msg p2p.MessageP2P) error {
	return handler.validateMessageByTimestamp(msg)
}

// ValidMessageByTimestamp -
func (netMes *networkMessenger) ValidMessageByTimestamp(msg p2p.MessageP2P) error {
	return netMes.MessageHandler.(*messagesHandler).ValidMessageByTimestamp(msg)
}

// MapHistogram -
func (handler *connectionsHandler) MapHistogram(input map[uint32]int) string {
	return handler.mapHistogram(input)
}

// MapHistogram -
func (netMes *networkMessenger) MapHistogram(input map[uint32]int) string {
	return netMes.ConnectionsHandler.(*connectionsHandler).MapHistogram(input)
}

// PubsubHasTopic -
func (handler *messagesHandler) PubsubHasTopic(expectedTopic string) bool {
	topics := handler.pubSub.GetTopics()

	for _, topic := range topics {
		if topic == expectedTopic {
			return true
		}
	}
	return false
}

// Disconnect -
func (netMes *networkMessenger) Disconnect(pid core.PeerID) error {
	return netMes.p2pHost.Network().ClosePeer(peer.ID(pid))
}

// BroadcastOnChannelBlocking -
func (netMes *networkMessenger) BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error {
	return netMes.MessageHandler.(*messagesHandler).BroadcastOnChannelBlocking(channel, topic, buff)
}

// BroadcastOnChannelBlocking -
func (handler *messagesHandler) BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error {
	return handler.broadcastOnChannelBlocking(channel, topic, buff)
}

// BroadcastOnChannelBlockingUsingPrivateKey -
func (handler *messagesHandler) BroadcastOnChannelBlockingUsingPrivateKey(
	channel string,
	topic string,
	buff []byte,
	pid core.PeerID,
	skBytes []byte,
) error {
	return handler.broadcastOnChannelBlockingUsingPrivateKey(channel, topic, buff, pid, skBytes)
}

// ProcessReceivedDirectMessage -
func (ds *directSender) ProcessReceivedDirectMessage(message *pb.Message, fromConnectedPeer peer.ID) error {
	return ds.processReceivedDirectMessage(message, fromConnectedPeer)
}

// SeenMessages -
func (ds *directSender) SeenMessages() *timecache.TimeCache {
	return ds.seenMessages
}

// Counter -
func (ds *directSender) Counter() uint64 {
	return ds.counter
}

// Mutexes -
func (mh *MutexHolder) Mutexes() types.Cacher {
	return mh.mutexes
}

// DirectSender -
func (handler *messagesHandler) DirectSender() *directSender {
	return handler.directSender.(*directSender)
}

// SetSignerInDirectSender sets the signer in the direct sender
func (netMes *networkMessenger) SetSignerInDirectSender(signer p2p.SignerVerifier) {
	netMes.MessageHandler.(*messagesHandler).DirectSender().signer = signer
}

func (oplb *OutgoingChannelLoadBalancer) Chans() []chan *SendableData {
	return oplb.chans
}

func (oplb *OutgoingChannelLoadBalancer) Names() []string {
	return oplb.names
}

func (oplb *OutgoingChannelLoadBalancer) NamesChans() map[string]chan *SendableData {
	return oplb.namesChans
}

func DefaultSendChannel() string {
	return defaultSendChannel
}

func NewPeersOnChannel(
	peersRatingHandler p2p.PeersRatingHandler,
	fetchPeersHandler func(topic string) []peer.ID,
	refreshInterval time.Duration,
	ttlInterval time.Duration,
) (*peersOnChannel, error) {
	return newPeersOnChannel(peersRatingHandler, fetchPeersHandler, refreshInterval, ttlInterval)
}

func (poc *peersOnChannel) SetPeersOnTopic(topic string, lastUpdated time.Time, peers []core.PeerID) {
	poc.mutPeers.Lock()
	poc.peers[topic] = peers
	poc.lastUpdated[topic] = lastUpdated
	poc.mutPeers.Unlock()
}

func (poc *peersOnChannel) GetPeers(topic string) []core.PeerID {
	poc.mutPeers.RLock()
	defer poc.mutPeers.RUnlock()

	return poc.peers[topic]
}

func (poc *peersOnChannel) SetTimeHandler(handler func() time.Time) {
	poc.getTimeHandler = handler
}

func GetPort(port string, handler func(int) error) (int, error) {
	return getPort(port, handler)
}

func CheckFreePort(port int) error {
	return checkFreePort(port)
}

func NewTopicProcessors() *topicProcessors {
	return newTopicProcessors()
}

func NewUnknownPeerShardResolver() *unknownPeerShardResolver {
	return &unknownPeerShardResolver{}
}

// NewMessagesHandlerWithNoRoutine -
func NewMessagesHandlerWithNoRoutine(args ArgMessagesHandler) *messagesHandler {
	ctx, cancel := context.WithCancel(context.Background())
	handler := &messagesHandler{
		ctx:                ctx,
		cancelFunc:         cancel,
		pubSub:             args.PubSub,
		directSender:       args.DirectSender,
		throttler:          args.Throttler,
		outgoingCLB:        args.OutgoingCLB,
		marshaller:         args.Marshaller,
		connMonitor:        args.ConnMonitor,
		peersRatingHandler: args.PeersRatingHandler,
		debugger:           args.Debugger,
		syncTimer:          args.SyncTimer,
		peerID:             args.PeerID,
		processors:         make(map[string]TopicProcessor),
		topics:             make(map[string]PubSubTopic),
		subscriptions:      make(map[string]PubSubSubscription),
	}

	_ = handler.directSender.RegisterDirectMessageProcessor(handler)
	return handler
}

// BlacklistPid -
func (handler *messagesHandler) BlacklistPid(pid core.PeerID, banDuration time.Duration) {
	handler.blacklistPid(pid, banDuration)
}

// TransformAndCheckMessage -
func (handler *messagesHandler) TransformAndCheckMessage(pbMsg *pubsub.Message, pid core.PeerID, topic string) (p2p.MessageP2P, error) {
	return handler.transformAndCheckMessage(pbMsg, pid, topic)
}

// IncreaseRatingIfNeeded -
func (handler *messagesHandler) IncreaseRatingIfNeeded(msg p2p.MessageP2P, from core.PeerID) {
	handler.increaseRatingIfNeeded(msg, from)
}

// NewMessagesHandlerWithTopics -
func NewMessagesHandlerWithTopics(args ArgMessagesHandler, topics map[string]PubSubTopic, withRoutine bool) *messagesHandler {
	handler := NewMessagesHandlerWithNoRoutine(args)
	handler.topics = topics

	if withRoutine {
		go handler.processChannelLoadBalancer(handler.outgoingCLB)
	}

	return handler
}

// NewMessagesHandlerWithNoRoutineAndProcessors -
func NewMessagesHandlerWithNoRoutineAndProcessors(args ArgMessagesHandler, processors map[string]TopicProcessor) *messagesHandler {
	handler := NewMessagesHandlerWithNoRoutine(args)
	handler.processors = processors

	return handler
}

// NewMessagesHandlerWithNoRoutineTopicsAndSubscriptions -
func NewMessagesHandlerWithNoRoutineTopicsAndSubscriptions(args ArgMessagesHandler, topics map[string]PubSubTopic, subscriptions map[string]PubSubSubscription) *messagesHandler {
	handler := NewMessagesHandlerWithNoRoutine(args)
	handler.topics = topics
	handler.subscriptions = subscriptions

	return handler
}

// NewConnectionsHandlerWithNoRoutine -
func NewConnectionsHandlerWithNoRoutine(args ArgConnectionsHandler) *connectionsHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &connectionsHandler{
		ctx:                  ctx,
		cancelFunc:           cancel,
		p2pHost:              args.P2pHost,
		peersOnChannel:       args.PeersOnChannel,
		peerShardResolver:    args.PeerShardResolver,
		sharder:              args.Sharder,
		preferredPeersHolder: args.PreferredPeersHolder,
		connMonitor:          args.ConnMonitor,
		peerDiscoverer:       args.PeerDiscoverer,
		peerID:               args.PeerID,
		connectionsMetric:    args.ConnectionsMetric,
	}
}

// PrintConnectionsStatus -
func (handler *connectionsHandler) PrintConnectionsStatus() {
	handler.printConnectionsStatus()
}
