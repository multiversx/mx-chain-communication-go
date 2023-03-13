package rating

import (
	"encoding/json"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	p2p "github.com/multiversx/mx-chain-p2p-go"
	"github.com/multiversx/mx-chain-storage-go/types"
)

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

// GetPeersRatings returns the current ratings
func (monitor *peersRatingMonitor) GetPeersRatings() string {
	ratingsMap := getRatings(monitor.topRatedCache)
	badRatings := getRatings(monitor.badRatedCache)
	ratingsMap = appendMaps(ratingsMap, badRatings)

	jsonMap, err := json.Marshal(&ratingsMap)
	if err != nil {
		return ""
	}

	return string(jsonMap)
}

func getRatings(cache types.Cacher) map[string]int32 {
	keys := cache.Keys()
	ratingsMap := make(map[string]int32, len(keys))
	for _, pidBytes := range keys {
		rating, found := cache.Get(pidBytes)
		if !found {
			continue
		}

		intRating, _ := rating.(int32)
		pid := core.PeerID(pidBytes)
		ratingsMap[pid.Pretty()] = intRating
	}

	return ratingsMap
}

func appendMaps(m1, m2 map[string]int32) map[string]int32 {
	for key, val := range m2 {
		m1[key] = val
	}
	return m1
}

// IsInterfaceNil returns true if there is no value under the interface
func (monitor *peersRatingMonitor) IsInterfaceNil() bool {
	return monitor == nil
}
