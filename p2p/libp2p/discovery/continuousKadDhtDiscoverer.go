package discovery

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

var _ p2p.PeerDiscoverer = (*continuousKadDhtDiscoverer)(nil)
var _ p2p.Reconnecter = (*continuousKadDhtDiscoverer)(nil)

const kadDhtName = "kad-dht discovery"

// ArgKadDht represents the kad-dht config argument DTO
type ArgKadDht struct {
	Context                     context.Context
	Host                        ConnectableHost
	PeersRefreshInterval        time.Duration
	SeedersReconnectionInterval time.Duration
	ProtocolID                  string
	InitialPeersList            []string
	BucketSize                  uint32
	RoutingTableRefresh         time.Duration
	KddSharder                  p2p.Sharder
	ConnectionWatcher           p2p.ConnectionsWatcher
	Logger                      p2p.Logger
}

// continuousKadDhtDiscoverer is the kad-dht discovery type implementation
// This implementation does not support pausing and resuming of the discovery process
type continuousKadDhtDiscoverer struct {
	host          ConnectableHost
	context       context.Context
	mutKadDht     sync.RWMutex
	kadDHT        *dht.IpfsDHT
	refreshCancel context.CancelFunc

	peersRefreshInterval time.Duration
	protocolID           string
	initialPeersList     []string
	bucketSize           uint32
	routingTableRefresh  time.Duration
	hostConnManagement   *hostWithConnectionManagement
	sharder              Sharder
	connectionWatcher    p2p.ConnectionsWatcher
	log                  p2p.Logger
}

// NewContinuousKadDhtDiscoverer creates a new kad-dht discovery type implementation
// initialPeersList can be nil or empty, no initial connection will be attempted, a warning message will appear
func NewContinuousKadDhtDiscoverer(arg ArgKadDht) (*continuousKadDhtDiscoverer, error) {
	sharder, err := prepareArguments(arg)
	if err != nil {
		return nil, err
	}

	sharder.SetSeeders(arg.InitialPeersList)

	return &continuousKadDhtDiscoverer{
		context:              arg.Context,
		host:                 arg.Host,
		sharder:              sharder,
		peersRefreshInterval: arg.PeersRefreshInterval,
		protocolID:           arg.ProtocolID,
		initialPeersList:     arg.InitialPeersList,
		bucketSize:           arg.BucketSize,
		routingTableRefresh:  arg.RoutingTableRefresh,
		connectionWatcher:    arg.ConnectionWatcher,
		log:                  arg.Logger,
	}, nil
}

func prepareArguments(arg ArgKadDht) (Sharder, error) {
	if check.IfNilReflect(arg.Context) {
		return nil, p2p.ErrNilContext
	}
	if check.IfNilReflect(arg.Host) {
		return nil, p2p.ErrNilHost
	}
	if check.IfNil(arg.KddSharder) {
		return nil, p2p.ErrNilSharder
	}
	if check.IfNil(arg.ConnectionWatcher) {
		return nil, p2p.ErrNilConnectionsWatcher
	}
	sharder, ok := arg.KddSharder.(Sharder)
	if !ok {
		return nil, fmt.Errorf("%w for sharder: expected discovery.Sharder type of interface", p2p.ErrWrongTypeAssertion)
	}
	if arg.PeersRefreshInterval < time.Second {
		return nil, fmt.Errorf("%w, PeersRefreshInterval should have been at least 1 second", p2p.ErrInvalidValue)
	}
	if arg.RoutingTableRefresh < time.Second {
		return nil, fmt.Errorf("%w, RoutingTableRefresh should have been at least 1 second", p2p.ErrInvalidValue)
	}
	if check.IfNil(arg.Logger) {
		return nil, p2p.ErrNilLogger
	}
	isListNilOrEmpty := len(arg.InitialPeersList) == 0
	if isListNilOrEmpty {
		arg.Logger.Warn("nil or empty initial peers list provided to kad dht implementation. " +
			"No initial connection will be done")
	}

	return sharder, nil
}

// Bootstrap will start the bootstrapping new peers process
func (ckdd *continuousKadDhtDiscoverer) Bootstrap() error {
	ckdd.mutKadDht.Lock()
	defer ckdd.mutKadDht.Unlock()

	if ckdd.kadDHT != nil {
		return p2p.ErrPeerDiscoveryProcessAlreadyStarted
	}

	return ckdd.startDHT()
}

func (ckdd *continuousKadDhtDiscoverer) startDHT() error {
	ctxrun, cancel := context.WithCancel(ckdd.context)
	var err error
	args := ArgsHostWithConnectionManagement{
		ConnectableHost:    ckdd.host,
		Sharder:            ckdd.sharder,
		ConnectionsWatcher: ckdd.connectionWatcher,
	}
	ckdd.hostConnManagement, err = NewHostWithConnectionManagement(args)
	if err != nil {
		cancel()
		return err
	}

	protocolID := protocol.ID(ckdd.protocolID)
	kademliaDHT, err := dht.New(
		ckdd.context,
		ckdd.hostConnManagement,
		dht.ProtocolPrefix(protocolID),
		dht.RoutingTableRefreshPeriod(ckdd.routingTableRefresh),
		dht.Mode(dht.ModeServer),
	)
	if err != nil {
		cancel()
		return err
	}

	go ckdd.connectToInitialAndBootstrap(ctxrun)

	ckdd.kadDHT = kademliaDHT
	ckdd.refreshCancel = cancel
	return nil
}

func (ckdd *continuousKadDhtDiscoverer) stopDHT() error {
	if ckdd.refreshCancel == nil {
		return nil
	}

	ckdd.refreshCancel()
	ckdd.refreshCancel = nil

	protocolID := protocol.ID(ckdd.protocolID)
	ckdd.host.RemoveStreamHandler(protocolID)

	err := ckdd.kadDHT.Close()

	ckdd.kadDHT = nil

	return err
}

func (ckdd *continuousKadDhtDiscoverer) connectToInitialAndBootstrap(ctx context.Context) {
	chanStartBootstrap := ckdd.connectToOnePeerFromInitialPeersList(
		ckdd.peersRefreshInterval,
		ckdd.initialPeersList,
	)

	// TODO: needs refactor
	go func() {
		select {
		case <-chanStartBootstrap:
		case <-ctx.Done():
			return
		}
		ckdd.bootstrap(ctx)
	}()
}

func (ckdd *continuousKadDhtDiscoverer) bootstrap(ctx context.Context) {
	ckdd.log.Debug("starting the p2p bootstrapping process")
	for {
		ckdd.mutKadDht.RLock()
		kadDht := ckdd.kadDHT
		ckdd.mutKadDht.RUnlock()

		shouldReconnect := kadDht != nil && kbucket.ErrLookupFailure == kadDht.Bootstrap(ckdd.context)
		if shouldReconnect {
			ckdd.log.Debug("pausing the p2p bootstrapping process")
			ckdd.ReconnectToNetwork(ctx)
			ckdd.log.Debug("resuming the p2p bootstrapping process")
		}

		select {
		case <-time.After(ckdd.peersRefreshInterval):
		case <-ctx.Done():
			ckdd.log.Debug("closing the p2p bootstrapping process")
			return
		}
	}
}

func (ckdd *continuousKadDhtDiscoverer) connectToOnePeerFromInitialPeersList(
	intervalBetweenAttempts time.Duration,
	initialPeersList []string,
) <-chan struct{} {

	chanDone := make(chan struct{}, 1)

	if len(initialPeersList) == 0 {
		chanDone <- struct{}{}
		return chanDone
	}

	go ckdd.tryConnectToSeeder(intervalBetweenAttempts, initialPeersList, chanDone)

	return chanDone
}

func (ckdd *continuousKadDhtDiscoverer) tryConnectToSeeder(
	intervalBetweenAttempts time.Duration,
	initialPeersList []string,
	chanDone chan struct{},
) {

	startIndex := 0

	for {
		initialPeer := initialPeersList[startIndex]
		err := ckdd.host.ConnectToPeer(ckdd.context, initialPeer)
		if err != nil {
			printConnectionErrorToSeeder(initialPeer, err, ckdd.log)
			startIndex++
			startIndex = startIndex % len(initialPeersList)
			select {
			case <-ckdd.context.Done():
				ckdd.log.Debug("context done in continuousKadDhtDiscoverer")
				return
			case <-time.After(intervalBetweenAttempts):
				continue
			}
		} else {
			ckdd.log.Debug("connected to seeder", "address", initialPeer)
		}

		break
	}
	chanDone <- struct{}{}
}

func printConnectionErrorToSeeder(peer string, err error, log p2p.Logger) {
	if errors.Is(err, p2p.ErrUnwantedPeer) {
		log.Trace("tryConnectToSeeder: unwanted peer",
			"seeder", peer,
			"error", err.Error(),
		)

		return
	}

	log.Debug("error connecting to seeder",
		"seeder", peer,
		"error", err.Error(),
	)
}

// Name returns the name of the kad dht peer discovery implementation
func (ckdd *continuousKadDhtDiscoverer) Name() string {
	return kadDhtName
}

// ReconnectToNetwork will try to connect to one peer from the initial peer list
func (ckdd *continuousKadDhtDiscoverer) ReconnectToNetwork(ctx context.Context) {
	select {
	case <-ckdd.connectToOnePeerFromInitialPeersList(ckdd.peersRefreshInterval, ckdd.initialPeersList):
	case <-ctx.Done():
		return
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ckdd *continuousKadDhtDiscoverer) IsInterfaceNil() bool {
	return ckdd == nil
}
