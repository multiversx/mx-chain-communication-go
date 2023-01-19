package libp2p

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/throttler"
	commonCrypto "github.com/ElrondNetwork/elrond-go-crypto"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	p2p "github.com/ElrondNetwork/elrond-go-p2p"
	"github.com/ElrondNetwork/elrond-go-p2p/config"
	"github.com/ElrondNetwork/elrond-go-p2p/debug"
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p/connectionMonitor"
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p/crypto"
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p/disabled"
	discoveryFactory "github.com/ElrondNetwork/elrond-go-p2p/libp2p/discovery/factory"
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p/metrics"
	metricsFactory "github.com/ElrondNetwork/elrond-go-p2p/libp2p/metrics/factory"
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p/networksharding/factory"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	// TestListenAddrWithIp4AndTcp defines the local host listening ip v.4 address and TCP used in testing
	TestListenAddrWithIp4AndTcp = "/ip4/127.0.0.1/tcp/"

	// DirectSendID represents the protocol ID for sending and receiving direct P2P messages
	DirectSendID = protocol.ID("/erd/directsend/1.0.0")

	durationCheckConnections        = time.Second
	refreshPeersOnTopic             = time.Second * 3
	ttlPeersOnTopic                 = time.Second * 10
	ttlConnectionsWatcher           = time.Hour * 2
	pubsubTimeCacheDuration         = 10 * time.Minute
	acceptMessagesInAdvanceDuration = 20 * time.Second // we are accepting the messages with timestamp in the future only for this delta
	pollWaitForConnectionsInterval  = time.Second
	broadcastGoRoutines             = 1000
	timeBetweenPeerPrints           = time.Second * 20
	timeBetweenExternalLoggersCheck = time.Second * 20
	minRangePortValue               = 1025
	noSignPolicy                    = pubsub.MessageSignaturePolicy(0) // should be used only in tests
	msgBindError                    = "address already in use"
	maxRetriesIfBindError           = 10

	baseErrorSuffix      = "when creating a new network messenger"
	pubSubMaxMessageSize = 1 << 21 // 2 MB
)

type messageSigningConfig bool

const (
	withMessageSigning    messageSigningConfig = true
	withoutMessageSigning messageSigningConfig = false
)

var log = logger.GetOrCreate("p2p/libp2p")

var _ p2p.Messenger = (*networkMessenger)(nil)
var externalPackages = []string{"dht", "nat", "basichost", "pubsub"}

func init() {
	pubsub.TimeCacheDuration = pubsubTimeCacheDuration

	for _, external := range externalPackages {
		_ = logger.GetOrCreate(fmt.Sprintf("external/%s", external))
	}
}

// TODO refactor this struct to have be a wrapper (with logic) over a glue code
// TODO[Sorin]: further cleanup of this struct
type networkMessenger struct {
	p2pSigner
	p2p.MessageHandler
	p2p.ConnectionsHandler

	ctx        context.Context
	cancelFunc context.CancelFunc
	p2pHost    ConnectableHost
	port       int
	// TODO refactor this (connMonitor & connMonitorWrapper)
	connMonitor             ConnectionMonitor
	connMonitorWrapper      p2p.ConnectionMonitorWrapper
	peerDiscoverer          p2p.PeerDiscoverer
	connectionsMetric       *metrics.Connections
	printConnectionsWatcher p2p.ConnectionsWatcher
	mutPeerTopicNotifiers   sync.RWMutex
	peerTopicNotifiers      []p2p.PeerTopicNotifier
}

// ArgsNetworkMessenger defines the options used to create a p2p wrapper
type ArgsNetworkMessenger struct {
	ListenAddress         string
	Marshaller            p2p.Marshaller
	P2pConfig             config.P2PConfig
	SyncTimer             p2p.SyncTimer
	PreferredPeersHolder  p2p.PreferredPeersHolderHandler
	NodeOperationMode     p2p.NodeOperation
	PeersRatingHandler    p2p.PeersRatingHandler
	ConnectionWatcherType string
	P2pPrivateKey         commonCrypto.PrivateKey
	P2pSingleSigner       commonCrypto.SingleSigner
	P2pKeyGenerator       commonCrypto.KeyGenerator
}

// NewNetworkMessenger creates a libP2P messenger by opening a port on the current machine
func NewNetworkMessenger(args ArgsNetworkMessenger) (*networkMessenger, error) {
	return newNetworkMessenger(args, withMessageSigning)
}

func newNetworkMessenger(args ArgsNetworkMessenger, messageSigning messageSigningConfig) (*networkMessenger, error) {
	if check.IfNil(args.Marshaller) {
		return nil, fmt.Errorf("%w %s", p2p.ErrNilMarshaller, baseErrorSuffix)
	}
	if check.IfNil(args.SyncTimer) {
		return nil, fmt.Errorf("%w %s", p2p.ErrNilSyncTimer, baseErrorSuffix)
	}
	if check.IfNil(args.PreferredPeersHolder) {
		return nil, fmt.Errorf("%w %s", p2p.ErrNilPreferredPeersHolder, baseErrorSuffix)
	}
	if check.IfNil(args.PeersRatingHandler) {
		return nil, fmt.Errorf("%w %s", p2p.ErrNilPeersRatingHandler, baseErrorSuffix)
	}
	if check.IfNil(args.P2pPrivateKey) {
		return nil, fmt.Errorf("%w %s", p2p.ErrNilP2pPrivateKey, baseErrorSuffix)
	}
	if check.IfNil(args.P2pSingleSigner) {
		return nil, fmt.Errorf("%w %s", p2p.ErrNilP2pSingleSigner, baseErrorSuffix)
	}
	if check.IfNil(args.P2pKeyGenerator) {
		return nil, fmt.Errorf("%w %s", p2p.ErrNilP2pKeyGenerator, baseErrorSuffix)
	}

	setupExternalP2PLoggers()

	p2pNode, err := constructNodeWithPortRetry(args)
	if err != nil {
		return nil, err
	}

	err = addComponentsToNode(args, p2pNode, messageSigning)
	if err != nil {
		log.LogIfError(p2pNode.p2pHost.Close())
		return nil, err
	}

	return p2pNode, nil
}

func constructNode(
	args ArgsNetworkMessenger,
) (*networkMessenger, error) {

	port, err := getPort(args.P2pConfig.Node.Port, checkFreePort)
	if err != nil {
		return nil, err
	}

	log.Debug("connectionWatcherType", "type", args.ConnectionWatcherType)
	connWatcher, err := metricsFactory.NewConnectionsWatcher(args.ConnectionWatcherType, ttlConnectionsWatcher)
	if err != nil {
		return nil, err
	}

	p2pPrivateKey, err := crypto.ConvertPrivateKeyToLibp2pPrivateKey(args.P2pPrivateKey)
	if err != nil {
		return nil, err
	}

	address := fmt.Sprintf(args.ListenAddress+"%d", port)
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(address),
		libp2p.Identity(p2pPrivateKey),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.DefaultTransports,
		// we need to disable relay option in order to save the node's bandwidth as much as possible
		libp2p.DisableRelay(),
		libp2p.NATPortMap(),
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	p2pSignerArgs := crypto.ArgsP2pSignerWrapper{
		PrivateKey: args.P2pPrivateKey,
		Signer:     args.P2pSingleSigner,
		KeyGen:     args.P2pKeyGenerator,
	}

	p2pSignerInstance, err := crypto.NewP2PSignerWrapper(p2pSignerArgs)
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	p2pNode := &networkMessenger{
		p2pSigner:               p2pSignerInstance,
		ctx:                     ctx,
		cancelFunc:              cancelFunc,
		p2pHost:                 NewConnectableHost(h),
		port:                    port,
		printConnectionsWatcher: connWatcher,
		peerTopicNotifiers:      make([]p2p.PeerTopicNotifier, 0),
	}

	return p2pNode, nil
}

func constructNodeWithPortRetry(
	args ArgsNetworkMessenger,
) (*networkMessenger, error) {

	var lastErr error
	for i := 0; i < maxRetriesIfBindError; i++ {
		p2pNode, err := constructNode(args)
		if err == nil {
			return p2pNode, nil
		}

		lastErr = err
		if !strings.Contains(err.Error(), msgBindError) {
			// not a bind error, return directly
			return nil, err
		}

		log.Debug("bind error in network messenger", "retry number", i+1, "error", err)
	}

	return nil, lastErr
}

func setupExternalP2PLoggers() {
	for _, external := range externalPackages {
		logLevel := logger.GetLoggerLogLevel("external/" + external)
		if logLevel > logger.LogTrace {
			continue
		}

		_ = logging.SetLogLevel(external, "DEBUG")
	}
}

func addComponentsToNode(
	args ArgsNetworkMessenger,
	p2pNode *networkMessenger,
	messageSigning messageSigningConfig,
) error {
	var err error

	preferredPeersHolder := args.PreferredPeersHolder
	peersRatingHandler := args.PeersRatingHandler
	marshaller := args.Marshaller

	pubSub, err := p2pNode.createPubSub(messageSigning)
	if err != nil {
		return err
	}

	peersOnChannelInstance, err := newPeersOnChannel(
		peersRatingHandler,
		pubSub.ListPeers,
		refreshPeersOnTopic,
		ttlPeersOnTopic)
	if err != nil {
		return err
	}

	sharder, err := p2pNode.createSharder(args)
	if err != nil {
		return err
	}

	err = p2pNode.createDiscoverer(args.P2pConfig, sharder)
	if err != nil {
		return err
	}

	err = p2pNode.createConnectionMonitor(args.P2pConfig, sharder, preferredPeersHolder)
	if err != nil {
		return err
	}

	p2pNode.createConnectionsMetric()

	ds, err := NewDirectSender(p2pNode.ctx, p2pNode.p2pHost, p2pNode, marshaller)
	if err != nil {
		return err
	}

	goRoutinesThrottler, err := throttler.NewNumGoRoutinesThrottler(broadcastGoRoutines)
	if err != nil {
		return err
	}

	argsMessageHandler := ArgMessagesHandler{
		PubSub:             pubSub,
		DirectSender:       ds,
		Throttler:          goRoutinesThrottler,
		OutgoingCLB:        NewOutgoingChannelLoadBalancer(),
		Marshaller:         marshaller,
		ConnMonitorWrapper: p2pNode.connMonitorWrapper,
		PeersRatingHandler: peersRatingHandler,
		Debugger:           debug.NewP2PDebugger(core.PeerID(p2pNode.p2pHost.ID())),
		SyncTimer:          args.SyncTimer,
		PeerID:             core.PeerID(p2pNode.p2pHost.ID()),
	}
	p2pNode.MessageHandler, err = NewMessagesHandler(argsMessageHandler)
	if err != nil {
		return err
	}

	argsConnectionsHandler := ArgConnectionsHandler{
		P2pHost:              p2pNode.p2pHost,
		PeersOnChannel:       peersOnChannelInstance,
		PeerShardResolver:    &unknownPeerShardResolver{},
		Sharder:              sharder,
		PreferredPeersHolder: preferredPeersHolder,
		ConnMonitor:          p2pNode.connMonitor,
	}
	p2pNode.ConnectionsHandler, err = NewConnectionsHandler(argsConnectionsHandler)
	if err != nil {
		return err
	}

	p2pNode.printLogs()

	return nil
}

func (netMes *networkMessenger) createPubSub(messageSigning messageSigningConfig) (PubSub, error) {
	optsPS := make([]pubsub.Option, 0)
	if messageSigning == withoutMessageSigning {
		log.Warn("signature verification is turned off in network messenger instance. NOT recommended in production environment")
		optsPS = append(optsPS, pubsub.WithMessageSignaturePolicy(noSignPolicy))
	}

	optsPS = append(optsPS,
		pubsub.WithPeerFilter(netMes.newPeerFound),
		pubsub.WithMaxMessageSize(pubSubMaxMessageSize),
	)

	return pubsub.NewGossipSub(netMes.ctx, netMes.p2pHost, optsPS...)
}

func (netMes *networkMessenger) newPeerFound(pid peer.ID, topic string) bool {
	netMes.mutPeerTopicNotifiers.RLock()
	defer netMes.mutPeerTopicNotifiers.RUnlock()
	for _, notifier := range netMes.peerTopicNotifiers {
		notifier.NewPeerFound(core.PeerID(pid), topic)
	}

	return true
}

func (netMes *networkMessenger) createSharder(argsNetMes ArgsNetworkMessenger) (p2p.Sharder, error) {
	args := factory.ArgsSharderFactory{
		PeerShardResolver:    &unknownPeerShardResolver{},
		Pid:                  netMes.p2pHost.ID(),
		P2pConfig:            argsNetMes.P2pConfig,
		PreferredPeersHolder: argsNetMes.PreferredPeersHolder,
		NodeOperationMode:    argsNetMes.NodeOperationMode,
	}

	return factory.NewSharder(args)
}

func (netMes *networkMessenger) createDiscoverer(p2pConfig config.P2PConfig, sharder p2p.Sharder) error {
	var err error

	args := discoveryFactory.ArgsPeerDiscoverer{
		Context:            netMes.ctx,
		Host:               netMes.p2pHost,
		Sharder:            sharder,
		P2pConfig:          p2pConfig,
		ConnectionsWatcher: netMes.printConnectionsWatcher,
	}

	netMes.peerDiscoverer, err = discoveryFactory.NewPeerDiscoverer(args)

	return err
}

func (netMes *networkMessenger) createConnectionMonitor(p2pConfig config.P2PConfig, sharderInstance p2p.Sharder, preferredPeersHolder p2p.PreferredPeersHolderHandler) error {
	reconnecter, ok := netMes.peerDiscoverer.(p2p.Reconnecter)
	if !ok {
		return fmt.Errorf("%w when converting peerDiscoverer to reconnecter interface", p2p.ErrWrongTypeAssertion)
	}

	sharder, ok := sharderInstance.(connectionMonitor.Sharder)
	if !ok {
		return fmt.Errorf("%w in networkMessenger.createConnectionMonitor", p2p.ErrWrongTypeAssertions)
	}

	args := connectionMonitor.ArgsConnectionMonitorSimple{
		Reconnecter:                reconnecter,
		Sharder:                    sharder,
		ThresholdMinConnectedPeers: p2pConfig.Node.ThresholdMinConnectedPeers,
		PreferredPeersHolder:       preferredPeersHolder,
		ConnectionsWatcher:         netMes.printConnectionsWatcher,
	}
	var err error
	netMes.connMonitor, err = connectionMonitor.NewLibp2pConnectionMonitorSimple(args)
	if err != nil {
		return err
	}

	cmw := newConnectionMonitorWrapper(
		netMes.p2pHost.Network(),
		netMes.connMonitor,
		&disabled.PeerDenialEvaluator{},
	)
	netMes.p2pHost.Network().Notify(cmw)
	netMes.connMonitorWrapper = cmw

	go func() {
		for {
			cmw.CheckConnectionsBlocking()
			select {
			case <-time.After(durationCheckConnections):
			case <-netMes.ctx.Done():
				log.Debug("peer monitoring go routine is stopping...")
				return
			}
		}
	}()

	return nil
}

func (netMes *networkMessenger) createConnectionsMetric() {
	netMes.connectionsMetric = metrics.NewConnections()
	netMes.p2pHost.Network().Notify(netMes.connectionsMetric)
}

func (netMes *networkMessenger) printLogs() {
	addresses := make([]interface{}, 0)
	for i, address := range netMes.p2pHost.Addrs() {
		addresses = append(addresses, fmt.Sprintf("addr%d", i))
		addresses = append(addresses, address.String()+"/p2p/"+netMes.ID().Pretty())
	}
	log.Info("listening on addresses", addresses...)

	go netMes.printLogsStats()
	go netMes.checkExternalLoggers()
}

func (netMes *networkMessenger) printLogsStats() {
	for {
		select {
		case <-netMes.ctx.Done():
			log.Debug("closing networkMessenger.printLogsStats go routine")
			return
		case <-time.After(timeBetweenPeerPrints):
		}

		conns := netMes.connectionsMetric.ResetNumConnections()
		disconns := netMes.connectionsMetric.ResetNumDisconnections()

		peersInfo := netMes.GetConnectedPeersInfo()
		log.Debug("network connection status",
			"known peers", len(netMes.Peers()),
			"connected peers", len(netMes.ConnectedPeers()),
			"intra shard validators", peersInfo.NumIntraShardValidators,
			"intra shard observers", peersInfo.NumIntraShardObservers,
			"cross shard validators", peersInfo.NumCrossShardValidators,
			"cross shard observers", peersInfo.NumCrossShardObservers,
			"full history observers", peersInfo.NumFullHistoryObservers,
			"unknown", len(peersInfo.UnknownPeers),
			"seeders", len(peersInfo.Seeders),
			"current shard", peersInfo.SelfShardID,
			"validators histogram", netMes.mapHistogram(peersInfo.NumValidatorsOnShard),
			"observers histogram", netMes.mapHistogram(peersInfo.NumObserversOnShard),
			"preferred peers histogram", netMes.mapHistogram(peersInfo.NumPreferredPeersOnShard),
		)

		connsPerSec := conns / uint32(timeBetweenPeerPrints/time.Second)
		disconnsPerSec := disconns / uint32(timeBetweenPeerPrints/time.Second)

		log.Debug("network connection metrics",
			"connections/s", connsPerSec,
			"disconnections/s", disconnsPerSec,
			"connections", conns,
			"disconnections", disconns,
			"time", timeBetweenPeerPrints,
		)
	}
}

func (netMes *networkMessenger) mapHistogram(input map[uint32]int) string {
	keys := make([]uint32, 0, len(input))
	for shard := range input {
		keys = append(keys, shard)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	vals := make([]string, 0, len(keys))
	for _, key := range keys {
		var shard string
		if key == core.MetachainShardId {
			shard = "meta"
		} else {
			shard = fmt.Sprintf("shard %d", key)
		}

		vals = append(vals, fmt.Sprintf("%s: %d", shard, input[key]))
	}

	return strings.Join(vals, ", ")
}

func (netMes *networkMessenger) checkExternalLoggers() {
	for {
		select {
		case <-netMes.ctx.Done():
			log.Debug("closing networkMessenger.checkExternalLoggers go routine")
			return
		case <-time.After(timeBetweenExternalLoggersCheck):
		}

		setupExternalP2PLoggers()
	}
}

// Close closes the host, connections and streams
func (netMes *networkMessenger) Close() error {
	log.Debug("closing network messenger's host...")

	var err error
	log.Debug("closing network messenger's messages handler...")
	errMH := netMes.MessageHandler.Close()
	if errMH != nil {
		err = errMH
		log.Warn("networkMessenger.Close",
			"component", "messagesHandler",
			"error", err)
	}

	log.Debug("closing network messenger's connections handler...")
	errCH := netMes.ConnectionsHandler.Close()
	if errCH != nil {
		err = errCH
		log.Warn("networkMessenger.Close",
			"component", "connectionsHandler",
			"error", err)
	}

	errHost := netMes.p2pHost.Close()
	if errHost != nil {
		err = errHost
		log.Warn("networkMessenger.Close",
			"component", "host",
			"error", err)
	}

	log.Debug("closing network messenger's print connection watcher...")
	errConnWatcher := netMes.printConnectionsWatcher.Close()
	if errConnWatcher != nil {
		err = errConnWatcher
		log.Warn("networkMessenger.Close",
			"component", "connectionsWatcher",
			"error", err)
	}

	log.Debug("closing network messenger's components through the context...")
	netMes.cancelFunc()

	log.Debug("closing network messenger's peerstore...")
	errPeerStore := netMes.p2pHost.Peerstore().Close()
	if errPeerStore != nil {
		err = errPeerStore
		log.Warn("networkMessenger.Close",
			"component", "peerstore",
			"error", err)
	}

	if err == nil {
		log.Info("network messenger closed successfully")
	}

	return err
}

// ID returns the messenger's ID
func (netMes *networkMessenger) ID() core.PeerID {
	h := netMes.p2pHost

	return core.PeerID(h.ID())
}

// Bootstrap will start the peer discovery mechanism
func (netMes *networkMessenger) Bootstrap() error {
	err := netMes.peerDiscoverer.Bootstrap()
	if err == nil {
		log.Info("started the network discovery process...")
	}
	return err
}

// SetPeerDenialEvaluator sets the peer black list handler
// TODO decide if we continue on using setters or switch to options. Refactor if necessary
func (netMes *networkMessenger) SetPeerDenialEvaluator(handler p2p.PeerDenialEvaluator) error {
	return netMes.connMonitorWrapper.SetPeerDenialEvaluator(handler)
}

// Port returns the port that this network messenger is using
func (netMes *networkMessenger) Port() int {
	return netMes.port
}

// AddPeerTopicNotifier will add a new peer topic notifier
func (netMes *networkMessenger) AddPeerTopicNotifier(notifier p2p.PeerTopicNotifier) error {
	if check.IfNil(notifier) {
		return p2p.ErrNilPeerTopicNotifier
	}

	netMes.mutPeerTopicNotifiers.Lock()
	netMes.peerTopicNotifiers = append(netMes.peerTopicNotifiers, notifier)
	netMes.mutPeerTopicNotifiers.Unlock()

	log.Debug("networkMessenger.AddPeerTopicNotifier", "type", fmt.Sprintf("%T", notifier))

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (netMes *networkMessenger) IsInterfaceNil() bool {
	return netMes == nil
}
