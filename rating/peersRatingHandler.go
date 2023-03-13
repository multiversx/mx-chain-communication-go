package rating

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	p2p "github.com/multiversx/mx-chain-p2p-go"
	"github.com/multiversx/mx-chain-storage-go/types"
)

const (
	topRatedTier   = "top rated tier"
	badRatedTier   = "bad rated tier"
	defaultRating  = int32(0)
	minRating      = -100
	maxRating      = 100
	increaseFactor = 2
	decreaseFactor = -1
	minNumOfPeers  = 1
	int32Size      = 4
)

var log = logger.GetOrCreate("p2p/peersRating")

// ArgPeersRatingHandler is the DTO used to create a new peers rating handler
type ArgPeersRatingHandler struct {
	TopRatedCache types.Cacher
	BadRatedCache types.Cacher
}

type peersRatingHandler struct {
	topRatedCache types.Cacher
	badRatedCache types.Cacher
	mut           sync.RWMutex
}

// NewPeersRatingHandler returns a new peers rating handler
func NewPeersRatingHandler(args ArgPeersRatingHandler) (*peersRatingHandler, error) {
	err := checkHandlerArgs(args)
	if err != nil {
		return nil, err
	}

	return &peersRatingHandler{
		topRatedCache: args.TopRatedCache,
		badRatedCache: args.BadRatedCache,
	}, nil
}

func checkHandlerArgs(args ArgPeersRatingHandler) error {
	if check.IfNil(args.TopRatedCache) {
		return fmt.Errorf("%w for TopRatedCache", p2p.ErrNilCacher)
	}
	if check.IfNil(args.BadRatedCache) {
		return fmt.Errorf("%w for BadRatedCache", p2p.ErrNilCacher)
	}

	return nil
}

// IncreaseRating increases the rating of a peer with the increase factor
func (prh *peersRatingHandler) IncreaseRating(pid core.PeerID) {
	// keep this section critical, as we do read + write
	prh.mut.Lock()
	defer prh.mut.Unlock()

	prh.updateRating(pid, increaseFactor)
}

// DecreaseRating decreases the rating of a peer with the decrease factor
func (prh *peersRatingHandler) DecreaseRating(pid core.PeerID) {
	// keep this section critical, as we do read + write
	prh.mut.Lock()
	defer prh.mut.Unlock()

	prh.updateRating(pid, decreaseFactor)
}

func (prh *peersRatingHandler) getOldRating(pid []byte) (int32, bool) {
	oldRating, found := prh.topRatedCache.Get(pid)
	if found {
		oldRatingInt, _ := oldRating.(int32)
		return oldRatingInt, found
	}

	oldRating, found = prh.badRatedCache.Get(pid)
	if found {
		oldRatingInt, _ := oldRating.(int32)
		return oldRatingInt, found
	}

	return defaultRating, found
}

func (prh *peersRatingHandler) updateRating(pid core.PeerID, updateFactor int32) {
	oldRating, found := prh.getOldRating(pid.Bytes())
	if !found {
		// new pid, add it with default rating
		prh.topRatedCache.Put(pid.Bytes(), defaultRating, int32Size)
		return
	}

	newRating := oldRating + updateFactor
	if newRating > maxRating {
		newRating = maxRating
	}

	if newRating < minRating {
		newRating = minRating
	}

	prh.updateRatingCacher(pid, oldRating, newRating)
}

func (prh *peersRatingHandler) updateRatingCacher(pid core.PeerID, oldRating, newRating int32) {
	oldTier := computeRatingTier(oldRating)
	newTier := computeRatingTier(newRating)
	if newTier == oldTier {
		if newTier == topRatedTier {
			prh.topRatedCache.Put(pid.Bytes(), newRating, int32Size)
		} else {
			prh.badRatedCache.Put(pid.Bytes(), newRating, int32Size)
		}

		return
	}

	prh.movePeerToNewTier(newRating, newTier, pid)
}

func computeRatingTier(peerRating int32) string {
	if peerRating >= defaultRating {
		return topRatedTier
	}

	return badRatedTier
}

func (prh *peersRatingHandler) movePeerToNewTier(newRating int32, newTier string, pid core.PeerID) {
	if newTier == topRatedTier {
		prh.badRatedCache.Remove(pid.Bytes())
		prh.topRatedCache.Put(pid.Bytes(), newRating, int32Size)
	} else {
		prh.topRatedCache.Remove(pid.Bytes())
		prh.badRatedCache.Put(pid.Bytes(), newRating, int32Size)
	}
}

// GetTopRatedPeersFromList returns a list of peers, searching them in the order of rating tiers
func (prh *peersRatingHandler) GetTopRatedPeersFromList(peers []core.PeerID, minNumOfPeersExpected int) []core.PeerID {
	prh.mut.Lock()
	defer prh.mut.Unlock()

	peersTopRated := make([]core.PeerID, 0)
	defer prh.displayPeersRating(&peersTopRated, minNumOfPeersExpected)

	isListEmpty := len(peers) == 0
	if minNumOfPeersExpected < minNumOfPeers || isListEmpty {
		return make([]core.PeerID, 0)
	}

	peersTopRated, peersBadRated := prh.splitPeersByTiers(peers)
	if len(peersTopRated) < minNumOfPeersExpected {
		peersTopRated = append(peersTopRated, peersBadRated...)
	}

	return peersTopRated
}

func (prh *peersRatingHandler) displayPeersRating(peers *[]core.PeerID, minNumOfPeersExpected int) {
	if log.GetLevel() != logger.LogTrace {
		return
	}

	strPeersRatings := ""
	for _, peer := range *peers {
		rating, ok := prh.topRatedCache.Get(peer.Bytes())
		if !ok {
			rating, _ = prh.badRatedCache.Get(peer.Bytes())
		}

		ratingInt, ok := rating.(int32)
		if ok {
			strPeersRatings += fmt.Sprintf("\n peerID: %s, rating: %d", peer.Pretty(), ratingInt)
		} else {
			strPeersRatings += fmt.Sprintf("\n peerID: %s, rating: invalid", peer.Pretty())
		}
	}

	log.Trace("Best peers to request from", "min requested", minNumOfPeersExpected, "peers ratings", strPeersRatings)
}

func (prh *peersRatingHandler) splitPeersByTiers(peers []core.PeerID) ([]core.PeerID, []core.PeerID) {
	topRated := make([]core.PeerID, 0)
	badRated := make([]core.PeerID, 0)

	for _, peer := range peers {
		if prh.topRatedCache.Has(peer.Bytes()) {
			topRated = append(topRated, peer)
		}

		if prh.badRatedCache.Has(peer.Bytes()) {
			badRated = append(badRated, peer)
		}
	}

	return topRated, badRated
}

// IsInterfaceNil returns true if there is no value under the interface
func (prh *peersRatingHandler) IsInterfaceNil() bool {
	return prh == nil
}
