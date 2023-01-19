package libp2p

import (
	"context"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	p2p "github.com/ElrondNetwork/elrond-go-p2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ArgConnectionsHandler is the DTO struct used to create a new instance of connections handler
type ArgConnectionsHandler struct {
	P2pHost              ConnectableHost
	PeersOnChannel       PeersOnChannel
	PeerShardResolver    p2p.PeerShardResolver
	Sharder              p2p.Sharder
	PreferredPeersHolder p2p.PreferredPeersHolderHandler
	ConnMonitor          ConnectionMonitor
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
}

// NewConnectionsHandler creates a new connections manager
func NewConnectionsHandler(args ArgConnectionsHandler) (*connectionsHandler, error) {
	err := checkArgConnectionsHandler(args)
	if err != nil {
		return nil, err
	}

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
	}, nil
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

	return nil
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
	addrs := make([]string, 0)

	x := handler.p2pHost.Addrs()
	for _, address := range x {
		addrs = append(addrs, address.String()+"/p2p/"+handler.peerID().Pretty())
	}

	return addrs
}

// ConnectToPeer tries to open a new connection to a peer
func (handler *connectionsHandler) ConnectToPeer(address string) error {
	return handler.p2pHost.ConnectToPeer(handler.ctx, address)
}

// WaitForConnections will wait the maxWaitingTime duration or until the target connected peers was achieved
func (handler *connectionsHandler) WaitForConnections(maxWaitingTime time.Duration, minNumOfPeers uint32) {
	startTime := time.Now()
	defer func() {
		log.Debug("connectionsHandler.WaitForConnections",
			"waited", time.Since(startTime), "num connected peers", len(handler.ConnectedPeers()))
	}()

	if minNumOfPeers == 0 {
		log.Debug("connectionsHandler.WaitForConnections", "waiting", maxWaitingTime)
		time.Sleep(maxWaitingTime)
		return
	}

	handler.waitForConnections(maxWaitingTime, minNumOfPeers)
}

func (handler *connectionsHandler) waitForConnections(maxWaitingTime time.Duration, minNumOfPeers uint32) {
	log.Debug("connectionsHandler.WaitForConnections", "waiting", maxWaitingTime, "min num of peers", minNumOfPeers)
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

	connectedPeers := make(map[core.PeerID]struct{})

	for _, conn := range h.Network().Conns() {
		p := core.PeerID(conn.RemotePeer())

		if handler.IsConnected(p) {
			connectedPeers[p] = struct{}{}
		}
	}

	peerList := make([]core.PeerID, len(connectedPeers))

	index := 0
	for k := range connectedPeers {
		peerList[index] = k
		index++
	}

	return peerList
}

// ConnectedAddresses returns all connected peer's addresses
func (handler *connectionsHandler) ConnectedAddresses() []string {
	h := handler.p2pHost
	conns := make([]string, 0)

	for _, c := range h.Network().Conns() {
		conns = append(conns, c.RemoteMultiaddr().String()+"/p2p/"+c.RemotePeer().String())
	}
	return conns
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

// ConnectedFullHistoryPeersOnTopic returns the connected peers on a provided topic
func (handler *connectionsHandler) ConnectedFullHistoryPeersOnTopic(topic string) []core.PeerID {
	peerList := handler.ConnectedPeersOnTopic(topic)
	fullHistoryList := make([]core.PeerID, 0)

	handler.mutPeerResolver.RLock()
	defer handler.mutPeerResolver.RUnlock()

	for _, topicPeer := range peerList {
		peerInfo := handler.peerShardResolver.GetPeerInfo(topicPeer)
		if peerInfo.PeerSubType == core.FullHistoryObserver {
			fullHistoryList = append(fullHistoryList, topicPeer)
		}
	}

	return fullHistoryList
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
		FullHistoryObservers:     make(map[uint32][]string),
		NumObserversOnShard:      make(map[uint32]int),
		NumValidatorsOnShard:     make(map[uint32]int),
		NumPreferredPeersOnShard: make(map[uint32]int),
	}

	handler.mutPeerResolver.RLock()
	defer handler.mutPeerResolver.RUnlock()

	selfPeerInfo := handler.peerShardResolver.GetPeerInfo(handler.peerID())
	connPeerInfo.SelfShardID = selfPeerInfo.ShardID

	for _, p := range peers {
		conns := handler.p2pHost.Network().ConnsToPeer(p)
		connString := "[invalid connection string]"
		if len(conns) > 0 {
			connString = conns[0].RemoteMultiaddr().String() + "/p2p/" + p.String()
		}

		pid := core.PeerID(p)
		peerInfo := handler.peerShardResolver.GetPeerInfo(pid)
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
		if peerInfo.PeerSubType == core.FullHistoryObserver {
			connPeerInfo.FullHistoryObservers[peerInfo.ShardID] = append(connPeerInfo.FullHistoryObservers[peerInfo.ShardID], connString)
			connPeerInfo.NumFullHistoryObservers++
			break
		}
		if selfPeerInfo.ShardID != peerInfo.ShardID {
			connPeerInfo.CrossShardObservers[peerInfo.ShardID] = append(connPeerInfo.CrossShardObservers[peerInfo.ShardID], connString)
			connPeerInfo.NumCrossShardObservers++
			break
		}

		connPeerInfo.IntraShardObservers[peerInfo.ShardID] = append(connPeerInfo.IntraShardObservers[peerInfo.ShardID], connString)
		connPeerInfo.NumIntraShardObservers++
	}
}

func (handler *connectionsHandler) peerID() core.PeerID {
	return core.PeerID(handler.p2pHost.ID())
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

// Close closes the messages handler
func (handler *connectionsHandler) Close() error {
	handler.cancelFunc()

	var err error
	log.Debug("closing connections handler's peers on channel...")
	errPoc := handler.peersOnChannel.Close()
	if errPoc != nil {
		err = errPoc
		log.Warn("connectionsHandler.Close",
			"component", "peersOnChannel",
			"error", errPoc)
	}

	log.Debug("closing connections handler's connection monitor...")
	errConnMonitor := handler.connMonitor.Close()
	if errConnMonitor != nil {
		err = errConnMonitor
		log.Warn("networkMessenger.Close",
			"component", "connMonitor",
			"error", errConnMonitor)
	}

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *connectionsHandler) IsInterfaceNil() bool {
	return handler == nil
}
