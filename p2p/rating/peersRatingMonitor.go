package rating

import (
	"encoding/json"
	"fmt"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-storage-go/types"
)

const unknownRating = "unknown"

// ArgPeersRatingMonitor is the DTO used to create a new peers rating monitor
type ArgPeersRatingMonitor struct {
	TopRatedCache types.Cacher
	BadRatedCache types.Cacher
}

type peersRatingMonitor struct {
	topRatedCache types.Cacher
	badRatedCache types.Cacher
}

// NewPeersRatingMonitor returns a new peers rating monitor
func NewPeersRatingMonitor(args ArgPeersRatingMonitor) (*peersRatingMonitor, error) {
	err := checkMonitorArgs(args)
	if err != nil {
		return nil, err
	}

	return &peersRatingMonitor{
		topRatedCache: args.TopRatedCache,
		badRatedCache: args.BadRatedCache,
	}, nil
}

func checkMonitorArgs(args ArgPeersRatingMonitor) error {
	if check.IfNil(args.TopRatedCache) {
		return fmt.Errorf("%w for TopRatedCache", p2p.ErrNilCacher)
	}
	if check.IfNil(args.BadRatedCache) {
		return fmt.Errorf("%w for BadRatedCache", p2p.ErrNilCacher)
	}

	return nil
}

// GetConnectedPeersRatings returns the ratings of the current connected peers
func (monitor *peersRatingMonitor) GetConnectedPeersRatings(connectionsHandler p2p.ConnectionsHandler) (string, error) {
	if check.IfNil(connectionsHandler) {
		return "", p2p.ErrNilConnectionsHandler
	}
	connectedPeersRatings := monitor.extractConnectedPeersRatings(connectionsHandler)

	jsonMap, err := json.Marshal(&connectedPeersRatings)
	if err != nil {
		return "", err
	}

	return string(jsonMap), nil
}

func (monitor *peersRatingMonitor) extractConnectedPeersRatings(connectionsHandler p2p.ConnectionsHandler) map[string]string {
	connectedPeers := connectionsHandler.ConnectedPeers()
	connectedPeersRatings := make(map[string]string, len(connectedPeers))
	for _, connectedPeer := range connectedPeers {
		connectedPeersRatings[connectedPeer.Pretty()] = monitor.fetchRating(connectedPeer)
	}

	return connectedPeersRatings
}

func (monitor *peersRatingMonitor) fetchRating(pid core.PeerID) string {
	rating, found := monitor.topRatedCache.Get(pid.Bytes())
	if found {
		return fmt.Sprintf("%d", rating)
	}

	rating, found = monitor.badRatedCache.Get(pid.Bytes())
	if found {
		return fmt.Sprintf("%d", rating)
	}

	return unknownRating
}

// IsInterfaceNil returns true if there is no value under the interface
func (monitor *peersRatingMonitor) IsInterfaceNil() bool {
	return monitor == nil
}
