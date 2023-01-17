package libp2p

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	p2p "github.com/ElrondNetwork/elrond-go-p2p"
	"github.com/ElrondNetwork/elrond-go-storage/types"
	"github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
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
func (netMes *networkMessenger) SetHost(newHost ConnectableHost) {
	netMes.p2pHost = newHost
}

// SetLoadBalancer -
func (handler *messagesHandler) SetLoadBalancer(outgoingCLB ChannelLoadBalancer) {
	handler.outgoingCLB = outgoingCLB
}

// SetLoadBalancer -
func (netMes *networkMessenger) SetLoadBalancer(outgoingCLB ChannelLoadBalancer) {
	netMes.messagesHandler.(*messagesHandler).SetLoadBalancer(outgoingCLB)
}

// SetPeerDiscoverer -
func (netMes *networkMessenger) SetPeerDiscoverer(discoverer p2p.PeerDiscoverer) {
	netMes.peerDiscoverer = discoverer
}

// PubsubCallback -
func (handler *messagesHandler) PubsubCallback(topicProcs TopicProcessor, topic string) func(ctx context.Context, pid peer.ID, message *pubsub.Message) bool {
	return handler.pubsubCallback(topicProcs, topic)
}

// PubsubCallback -
func (netMes *networkMessenger) PubsubCallback(handler p2p.MessageProcessor, topic string) func(ctx context.Context, pid peer.ID, message *pubsub.Message) bool {
	topicProcs := newTopicProcessors()
	_ = topicProcs.AddTopicProcessor("identifier", handler)

	return netMes.messagesHandler.(*messagesHandler).PubsubCallback(topicProcs, topic)
}

// ValidMessageByTimestamp -
func (handler *messagesHandler) ValidMessageByTimestamp(msg p2p.MessageP2P) error {
	return handler.validateMessageByTimestamp(msg)
}

// ValidMessageByTimestamp -
func (netMes *networkMessenger) ValidMessageByTimestamp(msg p2p.MessageP2P) error {
	return netMes.messagesHandler.(*messagesHandler).ValidMessageByTimestamp(msg)
}

// MapHistogram -
func (netMes *networkMessenger) MapHistogram(input map[uint32]int) string {
	return netMes.mapHistogram(input)
}

// PubsubHasTopic -
func (netMes *networkMessenger) PubsubHasTopic(expectedTopic string) bool {
	topics := netMes.pb.GetTopics()

	for _, topic := range topics {
		if topic == expectedTopic {
			return true
		}
	}
	return false
}

// HasProcessorForTopic -
func (handler *messagesHandler) HasProcessorForTopic(expectedTopic string) bool {
	processor := handler.topicsHandler.GetTopicProcessors(expectedTopic)
	return processor != nil
}

// HasProcessorForTopic -
func (netMes *networkMessenger) HasProcessorForTopic(expectedTopic string) bool {
	return netMes.messagesHandler.(*messagesHandler).HasProcessorForTopic(expectedTopic)
}

// Disconnect -
func (netMes *networkMessenger) Disconnect(pid core.PeerID) error {
	return netMes.p2pHost.Network().ClosePeer(peer.ID(pid))
}

// BroadcastOnChannelBlocking -
func (handler *messagesHandler) BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error {
	return handler.broadcastOnChannelBlocking(channel, topic, buff)
}

// BroadcastOnChannelBlocking -
func (netMes *networkMessenger) BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error {
	return netMes.messagesHandler.(*messagesHandler).BroadcastOnChannelBlocking(channel, topic, buff)
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
	return handler.ds.(*directSender)
}

// SetSignerInDirectSender sets the signer in the direct sender
func (netMes *networkMessenger) SetSignerInDirectSender(signer p2p.SignerVerifier) {
	netMes.messagesHandler.(*messagesHandler).DirectSender().signer = signer
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

func NewConnectionMonitorWrapper(
	network network.Network,
	connMonitor ConnectionMonitor,
	peerDenialEvaluator p2p.PeerDenialEvaluator,
) *connectionMonitorWrapper {
	return newConnectionMonitorWrapper(network, connMonitor, peerDenialEvaluator)
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
		pb:                 args.PubSub,
		ds:                 args.DirectSender,
		throttler:          args.Throttler,
		outgoingCLB:        args.OutgoingCLB,
		topicsHandler:      args.TopicsHandler,
		marshaller:         args.Marshaller,
		connMonitorWrapper: args.ConnMonitorWrapper,
		peersRatingHandler: args.PeersRatingHandler,
		debugger:           args.Debugger,
		syncTimer:          args.SyncTimer,
		idProvider:         args.IDProvider,
	}

	_ = handler.ds.RegisterMessageHandler(handler.directMessageHandler)
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
