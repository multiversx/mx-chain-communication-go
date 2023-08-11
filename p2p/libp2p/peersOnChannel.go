package libp2p

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// peersOnChannel manages peers on topics
// it buffers the data and refresh the peers list continuously (in refreshInterval intervals)
type peersOnChannel struct {
	mutPeers    sync.RWMutex
	peers       map[string][]core.PeerID
	lastUpdated map[string]time.Time

	refreshInterval   time.Duration
	ttlInterval       time.Duration
	fetchPeersHandler func(topic string) []peer.ID
	getTimeHandler    func() time.Time
	cancelFunc        context.CancelFunc
	log               p2p.Logger
}

// newPeersOnChannel returns a new peersOnChannel object
func newPeersOnChannel(
	fetchPeersHandler func(topic string) []peer.ID,
	refreshInterval time.Duration,
	ttlInterval time.Duration,
	logger p2p.Logger,
) (*peersOnChannel, error) {

	if fetchPeersHandler == nil {
		return nil, p2p.ErrNilFetchPeersOnTopicHandler
	}
	if refreshInterval == 0 {
		return nil, p2p.ErrInvalidDurationProvided
	}
	if ttlInterval == 0 {
		return nil, p2p.ErrInvalidDurationProvided
	}
	if check.IfNil(logger) {
		return nil, p2p.ErrNilLogger
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	poc := &peersOnChannel{
		peers:             make(map[string][]core.PeerID),
		lastUpdated:       make(map[string]time.Time),
		refreshInterval:   refreshInterval,
		ttlInterval:       ttlInterval,
		fetchPeersHandler: fetchPeersHandler,
		cancelFunc:        cancelFunc,
		log:               logger,
	}
	poc.getTimeHandler = poc.clockTime

	go poc.refreshPeersOnAllKnownTopics(ctx)

	return poc, nil
}

func (poc *peersOnChannel) clockTime() time.Time {
	return time.Now()
}

// ConnectedPeersOnChannel returns the known peers on a topic
// if the list was not initialized, it will trigger a manual fetch
func (poc *peersOnChannel) ConnectedPeersOnChannel(topic string) []core.PeerID {
	poc.mutPeers.RLock()
	peers := poc.peers[topic]
	poc.mutPeers.RUnlock()

	if peers != nil {
		return peers
	}

	return poc.refreshPeersOnTopic(topic)
}

// updateConnectedPeersOnTopic updates the connected peers on a topic and the last update timestamp
func (poc *peersOnChannel) updateConnectedPeersOnTopic(topic string, connectedPeers []core.PeerID) {
	poc.mutPeers.Lock()
	poc.peers[topic] = connectedPeers
	poc.lastUpdated[topic] = poc.getTimeHandler()
	poc.mutPeers.Unlock()
}

// refreshPeersOnAllKnownTopics iterates each topic, fetching its last timestamp
// it the timestamp + ttlInterval < time.Now, will trigger a fetch of connected peers on topic
func (poc *peersOnChannel) refreshPeersOnAllKnownTopics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			poc.log.Debug("refreshPeersOnAllKnownTopics's go routine is stopping...")
			return
		case <-time.After(poc.refreshInterval):
		}

		poc.log.Trace("peersOnChannel.refreshPeersOnAllKnownTopics - check")

		listTopicsToBeRefreshed := make([]string, 0)

		// build required topic list
		poc.mutPeers.RLock()
		for topic, lastRefreshed := range poc.lastUpdated {
			needsToBeRefreshed := poc.getTimeHandler().Sub(lastRefreshed) > poc.ttlInterval
			if needsToBeRefreshed {
				listTopicsToBeRefreshed = append(listTopicsToBeRefreshed, topic)
			}
		}
		poc.mutPeers.RUnlock()

		poc.log.Trace("peersOnChannel.refreshPeersOnAllKnownTopics", "listTopicsToBeRefreshed", strings.Join(listTopicsToBeRefreshed, ", "))

		for _, topic := range listTopicsToBeRefreshed {
			_ = poc.refreshPeersOnTopic(topic)
		}
	}
}

// refreshPeersOnTopic
func (poc *peersOnChannel) refreshPeersOnTopic(topic string) []core.PeerID {
	list := poc.fetchPeersHandler(topic)
	connectedPeers := make([]core.PeerID, len(list))
	peers := make([]string, 0, len(list))
	for i, pid := range list {
		peerID := core.PeerID(pid)
		connectedPeers[i] = peerID
		peers = append(peers, peerID.Pretty())
	}

	poc.updateConnectedPeersOnTopic(topic, connectedPeers)

	poc.log.Trace("refreshed peers on topic", "topic", topic, "connected peers", strings.Join(peers, ", "))

	return connectedPeers
}

// Close closes all underlying components
func (poc *peersOnChannel) Close() error {
	poc.cancelFunc()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (poc *peersOnChannel) IsInterfaceNil() bool {
	return poc == nil
}
