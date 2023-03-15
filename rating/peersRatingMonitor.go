package rating

import (
	"encoding/json"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	p2p "github.com/multiversx/mx-chain-p2p-go"
	"github.com/multiversx/mx-chain-storage-go/types"
)

const unknownRating = "unknown"

// ArgPeersRatingMonitor is the DTO used to create a new peers rating monitor
type ArgPeersRatingMonitor struct {
	TopRatedCache       types.Cacher
	BadRatedCache       types.Cacher
	ConnectionsProvider ConnectionsProvider
}

type peersRatingMonitor struct {
	topRatedCache       types.Cacher
	badRatedCache       types.Cacher
	ConnectionsProvider ConnectionsProvider
}

// NewPeersRatingMonitor returns a new peers rating monitor
func NewPeersRatingMonitor(args ArgPeersRatingMonitor) (*peersRatingMonitor, error) {
	err := checkMonitorArgs(args)
	if err != nil {
		return nil, err
	}

	return &peersRatingMonitor{
		topRatedCache:       args.TopRatedCache,
		badRatedCache:       args.BadRatedCache,
		ConnectionsProvider: args.ConnectionsProvider,
	}, nil
}

func checkMonitorArgs(args ArgPeersRatingMonitor) error {
	if check.IfNil(args.TopRatedCache) {
		return fmt.Errorf("%w for TopRatedCache", p2p.ErrNilCacher)
	}
	if check.IfNil(args.BadRatedCache) {
		return fmt.Errorf("%w for BadRatedCache", p2p.ErrNilCacher)
	}
	if check.IfNil(args.ConnectionsProvider) {
		return p2p.ErrNilConnectionsProvider
	}

	return nil
}

// GetConnectedPeersRatings returns the ratings of the current connected peers
func (monitor *peersRatingMonitor) GetConnectedPeersRatings() string {
	ratingsMap := getRatings(monitor.topRatedCache)
	badRatings := getRatings(monitor.badRatedCache)
	ratingsMap = appendMaps(ratingsMap, badRatings)
	connectedPeersRatings := monitor.extractConnectedPeersRatings(ratingsMap)

	jsonMap, err := json.Marshal(&connectedPeersRatings)
	if err != nil {
		return ""
	}

	return string(jsonMap)
}

func getRatings(cache types.Cacher) map[string]string {
	keys := cache.Keys()
	ratingsMap := make(map[string]string, len(keys))
	for _, pidBytes := range keys {
		rating, found := cache.Get(pidBytes)
		if !found {
			continue
		}

		pid := core.PeerID(pidBytes)
		ratingsMap[pid.Pretty()] = fmt.Sprintf("%d", rating)
	}

	return ratingsMap
}

func appendMaps(m1, m2 map[string]string) map[string]string {
	for key, val := range m2 {
		m1[key] = val
	}
	return m1
}

func (monitor *peersRatingMonitor) extractConnectedPeersRatings(allRatings map[string]string) map[string]string {
	connectedPeers := monitor.ConnectionsProvider.ConnectedPeers()
	connectedPeersRatings := make(map[string]string, len(connectedPeers))
	for _, connectedPeer := range connectedPeers {
		prettyPid := connectedPeer.Pretty()
		rating, found := allRatings[prettyPid]
		if found {
			connectedPeersRatings[prettyPid] = rating
			continue
		}

		connectedPeersRatings[prettyPid] = unknownRating
	}

	return connectedPeersRatings
}

// IsInterfaceNil returns true if there is no value under the interface
func (monitor *peersRatingMonitor) IsInterfaceNil() bool {
	return monitor == nil
}
