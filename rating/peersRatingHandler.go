package rating

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-p2p"
	"github.com/ElrondNetwork/elrond-go-storage/types"
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
	int64Size      = 8
	minDuration    = time.Second
)

var log = logger.GetOrCreate("p2p/peersRatingHandler")

// ArgPeersRatingHandler is the DTO used to create a new peers rating handler
type ArgPeersRatingHandler struct {
	TopRatedCache              types.Cacher
	BadRatedCache              types.Cacher
	MarkedForRemovalCache      types.Cacher
	AppStatusHandler           core.AppStatusHandler
	TimeWaitingForReconnection time.Duration
	TimeBetweenMetricsUpdate   time.Duration
	TimeBetweenCachersSweep    time.Duration
}

type ratingInfo struct {
	Rating                       int32 `json:"rating"`
	TimestampLastRequestToPid    int64 `json:"timestampLastRequestToPid"`
	TimestampLastResponseFromPid int64 `json:"timestampLastResponseFromPid"`
}

type peersRatingHandler struct {
	topRatedCache               types.Cacher
	badRatedCache               types.Cacher
	markedForRemovalCache       types.Cacher
	mut                         sync.RWMutex
	ratingsMap                  map[string]*ratingInfo
	appStatusHandler            core.AppStatusHandler
	timeWaitingForReconnection  time.Duration
	timeBetweenMetricsUpdate    time.Duration
	timeBetweenCachersSweep     time.Duration
	getTimeHandler              func() time.Time
	cancelMetricsUpdateLoopFunc func()
	cancelSweepLoopFunc         func()
}

// NewPeersRatingHandler returns a new peers rating handler
func NewPeersRatingHandler(args ArgPeersRatingHandler) (*peersRatingHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	prh := &peersRatingHandler{
		topRatedCache:              args.TopRatedCache,
		badRatedCache:              args.BadRatedCache,
		markedForRemovalCache:      args.MarkedForRemovalCache,
		appStatusHandler:           args.AppStatusHandler,
		timeWaitingForReconnection: args.TimeWaitingForReconnection,
		timeBetweenMetricsUpdate:   args.TimeBetweenMetricsUpdate,
		timeBetweenCachersSweep:    args.TimeBetweenCachersSweep,
		ratingsMap:                 make(map[string]*ratingInfo),
		getTimeHandler:             time.Now,
	}

	var ctxMetricsUpdate, ctxSweepCachers context.Context
	ctxMetricsUpdate, prh.cancelMetricsUpdateLoopFunc = context.WithCancel(context.Background())
	go prh.updateMetricsLoop(ctxMetricsUpdate)

	ctxSweepCachers, prh.cancelSweepLoopFunc = context.WithCancel(context.Background())
	go prh.sweepCachersLoop(ctxSweepCachers)

	return prh, nil
}

func checkArgs(args ArgPeersRatingHandler) error {
	if check.IfNil(args.TopRatedCache) {
		return fmt.Errorf("%w for TopRatedCache", p2p.ErrNilCacher)
	}
	if check.IfNil(args.BadRatedCache) {
		return fmt.Errorf("%w for BadRatedCache", p2p.ErrNilCacher)
	}
	if check.IfNil(args.MarkedForRemovalCache) {
		return fmt.Errorf("%w for MarkedForRemovalCache", p2p.ErrNilCacher)
	}
	if check.IfNil(args.AppStatusHandler) {
		return p2p.ErrNilAppStatusHandler
	}
	if args.TimeWaitingForReconnection < minDuration {
		return fmt.Errorf("%w for TimeWaitingForReconnection, received %f, min expected %d",
			p2p.ErrInvalidValue, args.TimeWaitingForReconnection.Seconds(), minDuration)
	}
	if args.TimeBetweenMetricsUpdate < minDuration {
		return fmt.Errorf("%w for TimeBetweenMetricsUpdate, received %f, min expected %d",
			p2p.ErrInvalidValue, args.TimeBetweenMetricsUpdate.Seconds(), minDuration)
	}
	if args.TimeBetweenCachersSweep < minDuration {
		return fmt.Errorf("%w for TimeBetweenCachersSweep, received %f, min expected %d",
			p2p.ErrInvalidValue, args.TimeBetweenCachersSweep.Seconds(), minDuration)
	}

	return nil
}

// AddPeers adds a new list of peers to the cache with rating 0
// this is called when peers list is refreshed
func (prh *peersRatingHandler) AddPeers(pids []core.PeerID) {
	prh.mut.Lock()
	defer prh.mut.Unlock()

	receivedPIDsMap := make(map[string]struct{}, len(pids))
	for _, pid := range pids {
		pidBytes := pid.Bytes()
		receivedPIDsMap[string(pidBytes)] = struct{}{}

		if prh.markedForRemovalCache.Has(pidBytes) {
			prh.markedForRemovalCache.Remove(pidBytes)
		}

		_, found := prh.getOldRating(pidBytes)
		if found {
			continue
		}

		prh.topRatedCache.Put(pidBytes, defaultRating, int32Size)
		prh.updateRatingsMap(pid, defaultRating, 0)
	}

	prh.markInactivePIDsForRemoval(receivedPIDsMap)
}

// IncreaseRating increases the rating of a peer with the increase factor
func (prh *peersRatingHandler) IncreaseRating(pid core.PeerID) {
	prh.mut.Lock()
	defer prh.mut.Unlock()

	prh.updateRatingIfNeeded(pid, increaseFactor)
}

// DecreaseRating decreases the rating of a peer with the decrease factor
func (prh *peersRatingHandler) DecreaseRating(pid core.PeerID) {
	prh.mut.Lock()
	defer prh.mut.Unlock()

	prh.updateRatingIfNeeded(pid, decreaseFactor)
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

func (prh *peersRatingHandler) markInactivePIDsForRemoval(receivedPIDs map[string]struct{}) {
	removalTimestamp := prh.getTimeHandler().Add(prh.timeWaitingForReconnection).Unix()
	topRatedPIDs := prh.topRatedCache.Keys()
	for _, pid := range topRatedPIDs {
		_, isPIDStillActive := receivedPIDs[string(pid)]
		if !isPIDStillActive {
			prh.markedForRemovalCache.Put(pid, removalTimestamp, int64Size)
		}
	}
	badRatedPIDs := prh.badRatedCache.Keys()
	for _, pid := range badRatedPIDs {
		_, isPIDStillActive := receivedPIDs[string(pid)]
		if !isPIDStillActive {
			prh.markedForRemovalCache.Put(pid, removalTimestamp, int64Size)
		}
	}
}

func (prh *peersRatingHandler) sweepCachersLoop(ctx context.Context) {
	timer := time.NewTimer(prh.timeBetweenCachersSweep)
	defer timer.Stop()

	for {
		timer.Reset(prh.timeBetweenCachersSweep)
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		prh.sweepCachers()
	}
}

func (prh *peersRatingHandler) sweepCachers() {
	prh.mut.Lock()
	defer prh.mut.Unlock()

	pids := prh.markedForRemovalCache.Keys()
	for _, pidBytes := range pids {
		value, exists := prh.markedForRemovalCache.Get(pidBytes)
		if !exists {
			continue
		}

		removalTimestamp, ok := value.(int64)
		if !ok {
			log.Error("could not parse data to int64")
			continue
		}

		if removalTimestamp <= prh.getTimeHandler().Unix() {
			prh.removePIDFromCachers(pidBytes)
		}
	}
}

func (prh *peersRatingHandler) removePIDFromCachers(pid []byte) {
	prh.markedForRemovalCache.Remove(pid)
	prh.topRatedCache.Remove(pid)
	prh.badRatedCache.Remove(pid)
	delete(prh.ratingsMap, core.PeerID(pid).Pretty())
}

func (prh *peersRatingHandler) updateMetricsLoop(ctx context.Context) {
	timer := time.NewTimer(prh.timeBetweenMetricsUpdate)
	defer timer.Stop()

	for {
		timer.Reset(prh.timeBetweenMetricsUpdate)
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		prh.updateMetrics()
	}
}

func (prh *peersRatingHandler) updateRatingIfNeeded(pid core.PeerID, updateFactor int32) {
	oldRating, found := prh.getOldRating(pid.Bytes())
	if !found {
		// new pid, add it with default rating
		prh.topRatedCache.Put(pid.Bytes(), defaultRating, int32Size)
		prh.updateRatingsMap(pid, defaultRating, updateFactor)
	}

	decreasingUnderMin := oldRating == minRating && updateFactor == decreaseFactor
	increasingOverMax := oldRating == maxRating && updateFactor == increaseFactor
	shouldSkipUpdate := decreasingUnderMin || increasingOverMax
	if shouldSkipUpdate {
		prh.updateRatingsMap(pid, oldRating, updateFactor)
		return
	}

	newRating := oldRating + updateFactor
	if newRating > maxRating {
		newRating = maxRating
	}

	if newRating < minRating {
		newRating = minRating
	}

	prh.updateRating(pid, oldRating, newRating)
	prh.updateRatingsMap(pid, newRating, updateFactor)
}

func (prh *peersRatingHandler) updateRating(pid core.PeerID, oldRating, newRating int32) {
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

	prh.movePeerToNewTier(newRating, pid)
}

func computeRatingTier(peerRating int32) string {
	if peerRating >= defaultRating {
		return topRatedTier
	}

	return badRatedTier
}

func (prh *peersRatingHandler) movePeerToNewTier(newRating int32, pid core.PeerID) {
	newTier := computeRatingTier(newRating)
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

func (prh *peersRatingHandler) updateRatingsMap(pid core.PeerID, newRating int32, updateFactor int32) {
	prettyPID := pid.Pretty()
	peerRatingInfo, exists := prh.ratingsMap[prettyPID]
	if !exists {
		prh.ratingsMap[prettyPID] = &ratingInfo{
			Rating:                       newRating,
			TimestampLastRequestToPid:    0,
			TimestampLastResponseFromPid: 0,
		}
		return
	}

	peerRatingInfo.Rating = newRating

	newTimeStamp := prh.getTimeHandler().Unix()
	if updateFactor == decreaseFactor {
		peerRatingInfo.TimestampLastRequestToPid = newTimeStamp
		return
	}

	peerRatingInfo.TimestampLastResponseFromPid = newTimeStamp
}

func (prh *peersRatingHandler) updateMetrics() {
	prh.mut.RLock()
	defer prh.mut.RUnlock()

	jsonMap, err := json.Marshal(&prh.ratingsMap)
	if err != nil {
		log.Debug("could not update metrics", "metric", p2p.MetricP2PPeersRating, "error", err.Error())
		return
	}

	prh.appStatusHandler.SetStringValue(p2p.MetricP2PPeersRating, string(jsonMap))
}

// Close stops the go routines started by this instance
func (prh *peersRatingHandler) Close() error {
	prh.cancelSweepLoopFunc()
	prh.cancelMetricsUpdateLoopFunc()
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (prh *peersRatingHandler) IsInterfaceNil() bool {
	return prh == nil
}
