package libp2p

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-p2p"
	"github.com/ElrondNetwork/elrond-go-p2p/data"
	"github.com/libp2p/go-libp2p-pubsub"
	pubsubPb "github.com/libp2p/go-libp2p-pubsub/pb"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

// TODO remove the header size of the message when commit d3c5ecd3a3e884206129d9f2a9a4ddfd5e7c8951 from
// https://github.com/libp2p/go-libp2p-pubsub/pull/189/commits will be part of a new release
var messageHeader = 64 * 1024 // 64kB
var maxSendBuffSize = (1 << 21) - messageHeader

const durationBetweenSends = time.Microsecond * 10

// ArgMessagesHandler is the DTO struct used to create a new instance of messages handler
type ArgMessagesHandler struct {
	PubSub             PubSub
	DirectSender       p2p.DirectSender
	Throttler          core.Throttler
	OutgoingCLB        ChannelLoadBalancer
	Marshaller         p2p.Marshaller
	ConnMonitorWrapper p2p.ConnectionMonitorWrapper
	PeersRatingHandler p2p.PeersRatingHandler
	Debugger           p2p.Debugger
	SyncTimer          p2p.SyncTimer
	PeerID             core.PeerID
}

type messagesHandler struct {
	ctx                context.Context
	cancelFunc         context.CancelFunc
	pubSub             PubSub
	directSender       p2p.DirectSender
	throttler          core.Throttler
	outgoingCLB        ChannelLoadBalancer
	marshaller         p2p.Marshaller
	connMonitorWrapper p2p.ConnectionMonitorWrapper
	peersRatingHandler p2p.PeersRatingHandler
	debugger           p2p.Debugger
	syncTimer          p2p.SyncTimer
	peerID             core.PeerID

	mutTopics     sync.RWMutex
	processors    map[string]TopicProcessor
	topics        map[string]PubSubTopic
	subscriptions map[string]PubSubSubscription
}

// NewMessagesHandler creates a new instance of messages handler
func NewMessagesHandler(args ArgMessagesHandler) (*messagesHandler, error) {
	err := checkArgMessagesHandler(args)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	handler := &messagesHandler{
		ctx:                ctx,
		cancelFunc:         cancel,
		pubSub:             args.PubSub,
		directSender:       args.DirectSender,
		throttler:          args.Throttler,
		outgoingCLB:        args.OutgoingCLB,
		marshaller:         args.Marshaller,
		connMonitorWrapper: args.ConnMonitorWrapper,
		peersRatingHandler: args.PeersRatingHandler,
		debugger:           args.Debugger,
		syncTimer:          args.SyncTimer,
		peerID:             args.PeerID,
		processors:         make(map[string]TopicProcessor),
		topics:             make(map[string]PubSubTopic),
		subscriptions:      make(map[string]PubSubSubscription),
	}

	err = handler.directSender.RegisterDirectMessageProcessor(handler)
	if err != nil {
		return nil, err
	}

	go handler.processChannelLoadBalancer(handler.outgoingCLB)

	return handler, nil
}

func checkArgMessagesHandler(args ArgMessagesHandler) error {
	if args.PubSub == nil {
		return p2p.ErrNilPubSub
	}
	if check.IfNil(args.DirectSender) {
		return p2p.ErrNilDirectSender
	}
	if check.IfNil(args.Throttler) {
		return p2p.ErrNilThrottler
	}
	if check.IfNil(args.OutgoingCLB) {
		return p2p.ErrNilChannelLoadBalancer
	}
	if check.IfNil(args.Marshaller) {
		return p2p.ErrNilMarshaller
	}
	if check.IfNil(args.ConnMonitorWrapper) {
		return p2p.ErrNilConnectionMonitorWrapper
	}
	if check.IfNil(args.PeersRatingHandler) {
		return p2p.ErrNilPeersRatingHandler
	}
	if check.IfNil(args.Debugger) {
		return p2p.ErrNilDebugger
	}
	if check.IfNil(args.SyncTimer) {
		return p2p.ErrNilSyncTimer
	}

	return nil
}

func (handler *messagesHandler) processChannelLoadBalancer(outgoingCLB ChannelLoadBalancer) {
	for {
		select {
		case <-time.After(durationBetweenSends):
		case <-handler.ctx.Done():
			log.Debug("closing messages handler's send from channel load balancer go routine")
			return
		}

		sendableData := outgoingCLB.CollectOneElementFromChannels()
		if sendableData == nil {
			continue
		}

		handler.mutTopics.RLock()
		topic := handler.topics[sendableData.Topic]
		handler.mutTopics.RUnlock()
		if topic == nil {
			log.Warn("writing on a topic that the node did not register on - message dropped",
				"topic", sendableData.Topic,
			)

			continue
		}

		packedSendableDataBuff := handler.createMessageBytes(sendableData.Buff)
		if len(packedSendableDataBuff) == 0 {
			continue
		}

		errPublish := handler.publish(topic, sendableData, packedSendableDataBuff)
		if errPublish != nil {
			log.Trace("error sending data", "error", errPublish)
		}
	}
}

func (handler *messagesHandler) publish(topic PubSubTopic, data *SendableData, packedSendableDataBuff []byte) error {
	options := make([]pubsub.PubOpt, 0, 1)

	if data.Sk != nil {
		options = append(options, pubsub.WithSecretKeyAndPeerId(data.Sk, data.ID))
	}

	return topic.Publish(handler.ctx, packedSendableDataBuff, options...)
}

// Broadcast tries to send a byte buffer onto a topic using the topic name as channel
func (handler *messagesHandler) Broadcast(topic string, buff []byte) {
	handler.BroadcastOnChannel(topic, topic, buff)
}

// BroadcastOnChannel tries to send a byte buffer onto a topic using provided channel
func (handler *messagesHandler) BroadcastOnChannel(channel string, topic string, buff []byte) {
	go func() {
		err := handler.broadcastOnChannelBlocking(channel, topic, buff)
		if err != nil {
			log.Warn("p2p broadcast", "error", err.Error())
		}
	}()
}

// broadcastOnChannelBlocking tries to send a byte buffer onto a topic using provided channel
// It is a blocking method. It needs to be launched on a go routine
func (handler *messagesHandler) broadcastOnChannelBlocking(channel string, topic string, buff []byte) error {
	err := handler.checkSendableData(buff)
	if err != nil {
		return err
	}

	if !handler.throttler.CanProcess() {
		return p2p.ErrTooManyGoroutines
	}

	handler.throttler.StartProcessing()

	sendable := &SendableData{
		Buff:  buff,
		Topic: topic,
		ID:    peer.ID(handler.peerID),
	}
	handler.outgoingCLB.GetChannelOrDefault(channel) <- sendable
	handler.throttler.EndProcessing()
	return nil
}

// BroadcastUsingPrivateKey tries to send a byte buffer onto a topic using the topic name as channel
func (handler *messagesHandler) BroadcastUsingPrivateKey(
	topic string,
	buff []byte,
	pid core.PeerID,
	skBytes []byte,
) {
	handler.BroadcastOnChannelUsingPrivateKey(topic, topic, buff, pid, skBytes)
}

// BroadcastOnChannelUsingPrivateKey tries to send a byte buffer onto a topic using provided channel
func (handler *messagesHandler) BroadcastOnChannelUsingPrivateKey(
	channel string,
	topic string,
	buff []byte,
	pid core.PeerID,
	skBytes []byte,
) {
	go func() {
		err := handler.broadcastOnChannelBlockingUsingPrivateKey(channel, topic, buff, pid, skBytes)
		if err != nil {
			log.Warn("p2p broadcast using private key", "error", err.Error())
		}
	}()
}

// broadcastOnChannelBlockingUsingPrivateKey tries to send a byte buffer onto a topic using provided channel
// It is a blocking method. It needs to be launched on a go routine
func (handler *messagesHandler) broadcastOnChannelBlockingUsingPrivateKey(
	channel string,
	topic string,
	buff []byte,
	pid core.PeerID,
	skBytes []byte,
) error {
	id := peer.ID(pid)
	sk, err := libp2pCrypto.UnmarshalSecp256k1PrivateKey(skBytes)
	if err != nil {
		return err
	}

	err = handler.checkSendableData(buff)
	if err != nil {
		return err
	}

	if !handler.throttler.CanProcess() {
		return p2p.ErrTooManyGoroutines
	}

	handler.throttler.StartProcessing()

	sendable := &SendableData{
		Buff:  buff,
		Topic: topic,
		Sk:    sk,
		ID:    id,
	}
	handler.outgoingCLB.GetChannelOrDefault(channel) <- sendable
	handler.throttler.EndProcessing()
	return nil
}

func (handler *messagesHandler) checkSendableData(buff []byte) error {
	if len(buff) > maxSendBuffSize {
		return fmt.Errorf("%w, to be sent: %d, maximum: %d", p2p.ErrMessageTooLarge, len(buff), maxSendBuffSize)
	}
	if len(buff) == 0 {
		return p2p.ErrEmptyBufferToSend
	}

	return nil
}

// RegisterMessageProcessor registers a message process on a topic. The function allows registering multiple handlers
// on a topic. Each handler should be associated with a new identifier on the same topic. Using same identifier on different
// topics is allowed. The order of handler calling on a particular topic is not deterministic.
func (handler *messagesHandler) RegisterMessageProcessor(topic string, identifier string, msgProcessor p2p.MessageProcessor) error {
	if check.IfNil(msgProcessor) {
		return fmt.Errorf("%w when calling messagesHandler.RegisterMessageProcessor for topic %s",
			p2p.ErrNilValidator, topic)
	}

	handler.mutTopics.Lock()
	defer handler.mutTopics.Unlock()

	topicProcs := handler.processors[topic]
	if topicProcs == nil {
		topicProcs = newTopicProcessors()
		handler.processors[topic] = topicProcs

		err := handler.pubSub.RegisterTopicValidator(topic, handler.pubsubCallback(topicProcs, topic))
		if err != nil {
			return err
		}
	}

	err := topicProcs.AddTopicProcessor(identifier, msgProcessor)
	if err != nil {
		return fmt.Errorf("%w, topic %s", err, topic)
	}

	return nil
}

func (handler *messagesHandler) pubsubCallback(topicProcs TopicProcessor, topic string) func(ctx context.Context, pid peer.ID, message *pubsub.Message) bool {
	return func(ctx context.Context, pid peer.ID, message *pubsub.Message) bool {
		fromConnectedPeer := core.PeerID(pid)
		msg, err := handler.transformAndCheckMessage(message, fromConnectedPeer, topic)
		if err != nil {
			log.Trace("p2p validator - new message", "error", err.Error(), "topic", topic)
			return false
		}

		identifiers, msgProcessors := topicProcs.GetList()
		messageOk := true
		for index, msgProc := range msgProcessors {
			err = msgProc.ProcessReceivedMessage(msg, fromConnectedPeer)
			if err != nil {
				log.Trace("p2p validator",
					"error", err.Error(),
					"topic", topic,
					"originator", p2p.MessageOriginatorPid(msg),
					"from connected peer", p2p.PeerIdToShortString(fromConnectedPeer),
					"seq no", p2p.MessageOriginatorSeq(msg),
					"topic identifier", identifiers[index],
				)
				messageOk = false
			}
		}
		handler.processDebugMessage(topic, fromConnectedPeer, uint64(len(message.Data)), !messageOk)

		if messageOk {
			handler.peersRatingHandler.IncreaseRating(fromConnectedPeer)
		}

		return messageOk
	}
}

func (handler *messagesHandler) transformAndCheckMessage(pbMsg *pubsub.Message, pid core.PeerID, topic string) (p2p.MessageP2P, error) {
	msg, errUnmarshal := NewMessage(pbMsg, handler.marshaller)
	if errUnmarshal != nil {
		// this error is so severe that will need to blacklist both the originator and the connected peer as there is
		// no way this node can communicate with them
		handler.blacklistPid(pid, p2p.WrongP2PMessageBlacklistDuration)
		if pbMsg != nil {
			pidFrom := core.PeerID(pbMsg.From)
			handler.blacklistPid(pidFrom, p2p.WrongP2PMessageBlacklistDuration)
		}
		return nil, errUnmarshal
	}

	err := handler.checkMessage(msg, pid, topic)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (handler *messagesHandler) checkMessage(msg p2p.MessageP2P, pid core.PeerID, topic string) error {
	err := handler.validateMessageByTimestamp(msg)
	if err != nil {
		// not reprocessing nor re-broadcasting the same message over and over again
		log.Trace("received an invalid message",
			"originator pid", p2p.MessageOriginatorPid(msg),
			"from connected pid", p2p.PeerIdToShortString(pid),
			"sequence", hex.EncodeToString(msg.SeqNo()),
			"timestamp", msg.Timestamp(),
			"error", err,
		)
		handler.processDebugMessage(topic, pid, uint64(len(msg.Data())), true)

		return err
	}

	return nil
}

func (handler *messagesHandler) blacklistPid(pid core.PeerID, banDuration time.Duration) {
	if handler.connMonitorWrapper.PeerDenialEvaluator().IsDenied(pid) {
		return
	}
	if len(pid) == 0 {
		return
	}

	log.Debug("blacklisted due to incompatible p2p message",
		"pid", pid.Pretty(),
		"time", banDuration,
	)

	err := handler.connMonitorWrapper.PeerDenialEvaluator().UpsertPeerID(pid, banDuration)
	if err != nil {
		log.Warn("error blacklisting peer ID in network messenger",
			"pid", pid.Pretty(),
			"error", err.Error(),
		)
	}
}

// validateMessageByTimestamp will check that the message time stamp should be in the interval
// (now-pubsubTimeCacheDuration+acceptMessagesInAdvanceDuration, now+acceptMessagesInAdvanceDuration)
func (handler *messagesHandler) validateMessageByTimestamp(msg p2p.MessageP2P) error {
	now := handler.syncTimer.CurrentTime()
	isInFuture := now.Add(acceptMessagesInAdvanceDuration).Unix() < msg.Timestamp()
	if isInFuture {
		return fmt.Errorf("%w, self timestamp %d, message timestamp %d",
			p2p.ErrMessageTooNew, now.Unix(), msg.Timestamp())
	}

	past := now.Unix() - int64(pubsubTimeCacheDuration.Seconds())
	if msg.Timestamp() < past {
		return fmt.Errorf("%w, self timestamp %d, message timestamp %d",
			p2p.ErrMessageTooOld, now.Unix(), msg.Timestamp())
	}

	return nil
}

func (handler *messagesHandler) processDebugMessage(topic string, fromConnectedPeer core.PeerID, size uint64, isRejected bool) {
	if fromConnectedPeer == handler.peerID {
		handler.debugger.AddOutgoingMessage(topic, size, isRejected)
	} else {
		handler.debugger.AddIncomingMessage(topic, size, isRejected)
	}
}

// UnregisterMessageProcessor unregisters a message processes on a topic
func (handler *messagesHandler) UnregisterMessageProcessor(topic string, identifier string) error {
	handler.mutTopics.Lock()
	defer handler.mutTopics.Unlock()

	topicProcs := handler.processors[topic]
	if topicProcs == nil {
		return nil
	}

	err := topicProcs.RemoveTopicProcessor(identifier)
	if err != nil {
		return err
	}

	identifiers, _ := topicProcs.GetList()
	if len(identifiers) == 0 {
		handler.processors[topic] = nil

		return handler.pubSub.UnregisterTopicValidator(topic)
	}

	return nil
}

// UnregisterAllMessageProcessors will unregister all message processors for topics
func (handler *messagesHandler) UnregisterAllMessageProcessors() error {
	handler.mutTopics.Lock()
	defer handler.mutTopics.Unlock()

	for topic := range handler.processors {
		err := handler.pubSub.UnregisterTopicValidator(topic)
		if err != nil {
			return err
		}

		delete(handler.processors, topic)
	}
	return nil
}

// SendToConnectedPeer sends a direct message to a connected peer
func (handler *messagesHandler) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	err := handler.checkSendableData(buff)
	if err != nil {
		return err
	}

	buffToSend := handler.createMessageBytes(buff)
	if len(buffToSend) == 0 {
		return nil
	}

	if peerID == handler.peerID {
		return handler.sendDirectToSelf(topic, buffToSend)
	}

	err = handler.directSender.Send(topic, buffToSend, peerID)
	handler.debugger.AddOutgoingMessage(topic, uint64(len(buffToSend)), err != nil)

	return err
}

func (handler *messagesHandler) sendDirectToSelf(topic string, buff []byte) error {
	pubSubMsg := &pubsub.Message{
		Message: &pubsubPb.Message{
			From:      handler.peerID.Bytes(),
			Data:      buff,
			Seqno:     handler.directSender.NextSequenceNumber(),
			Topic:     &topic,
			Signature: handler.peerID.Bytes(),
		},
	}

	msg, err := NewMessage(pubSubMsg, handler.marshaller)
	if err != nil {
		return err
	}

	return handler.ProcessReceivedMessage(msg, handler.peerID)
}

// ProcessReceivedMessage handles received direct messages
func (handler *messagesHandler) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	topic := message.Topic()
	err := handler.checkMessage(message, fromConnectedPeer, topic)
	if err != nil {
		return err
	}

	handler.mutTopics.RLock()
	topicProcs := handler.processors[topic]
	handler.mutTopics.RUnlock()

	if check.IfNil(topicProcs) {
		return fmt.Errorf("%w on HandleDirectMessageReceived for topic %s", p2p.ErrNilValidator, topic)
	}
	identifiers, msgProcessors := topicProcs.GetList()

	go func(msg p2p.MessageP2P) {
		if check.IfNil(msg) {
			return
		}

		// we won't recheck the message id against the cacher here as there might be collisions since we are using
		// a separate sequence counter for direct sender
		messageOk := true
		for index, msgProc := range msgProcessors {
			errProcess := msgProc.ProcessReceivedMessage(msg, fromConnectedPeer)
			if errProcess != nil {
				log.Trace("p2p validator",
					"error", errProcess.Error(),
					"topic", msg.Topic(),
					"originator", p2p.MessageOriginatorPid(msg),
					"from connected peer", p2p.PeerIdToShortString(fromConnectedPeer),
					"seq no", p2p.MessageOriginatorSeq(msg),
					"topic identifier", identifiers[index],
				)
				messageOk = false
			}
		}

		handler.debugger.AddIncomingMessage(msg.Topic(), uint64(len(msg.Data())), !messageOk)

		if messageOk {
			handler.peersRatingHandler.IncreaseRating(fromConnectedPeer)
		}
	}(message)

	return nil
}

func (handler *messagesHandler) createMessageBytes(buff []byte) []byte {
	message := &data.TopicMessage{
		Version:   currentTopicMessageVersion,
		Payload:   buff,
		Timestamp: handler.syncTimer.CurrentTime().Unix(),
	}

	buffToSend, errMarshal := handler.marshaller.Marshal(message)
	if errMarshal != nil {
		log.Warn("error sending data", "error", errMarshal)
		return nil
	}

	return buffToSend
}

// CreateTopic opens a new topic using pubsub infrastructure
func (handler *messagesHandler) CreateTopic(name string, createChannelForTopic bool) error {
	handler.mutTopics.Lock()
	defer handler.mutTopics.Unlock()

	_, found := handler.topics[name]
	if found {
		return nil
	}

	topic, err := handler.pubSub.Join(name)
	if err != nil {
		return fmt.Errorf("%w for topic %s", err, name)
	}

	handler.topics[name] = topic
	subscrRequest, err := topic.Subscribe()
	if err != nil {
		return fmt.Errorf("%w for topic %s", err, name)
	}

	handler.subscriptions[name] = subscrRequest
	if createChannelForTopic {
		err = handler.outgoingCLB.AddChannel(name)
	}

	// just a dummy func to consume messages received by the newly created topic
	go func() {
		var errSubscrNext error
		for {
			_, errSubscrNext = subscrRequest.Next(handler.ctx)
			if errSubscrNext != nil {
				log.Debug("closed subscription",
					"topic", subscrRequest.Topic(),
					"err", errSubscrNext,
				)
				return
			}
		}
	}()

	return err
}

// HasTopic returns true if the topic has been created
func (handler *messagesHandler) HasTopic(name string) bool {
	handler.mutTopics.RLock()
	defer handler.mutTopics.RUnlock()

	_, ok := handler.topics[name]
	return ok
}

// UnJoinAllTopics call close on all topics
func (handler *messagesHandler) UnJoinAllTopics() error {
	handler.mutTopics.Lock()
	defer handler.mutTopics.Unlock()

	var errFound error
	for topicName, t := range handler.topics {
		subscription := handler.subscriptions[topicName]
		if subscription != nil {
			subscription.Cancel()
		}

		err := t.Close()
		if err != nil {
			log.Warn("error closing topic",
				"topic", topicName,
				"error", err,
			)
			errFound = err
		}

		delete(handler.topics, topicName)
	}

	return errFound
}

// Close closes the messages handler
func (handler *messagesHandler) Close() error {
	handler.cancelFunc()

	var err error
	log.Debug("closing messages handler's outgoing load balancer...")
	errOCLB := handler.outgoingCLB.Close()
	if errOCLB != nil {
		err = errOCLB
		log.Warn("messagesHandler.Close",
			"component", "outgoingCLB",
			"error", err)
	}

	log.Debug("closing messages handler's debugger...")
	errDebugger := handler.debugger.Close()
	if errDebugger != nil {
		err = errDebugger
		log.Warn("messagesHandler.Close",
			"component", "debugger",
			"error", err)
	}

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *messagesHandler) IsInterfaceNil() bool {
	return handler == nil
}
