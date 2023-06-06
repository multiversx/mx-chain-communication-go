package facade

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("p2p/facade")

type networkMessengersFacade struct {
	messengers map[p2p.NetworkMessengerType]p2p.Messenger
}

// NewNetworkMessengersFacade creates a new networkMessengersFacade instance
func NewNetworkMessengersFacade(messengers ...p2p.Messenger) (*networkMessengersFacade, error) {
	if len(messengers) == 0 {
		return nil, p2p.ErrEmptyMessengersList
	}

	facade := &networkMessengersFacade{
		messengers: make(map[p2p.NetworkMessengerType]p2p.Messenger, len(messengers)),
	}

	for _, messenger := range messengers {
		if check.IfNilReflect(messenger) {
			return nil, p2p.ErrNilMessenger
		}

		facade.messengers[messenger.Type()] = messenger
	}

	return facade, nil
}

// CreateTopic opens a new topic using pubsub infrastructure on the main messenger
func (facade *networkMessengersFacade) CreateTopic(name string, createChannelForTopic bool) error {
	return facade.CreateTopicForMessenger(p2p.RegularMessenger, name, createChannelForTopic)
}

// CreateTopicForMessenger opens a new topic using pubsub infrastructure on the provided messenger
func (facade *networkMessengersFacade) CreateTopicForMessenger(messengerType p2p.NetworkMessengerType, name string, createChannelForTopic bool) error {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling CreateTopic on an unknown messenger, returning error...")
		return fmt.Errorf("%w of provided type %s", p2p.ErrUnknownMessenger, messengerType)
	}

	return messenger.CreateTopic(name, createChannelForTopic)
}

// HasTopic returns true if the topic has been created on the main messenger
func (facade *networkMessengersFacade) HasTopic(name string) bool {
	return facade.HasMessengerTopic(p2p.RegularMessenger, name)
}

// HasMessengerTopic returns true if the topic has been created on the provided messenger
func (facade *networkMessengersFacade) HasMessengerTopic(messengerType p2p.NetworkMessengerType, name string) bool {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling HasTopic on an unknown messenger, returning false...")
		return false
	}

	return messenger.HasTopic(name)
}

// RegisterMessageProcessor registers a message process on a topic on the main messenger
func (facade *networkMessengersFacade) RegisterMessageProcessor(topic string, identifier string, handler p2p.MessageProcessor) error {
	return facade.RegisterMessageProcessorForMessenger(p2p.RegularMessenger, topic, identifier, handler)
}

// RegisterMessageProcessorForMessenger registers a message process on a topic on the provided messenger
func (facade *networkMessengersFacade) RegisterMessageProcessorForMessenger(messengerType p2p.NetworkMessengerType, topic string, identifier string, handler p2p.MessageProcessor) error {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling RegisterMessageProcessor on an unknown messenger, returning error...")
		return fmt.Errorf("%w of provided type %s", p2p.ErrUnknownMessenger, messengerType)
	}

	return messenger.RegisterMessageProcessor(topic, identifier, handler)
}

// UnregisterAllMessageProcessors will unregister all message processors for topics on the main messenger
func (facade *networkMessengersFacade) UnregisterAllMessageProcessors() error {
	return facade.UnregisterAllMessageProcessorsForMessenger(p2p.RegularMessenger)
}

// UnregisterAllMessageProcessorsForMessenger will unregister all message processors for topics on the provided messenger
func (facade *networkMessengersFacade) UnregisterAllMessageProcessorsForMessenger(messengerType p2p.NetworkMessengerType) error {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling UnregisterAllMessageProcessors on an unknown messenger, returning error...")
		return fmt.Errorf("%w of provided type %s", p2p.ErrUnknownMessenger, messengerType)
	}

	return messenger.UnregisterAllMessageProcessors()
}

// UnregisterMessageProcessor unregisters a message processes on a topic on the main messenger
func (facade *networkMessengersFacade) UnregisterMessageProcessor(topic string, identifier string) error {
	return facade.UnregisterMessageProcessorForMessenger(p2p.RegularMessenger, topic, identifier)
}

// UnregisterMessageProcessorForMessenger unregisters a message processes on a topic on the provided messenger
func (facade *networkMessengersFacade) UnregisterMessageProcessorForMessenger(messengerType p2p.NetworkMessengerType, topic string, identifier string) error {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling UnregisterMessageProcessor on an unknown messenger, returning error...")
		return fmt.Errorf("%w of provided type %s", p2p.ErrUnknownMessenger, messengerType)
	}

	return messenger.UnregisterMessageProcessor(topic, identifier)
}

// Broadcast tries to send a byte buffer onto a topic using the topic name as channel on the main messenger
func (facade *networkMessengersFacade) Broadcast(topic string, buff []byte) {
	facade.broadcastOnMessenger(p2p.RegularMessenger, topic, buff)
}

func (facade *networkMessengersFacade) broadcastOnMessenger(messengerType p2p.NetworkMessengerType, topic string, buff []byte) {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling Broadcast on an unknown messenger, returning...")
		return
	}

	messenger.Broadcast(topic, buff)
}

// BroadcastOnChannel tries to send a byte buffer onto a topic using provided channel on the main messenger
func (facade *networkMessengersFacade) BroadcastOnChannel(channel string, topic string, buff []byte) {
	facade.broadcastOnChannelOnMessenger(p2p.RegularMessenger, channel, topic, buff)
}

func (facade *networkMessengersFacade) broadcastOnChannelOnMessenger(messengerType p2p.NetworkMessengerType, channel string, topic string, buff []byte) {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling BroadcastOnChannel on an unknown messenger, returning...")
		return
	}

	messenger.BroadcastOnChannel(channel, topic, buff)
}

// BroadcastUsingPrivateKey tries to send a byte buffer onto a topic using the topic name as channel on the main messenger
func (facade *networkMessengersFacade) BroadcastUsingPrivateKey(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	facade.broadcastUsingPrivateKeyOnMessenger(p2p.RegularMessenger, topic, buff, pid, skBytes)
}

func (facade *networkMessengersFacade) broadcastUsingPrivateKeyOnMessenger(messengerType p2p.NetworkMessengerType, topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling BroadcastUsingPrivateKey on an unknown messenger, returning...")
		return
	}

	messenger.BroadcastUsingPrivateKey(topic, buff, pid, skBytes)
}

// BroadcastOnChannelUsingPrivateKey tries to send a byte buffer onto a topic using provided channel on the main messenger
func (facade *networkMessengersFacade) BroadcastOnChannelUsingPrivateKey(channel string, topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	facade.broadcastOnChannelUsingPrivateKeyOnMessenger(p2p.RegularMessenger, channel, topic, buff, pid, skBytes)
}

func (facade *networkMessengersFacade) broadcastOnChannelUsingPrivateKeyOnMessenger(messengerType p2p.NetworkMessengerType, channel string, topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling BroadcastOnChannelUsingPrivateKey on an unknown messenger, returning...")
		return
	}

	messenger.BroadcastOnChannelUsingPrivateKey(channel, topic, buff, pid, skBytes)
}

// SendToConnectedPeer sends a direct message to a connected peer on the main messenger
func (facade *networkMessengersFacade) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	return facade.SendToConnectedPeerOnMessenger(p2p.RegularMessenger, topic, buff, peerID)
}

// SendToConnectedPeerOnMessenger sends a direct message to a connected peer on the provided messenger
func (facade *networkMessengersFacade) SendToConnectedPeerOnMessenger(messengerType p2p.NetworkMessengerType, topic string, buff []byte, peerID core.PeerID) error {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling SendToConnectedPeer on an unknown messenger, returning error...")
		return fmt.Errorf("%w of provided type %s", p2p.ErrUnknownMessenger, messengerType)
	}

	return messenger.SendToConnectedPeer(topic, buff, peerID)
}

// UnJoinAllTopics call close on all topics on the main messenger
func (facade *networkMessengersFacade) UnJoinAllTopics() error {
	return facade.unJoinAllTopicsOnMessenger(p2p.RegularMessenger)
}

func (facade *networkMessengersFacade) unJoinAllTopicsOnMessenger(messengerType p2p.NetworkMessengerType) error {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling UnJoinAllTopics on an unknown messenger, returning error...")
		return fmt.Errorf("%w of provided type %s", p2p.ErrUnknownMessenger, messengerType)
	}

	return messenger.UnJoinAllTopics()
}

// Bootstrap will start the peer discovery mechanism on the main messenger
func (facade *networkMessengersFacade) Bootstrap() error {
	return facade.BootstrapMessenger(p2p.RegularMessenger)
}

// BootstrapMessenger will start the peer discovery mechanism on the provided messenger
func (facade *networkMessengersFacade) BootstrapMessenger(messengerType p2p.NetworkMessengerType) error {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling Bootstrap on an unknown messenger, returning error...")
		return fmt.Errorf("%w of provided type %s", p2p.ErrUnknownMessenger, messengerType)
	}

	return messenger.Bootstrap()
}

// Peers returns the list of all known peers ID (including self) of the main messenger
func (facade *networkMessengersFacade) Peers() []core.PeerID {
	return facade.peersOfMessenger(p2p.RegularMessenger)
}

func (facade *networkMessengersFacade) peersOfMessenger(messengerType p2p.NetworkMessengerType) []core.PeerID {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling Peers on an unknown messenger, returning empty slice...")
		return []core.PeerID{}
	}

	return messenger.Peers()
}

// Addresses returns all addresses found in peerstore's of the main messenger
func (facade *networkMessengersFacade) Addresses() []string {
	return facade.addressesOfMessenger(p2p.RegularMessenger)
}

func (facade *networkMessengersFacade) addressesOfMessenger(messengerType p2p.NetworkMessengerType) []string {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling Addresses on an unknown messenger, returning empty slice...")
		return []string{}
	}

	return messenger.Addresses()
}

// ConnectToPeer tries to open a new connection to a peer on the main messenger
func (facade *networkMessengersFacade) ConnectToPeer(address string) error {
	return facade.connectMessengerToPeer(p2p.RegularMessenger, address)
}

func (facade *networkMessengersFacade) connectMessengerToPeer(messengerType p2p.NetworkMessengerType, address string) error {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling ConnectToPeer on an unknown messenger, returning error...")
		return fmt.Errorf("%w of provided type %s", p2p.ErrUnknownMessenger, messengerType)
	}

	return messenger.ConnectToPeer(address)
}

// IsConnected returns true if the provided messenger is connected to the main peer
func (facade *networkMessengersFacade) IsConnected(peerID core.PeerID) bool {
	return facade.isConnectedOnMessenger(p2p.RegularMessenger, peerID)
}

func (facade *networkMessengersFacade) isConnectedOnMessenger(messengerType p2p.NetworkMessengerType, peerID core.PeerID) bool {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling IsConnected on an unknown messenger, returning false...")
		return false
	}

	return messenger.IsConnected(peerID)
}

// ConnectedPeers returns the current connected peers list of the main messenger
func (facade *networkMessengersFacade) ConnectedPeers() []core.PeerID {
	return facade.connectedPeersOfMessenger(p2p.RegularMessenger)
}

func (facade *networkMessengersFacade) connectedPeersOfMessenger(messengerType p2p.NetworkMessengerType) []core.PeerID {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling ConnectedPeers on an unknown messenger, returning empty slice...")
		return []core.PeerID{}
	}

	return messenger.ConnectedPeers()
}

// ConnectedAddresses returns all connected peer's addresses of the main messenger
func (facade *networkMessengersFacade) ConnectedAddresses() []string {
	return facade.connectedAddressesOfMessenger(p2p.RegularMessenger)
}

func (facade *networkMessengersFacade) connectedAddressesOfMessenger(messengerType p2p.NetworkMessengerType) []string {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling ConnectedAddresses on an unknown messenger, returning empty slice...")
		return []string{}
	}

	return messenger.ConnectedAddresses()
}

// PeerAddresses returns the peer's addresses of the provided pid, on the main messenger
func (facade *networkMessengersFacade) PeerAddresses(pid core.PeerID) []string {
	return facade.peerAddressesOfMessenger(p2p.RegularMessenger, pid)
}

func (facade *networkMessengersFacade) peerAddressesOfMessenger(messengerType p2p.NetworkMessengerType, pid core.PeerID) []string {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling PeerAddresses on an unknown messenger, returning empty slice...")
		return []string{}
	}

	return messenger.PeerAddresses(pid)
}

// ConnectedPeersOnTopic returns the connected peers on a provided topic of the main messenger
func (facade *networkMessengersFacade) ConnectedPeersOnTopic(topic string) []core.PeerID {
	return facade.ConnectedPeersOnTopicForMessenger(p2p.RegularMessenger, topic)
}

// ConnectedPeersOnTopicForMessenger returns the connected peers on a provided topic of the provided messenger
func (facade *networkMessengersFacade) ConnectedPeersOnTopicForMessenger(messengerType p2p.NetworkMessengerType, topic string) []core.PeerID {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling ConnectedPeersOnTopic on an unknown messenger, returning empty slice...")
		return []core.PeerID{}
	}

	return messenger.ConnectedPeersOnTopic(topic)
}

// SetPeerShardResolver sets the peer shard resolver on the main messenger
func (facade *networkMessengersFacade) SetPeerShardResolver(peerShardResolver p2p.PeerShardResolver) error {
	return facade.SetPeerShardResolverForMessenger(p2p.RegularMessenger, peerShardResolver)
}

// SetPeerShardResolverForMessenger sets the peer shard resolver on the provided messenger
func (facade *networkMessengersFacade) SetPeerShardResolverForMessenger(messengerType p2p.NetworkMessengerType, peerShardResolver p2p.PeerShardResolver) error {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling SetPeerShardResolver on an unknown messenger, returning error...")
		return fmt.Errorf("%w of provided type %s", p2p.ErrUnknownMessenger, messengerType)
	}

	return messenger.SetPeerShardResolver(peerShardResolver)
}

// GetConnectedPeersInfo gets the current connected peers information of the main messenger
func (facade *networkMessengersFacade) GetConnectedPeersInfo() *p2p.ConnectedPeersInfo {
	return facade.GetConnectedPeersInfoForMessenger(p2p.RegularMessenger)
}

// GetConnectedPeersInfoForMessenger gets the current connected peers information of the provided messenger
func (facade *networkMessengersFacade) GetConnectedPeersInfoForMessenger(messengerType p2p.NetworkMessengerType) *p2p.ConnectedPeersInfo {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling GetConnectedPeersInfo on an unknown messenger, returning...")
		return &p2p.ConnectedPeersInfo{}
	}

	return messenger.GetConnectedPeersInfo()
}

// WaitForConnections will wait the maxWaitingTime duration or until the target connected peers was achieved on the main messenger
func (facade *networkMessengersFacade) WaitForConnections(maxWaitingTime time.Duration, minNumOfPeers uint32) {
	facade.WaitForConnectionsOnMessenger(p2p.RegularMessenger, maxWaitingTime, minNumOfPeers)
}

// WaitForConnectionsOnMessenger will wait the maxWaitingTime duration or until the target connected peers was achieved on the provided messenger
func (facade *networkMessengersFacade) WaitForConnectionsOnMessenger(messengerType p2p.NetworkMessengerType, maxWaitingTime time.Duration, minNumOfPeers uint32) {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling WaitForConnections on an unknown messenger, returning...")
		return
	}

	messenger.WaitForConnections(maxWaitingTime, minNumOfPeers)
}

// IsConnectedToTheNetwork returns true if the provided messenger is connected to the network
func (facade *networkMessengersFacade) IsConnectedToTheNetwork() bool {
	messenger, found := facade.messengers[p2p.RegularMessenger]
	if !found {
		log.Warn("missing regular messenger, programming error...")
		return false
	}

	return messenger.IsConnectedToTheNetwork()
}

// ThresholdMinConnectedPeers returns the minimum connected peers before triggering a new reconnection on the main messenger
func (facade *networkMessengersFacade) ThresholdMinConnectedPeers() int {
	return facade.thresholdMinConnectedPeersOnMessenger(p2p.RegularMessenger)
}

func (facade *networkMessengersFacade) thresholdMinConnectedPeersOnMessenger(messengerType p2p.NetworkMessengerType) int {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling ThresholdMinConnectedPeers on an unknown messenger, returning 0...")
		return 0
	}

	return messenger.ThresholdMinConnectedPeers()
}

// SetThresholdMinConnectedPeers sets the minimum connected peers before triggering a new reconnection on the main messenger
func (facade *networkMessengersFacade) SetThresholdMinConnectedPeers(minConnectedPeers int) error {
	return facade.setThresholdMinConnectedPeersOnMessenger(p2p.RegularMessenger, minConnectedPeers)
}

func (facade *networkMessengersFacade) setThresholdMinConnectedPeersOnMessenger(messengerType p2p.NetworkMessengerType, minConnectedPeers int) error {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling SetThresholdMinConnectedPeers on an unknown messenger, returning error...")
		return fmt.Errorf("%w of provided type %s", p2p.ErrUnknownMessenger, messengerType)
	}

	return messenger.SetThresholdMinConnectedPeers(minConnectedPeers)
}

// SetPeerDenialEvaluator sets the peer black list handler on the main messenger
func (facade *networkMessengersFacade) SetPeerDenialEvaluator(handler p2p.PeerDenialEvaluator) error {
	return facade.setPeerDenialEvaluatorOnMessenger(p2p.RegularMessenger, handler)
}

func (facade *networkMessengersFacade) setPeerDenialEvaluatorOnMessenger(messengerType p2p.NetworkMessengerType, handler p2p.PeerDenialEvaluator) error {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling SetPeerDenialEvaluator on an unknown messenger, returning error...")
		return fmt.Errorf("%w of provided type %s", p2p.ErrUnknownMessenger, messengerType)
	}

	return messenger.SetPeerDenialEvaluator(handler)
}

// ID returns the common peer id
func (facade *networkMessengersFacade) ID() core.PeerID {
	messenger, found := facade.messengers[p2p.RegularMessenger]
	if !found {
		log.Warn("missing regular messenger, programming error...")
		return ""
	}

	return messenger.ID()
}

// Port returns the port that the provided messenger is using by the main messenger
func (facade *networkMessengersFacade) Port() int {
	return facade.portOfMessenger(p2p.RegularMessenger)
}

func (facade *networkMessengersFacade) portOfMessenger(messengerType p2p.NetworkMessengerType) int {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling Port on an unknown messenger, returning 0...")
		return 0
	}

	return messenger.Port()
}

// Sign calls Sign on the main messenger
func (facade *networkMessengersFacade) Sign(payload []byte) ([]byte, error) {
	messenger, found := facade.messengers[p2p.RegularMessenger]
	if !found {
		log.Warn("missing regular messenger, programming error...")
		return nil, p2p.ErrUnknownMessenger
	}

	return messenger.Sign(payload)
}

// Verify calls Verify on the main messenger
func (facade *networkMessengersFacade) Verify(payload []byte, pid core.PeerID, signature []byte) error {
	messenger, found := facade.messengers[p2p.RegularMessenger]
	if !found {
		log.Warn("missing regular messenger, programming error...")
		return p2p.ErrUnknownMessenger
	}

	return messenger.Verify(payload, pid, signature)
}

// SignUsingPrivateKey calls SignUsingPrivateKey on the main messenger
func (facade *networkMessengersFacade) SignUsingPrivateKey(skBytes []byte, payload []byte) ([]byte, error) {
	messenger, found := facade.messengers[p2p.RegularMessenger]
	if !found {
		log.Warn("missing regular messenger, programming error...")
		return nil, p2p.ErrUnknownMessenger
	}

	return messenger.SignUsingPrivateKey(skBytes, payload)
}

// AddPeerTopicNotifier adds a new peer topic notifier on the main messenger
func (facade *networkMessengersFacade) AddPeerTopicNotifier(notifier p2p.PeerTopicNotifier) error {
	return facade.addPeerTopicNotifierOnMessenger(p2p.RegularMessenger, notifier)
}

func (facade *networkMessengersFacade) addPeerTopicNotifierOnMessenger(messengerType p2p.NetworkMessengerType, notifier p2p.PeerTopicNotifier) error {
	messenger, found := facade.messengers[messengerType]
	if !found {
		log.Warn("calling AddPeerTopicNotifier on an unknown messenger, returning error...")
		return fmt.Errorf("%w of type %s", p2p.ErrUnknownMessenger, messengerType)
	}

	return messenger.AddPeerTopicNotifier(notifier)
}

// Type returns a specific facade type
func (facade *networkMessengersFacade) Type() p2p.NetworkMessengerType {
	return "Facade"
}

// Close cals close method on all managed messengers
func (facade *networkMessengersFacade) Close() error {
	var lastErr error
	for _, messenger := range facade.messengers {
		err := messenger.Close()
		if err != nil {
			lastErr = fmt.Errorf("%w while closing messenger %s", err, messenger.Type())
		}
	}

	return lastErr
}

// IsInterfaceNil returns true if there is no value under the interface
func (facade *networkMessengersFacade) IsInterfaceNil() bool {
	return facade == nil
}
