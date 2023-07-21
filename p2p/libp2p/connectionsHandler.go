package libp2p

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const timeBetweenPeerPrints = time.Second * 20

// ArgConnectionsHandler is the DTO struct used to create a new instance of connections handler
type ArgConnectionsHandler struct {
	P2pHost              ConnectableHost
	PeersOnChannel       PeersOnChannel
	PeerShardResolver    p2p.PeerShardResolver
	Sharder              p2p.Sharder
	PreferredPeersHolder p2p.PreferredPeersHolderHandler
	ConnMonitor          ConnectionMonitor
	PeerDiscoverer       p2p.PeerDiscoverer
	PeerID               core.PeerID
	ConnectionsMetric    ConnectionsMetric
	NetworkType          p2p.NetworkType
	Logger               p2p.Logger
}

type connectionsHandler struct {
	ctx                  context.Context
	cancelFunc           context.CancelFunc
	p2pHost              ConnectableHost
	peersOnChannel       PeersOnChannel
	mutPeerResolver      sync.RWMutex
	peerShardResolver    p2p.PeerShardResolver
	sharder              p2p.Sharder
	preferredPeersHolder p2p.PreferredPeersHolderHandler
	connMonitor          ConnectionMonitor
	peerDiscoverer       p2p.PeerDiscoverer
	peerID               core.PeerID
	connectionsMetric    ConnectionsMetric
	networkType          p2p.NetworkType
	log                  p2p.Logger
}

// NewConnectionsHandler creates a new connections manager
func NewConnectionsHandler(args ArgConnectionsHandler) (*connectionsHandler, error) {
	err := checkArgConnectionsHandler(args)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	handler := &connectionsHandler{
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
		log:                  args.Logger,
		networkType:          args.NetworkType,
	}

	go handler.printLogs()

	return handler, nil
}

func checkArgConnectionsHandler(args ArgConnectionsHandler) error {
	if check.IfNil(args.P2pHost) {
		return p2p.ErrNilHost
	}
	if check.IfNil(args.PeersOnChannel) {
		return p2p.ErrNilPeersOnChannel
	}
	if check.IfNil(args.PeerShardResolver) {
		return p2p.ErrNilPeerShardResolver
	}
	if check.IfNil(args.Sharder) {
		return p2p.ErrNilSharder
	}
	if check.IfNil(args.PreferredPeersHolder) {
		return p2p.ErrNilPreferredPeersHolder
	}
	if check.IfNil(args.ConnMonitor) {
		return p2p.ErrNilConnectionMonitor
	}
	if check.IfNil(args.PeerDiscoverer) {
		return p2p.ErrNilPeerDiscoverer
	}
	if check.IfNil(args.ConnectionsMetric) {
		return p2p.ErrNilConnectionsMetric
	}
	if check.IfNil(args.Logger) {
		return p2p.ErrNilLogger
	}

	return nil
}

// Bootstrap will start the peer discovery mechanism
func (handler *connectionsHandler) Bootstrap() error {
	err := handler.peerDiscoverer.Bootstrap()
	if err == nil {
		handler.log.Info("started the network discovery process...")
	}
	return err
}

// Peers returns the list of all known peers ID (including self)
func (handler *connectionsHandler) Peers() []core.PeerID {
	peers := make([]core.PeerID, 0)

	for _, p := range handler.p2pHost.Peerstore().Peers() {
		peers = append(peers, core.PeerID(p))
	}
	return peers
}

// Addresses returns all addresses found in peerstore
func (handler *connectionsHandler) Addresses() []string {
	addresses := handler.p2pHost.Addrs()
	result := make([]string, 0, len(addresses))
	for _, address := range addresses {
		result = append(result, address.String()+"/p2p/"+handler.peerID.Pretty())
	}

	return result
}

// ConnectToPeer tries to open a new connection to a peer
func (handler *connectionsHandler) ConnectToPeer(address string) error {
	return handler.p2pHost.ConnectToPeer(handler.ctx, address)
}

// WaitForConnections will wait the maxWaitingTime duration or until the target connected peers was achieved
func (handler *connectionsHandler) WaitForConnections(maxWaitingTime time.Duration, minNumOfPeers uint32) {
	startTime := time.Now()
	defer func() {
		handler.log.Debug("connectionsHandler.WaitForConnections",
			"waited", time.Since(startTime), "num connected peers", len(handler.ConnectedPeers()))
	}()

	if minNumOfPeers == 0 {
		handler.log.Debug("connectionsHandler.WaitForConnections", "waiting", maxWaitingTime)
		time.Sleep(maxWaitingTime)
		return
	}

	handler.waitForConnections(maxWaitingTime, minNumOfPeers)
}

func (handler *connectionsHandler) waitForConnections(maxWaitingTime time.Duration, minNumOfPeers uint32) {
	handler.log.Debug("connectionsHandler.WaitForConnections", "waiting", maxWaitingTime, "min num of peers", minNumOfPeers)
	ctxMaxWaitingTime, cancel := context.WithTimeout(context.Background(), maxWaitingTime)
	defer cancel()

	for {
		if handler.shouldStopWaiting(ctxMaxWaitingTime, minNumOfPeers) {
			return
		}
	}
}

func (handler *connectionsHandler) shouldStopWaiting(ctxMaxWaitingTime context.Context, minNumOfPeers uint32) bool {
	ctx, cancel := context.WithTimeout(context.Background(), pollWaitForConnectionsInterval)
	defer cancel()

	select {
	case <-ctxMaxWaitingTime.Done():
		return true
	case <-ctx.Done():
		return int(minNumOfPeers) <= len(handler.ConnectedPeers())
	}
}

// IsConnected returns true if current node is connected to provided peer
func (handler *connectionsHandler) IsConnected(peerID core.PeerID) bool {
	h := handler.p2pHost

	connectedness := h.Network().Connectedness(peer.ID(peerID))

	return connectedness == network.Connected
}

// ConnectedPeers returns the current connected peers list
func (handler *connectionsHandler) ConnectedPeers() []core.PeerID {
	h := handler.p2pHost

	connections := h.Network().Conns()
	result := make([]core.PeerID, 0, len(connections))
	connectedPeersMap := make(map[core.PeerID]struct{})
	for _, conn := range connections {
		p := core.PeerID(conn.RemotePeer())
		if !handler.IsConnected(p) {
			continue
		}

		_, alreadyAdded := connectedPeersMap[p]
		if alreadyAdded {
			continue
		}

		connectedPeersMap[p] = struct{}{}
		result = append(result, p)
	}

	return result
}

// ConnectedAddresses returns all connected peer's addresses
func (handler *connectionsHandler) ConnectedAddresses() []string {
	h := handler.p2pHost

	connections := h.Network().Conns()
	addresses := make([]string, 0, len(connections))
	for _, c := range connections {
		addresses = append(addresses, c.RemoteMultiaddr().String()+"/p2p/"+c.RemotePeer().String())
	}
	return addresses
}

// PeerAddresses returns the peer's addresses or empty slice if the peer is unknown
func (handler *connectionsHandler) PeerAddresses(pid core.PeerID) []string {
	h := handler.p2pHost

	result := make([]string, 0)

	// check if the peer is connected to the node and append its address
	for _, c := range h.Network().Conns() {
		if string(c.RemotePeer()) == string(pid.Bytes()) {
			result = append(result, c.RemoteMultiaddr().String())
			break
		}
	}

	// check in peerstore (maybe it is known but not connected)
	addresses := h.Peerstore().Addrs(peer.ID(pid.Bytes()))
	for _, addr := range addresses {
		result = append(result, addr.String())
	}

	return result
}

// ConnectedPeersOnTopic returns the connected peers on a provided topic
func (handler *connectionsHandler) ConnectedPeersOnTopic(topic string) []core.PeerID {
	return handler.peersOnChannel.ConnectedPeersOnChannel(topic)
}

// SetPeerShardResolver sets the peer shard resolver component that is able to resolve the link
// between peerID and shardId
func (handler *connectionsHandler) SetPeerShardResolver(peerShardResolver p2p.PeerShardResolver) error {
	if check.IfNil(peerShardResolver) {
		return p2p.ErrNilPeerShardResolver
	}

	err := handler.sharder.SetPeerShardResolver(peerShardResolver)
	if err != nil {
		return err
	}

	handler.mutPeerResolver.Lock()
	handler.peerShardResolver = peerShardResolver
	handler.mutPeerResolver.Unlock()

	return nil
}

// GetConnectedPeersInfo gets the current connected peers information
func (handler *connectionsHandler) GetConnectedPeersInfo() *p2p.ConnectedPeersInfo {
	peers := handler.p2pHost.Network().Peers()
	connPeerInfo := &p2p.ConnectedPeersInfo{
		UnknownPeers:             make([]string, 0),
		Seeders:                  make([]string, 0),
		IntraShardValidators:     make(map[uint32][]string),
		IntraShardObservers:      make(map[uint32][]string),
		CrossShardValidators:     make(map[uint32][]string),
		CrossShardObservers:      make(map[uint32][]string),
		NumObserversOnShard:      make(map[uint32]int),
		NumValidatorsOnShard:     make(map[uint32]int),
		NumPreferredPeersOnShard: make(map[uint32]int),
	}

	handler.mutPeerResolver.RLock()
	peerShardResolver := handler.peerShardResolver
	handler.mutPeerResolver.RUnlock()

	selfPeerInfo := peerShardResolver.GetPeerInfo(handler.peerID)
	connPeerInfo.SelfShardID = selfPeerInfo.ShardID

	for _, p := range peers {
		conns := handler.p2pHost.Network().ConnsToPeer(p)
		connString := "[invalid connection string]"
		if len(conns) > 0 {
			connString = conns[0].RemoteMultiaddr().String() + "/p2p/" + p.String()
		}

		pid := core.PeerID(p)
		peerInfo := peerShardResolver.GetPeerInfo(pid)
		handler.appendPeerInfo(peerInfo, selfPeerInfo, pid, connPeerInfo, connString)

		if handler.preferredPeersHolder.Contains(pid) {
			connPeerInfo.NumPreferredPeersOnShard[peerInfo.ShardID]++
		}
	}

	return connPeerInfo
}

func (handler *connectionsHandler) appendPeerInfo(peerInfo, selfPeerInfo core.P2PPeerInfo, pid core.PeerID, connPeerInfo *p2p.ConnectedPeersInfo, connString string) {
	switch peerInfo.PeerType {
	case core.UnknownPeer:
		if handler.sharder.IsSeeder(pid) {
			connPeerInfo.Seeders = append(connPeerInfo.Seeders, connString)
		} else {
			connPeerInfo.UnknownPeers = append(connPeerInfo.UnknownPeers, connString)
		}
	case core.ValidatorPeer:
		connPeerInfo.NumValidatorsOnShard[peerInfo.ShardID]++
		if selfPeerInfo.ShardID != peerInfo.ShardID {
			connPeerInfo.CrossShardValidators[peerInfo.ShardID] = append(connPeerInfo.CrossShardValidators[peerInfo.ShardID], connString)
			connPeerInfo.NumCrossShardValidators++
		} else {
			connPeerInfo.IntraShardValidators[peerInfo.ShardID] = append(connPeerInfo.IntraShardValidators[peerInfo.ShardID], connString)
			connPeerInfo.NumIntraShardValidators++
		}
	case core.ObserverPeer:
		connPeerInfo.NumObserversOnShard[peerInfo.ShardID]++
		if selfPeerInfo.ShardID != peerInfo.ShardID {
			connPeerInfo.CrossShardObservers[peerInfo.ShardID] = append(connPeerInfo.CrossShardObservers[peerInfo.ShardID], connString)
			connPeerInfo.NumCrossShardObservers++
			break
		}

		connPeerInfo.IntraShardObservers[peerInfo.ShardID] = append(connPeerInfo.IntraShardObservers[peerInfo.ShardID], connString)
		connPeerInfo.NumIntraShardObservers++
	}
}

// IsConnectedToTheNetwork returns true if the current node is connected to the network
func (handler *connectionsHandler) IsConnectedToTheNetwork() bool {
	netw := handler.p2pHost.Network()
	return handler.connMonitor.IsConnectedToTheNetwork(netw)
}

// SetThresholdMinConnectedPeers sets the minimum connected peers before triggering a new reconnection
func (handler *connectionsHandler) SetThresholdMinConnectedPeers(minConnectedPeers int) error {
	if minConnectedPeers < 0 {
		return p2p.ErrInvalidValue
	}

	netw := handler.p2pHost.Network()
	handler.connMonitor.SetThresholdMinConnectedPeers(minConnectedPeers, netw)

	return nil
}

// ThresholdMinConnectedPeers returns the minimum connected peers before triggering a new reconnection
func (handler *connectionsHandler) ThresholdMinConnectedPeers() int {
	return handler.connMonitor.ThresholdMinConnectedPeers()
}

// SetPeerDenialEvaluator sets the peer black list handler
func (handler *connectionsHandler) SetPeerDenialEvaluator(peerDenialEvaluator p2p.PeerDenialEvaluator) error {
	return handler.connMonitor.SetPeerDenialEvaluator(peerDenialEvaluator)
}

// Close closes the messages handler
func (handler *connectionsHandler) Close() error {
	handler.cancelFunc()

	var err error
	handler.log.Debug("closing connections handler's peers on channel...")
	errPoc := handler.peersOnChannel.Close()
	if errPoc != nil {
		err = errPoc
		handler.log.Warn("connectionsHandler.Close",
			"component", "peersOnChannel",
			"error", errPoc)
	}

	handler.log.Debug("closing connections handler's connection monitor...")
	errConnMonitor := handler.connMonitor.Close()
	if errConnMonitor != nil {
		err = errConnMonitor
		handler.log.Warn("connectionsHandler.Close",
			"component", "connMonitor",
			"error", errConnMonitor)
	}

	return err
}

func (handler *connectionsHandler) printLogs() {
	timer := time.NewTimer(timeBetweenPeerPrints)
	defer timer.Stop()

	for {
		timer.Reset(timeBetweenPeerPrints)
		select {
		case <-handler.ctx.Done():
			handler.log.Debug("closing connectionsHandler.printLogsStats go routine")
			return
		case <-timer.C:
			handler.printConnectionsStatus()
		}
	}
}

func (handler *connectionsHandler) printConnectionsStatus() {
	conns := handler.connectionsMetric.ResetNumConnections()
	disconns := handler.connectionsMetric.ResetNumDisconnections()

	peersInfo := handler.GetConnectedPeersInfo()
	handler.log.Debug("network connection status",
		"network", handler.networkType,
		"known peers", len(handler.Peers()),
		"connected peers", len(handler.ConnectedPeers()),
		"intra shard validators", peersInfo.NumIntraShardValidators,
		"intra shard observers", peersInfo.NumIntraShardObservers,
		"cross shard validators", peersInfo.NumCrossShardValidators,
		"cross shard observers", peersInfo.NumCrossShardObservers,
		"unknown", len(peersInfo.UnknownPeers),
		"seeders", len(peersInfo.Seeders),
		"current shard", peersInfo.SelfShardID,
		"validators histogram", handler.mapHistogram(peersInfo.NumValidatorsOnShard),
		"observers histogram", handler.mapHistogram(peersInfo.NumObserversOnShard),
		"preferred peers histogram", handler.mapHistogram(peersInfo.NumPreferredPeersOnShard),
	)

	connsPerSec := conns / uint32(timeBetweenPeerPrints/time.Second)
	disconnsPerSec := disconns / uint32(timeBetweenPeerPrints/time.Second)

	handler.log.Debug("network connection metrics",
		"network", handler.networkType,
		"connections/s", connsPerSec,
		"disconnections/s", disconnsPerSec,
		"connections", conns,
		"disconnections", disconns,
		"time", timeBetweenPeerPrints,
	)
}

func (handler *connectionsHandler) mapHistogram(input map[uint32]int) string {
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

// IsInterfaceNil returns true if there is no value under the interface
func (handler *connectionsHandler) IsInterfaceNil() bool {
	return handler == nil
}
