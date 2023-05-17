package rating

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/mock"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreMocks "github.com/multiversx/mx-chain-core-go/data/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArgPeersRatingMonitor() ArgPeersRatingMonitor {
	return ArgPeersRatingMonitor{
		TopRatedCache:       &mock.CacherStub{},
		BadRatedCache:       &mock.CacherStub{},
		ConnectionsProvider: &mock.ConnectionsProviderStub{},
	}
}

func TestNewPeersRatingMonitor(t *testing.T) {
	t.Parallel()

	t.Run("nil top rated cache should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeersRatingMonitor()
		args.TopRatedCache = nil

		monitor, err := NewPeersRatingMonitor(args)
		assert.True(t, errors.Is(err, p2p.ErrNilCacher))
		assert.True(t, strings.Contains(err.Error(), "TopRatedCache"))
		assert.True(t, check.IfNil(monitor))
	})
	t.Run("nil bad rated cache should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeersRatingMonitor()
		args.BadRatedCache = nil

		monitor, err := NewPeersRatingMonitor(args)
		assert.True(t, errors.Is(err, p2p.ErrNilCacher))
		assert.True(t, strings.Contains(err.Error(), "BadRatedCache"))
		assert.True(t, check.IfNil(monitor))
	})
	t.Run("nil connections provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeersRatingMonitor()
		args.ConnectionsProvider = nil

		monitor, err := NewPeersRatingMonitor(args)
		assert.Equal(t, p2p.ErrNilConnectionsProvider, err)
		assert.True(t, check.IfNil(monitor))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := ArgPeersRatingMonitor{
			TopRatedCache: coreMocks.NewCacherMock(),
			BadRatedCache: coreMocks.NewCacherMock(),
		}
		topPid1 := core.PeerID("top_pid1") // won't be returned as connection
		args.TopRatedCache.Put(topPid1.Bytes(), int32(100), 32)
		topPid2 := core.PeerID("top_pid2")
		args.TopRatedCache.Put(topPid2.Bytes(), int32(50), 32)
		topPid3 := core.PeerID("top_pid3")
		args.TopRatedCache.Put(topPid3.Bytes(), int32(0), 32)
		commonPid := core.PeerID("common_pid")
		args.TopRatedCache.Put(commonPid.Bytes(), int32(10), 32)

		badPid1 := core.PeerID("bad_pid1") // won't be returned as connection
		args.BadRatedCache.Put(badPid1.Bytes(), int32(-100), 32)
		badPid2 := core.PeerID("bad_pid2")
		args.BadRatedCache.Put(badPid2.Bytes(), int32(-50), 32)
		badPid3 := core.PeerID("bad_pid3")
		args.BadRatedCache.Put(badPid3.Bytes(), int32(-10), 32)
		args.BadRatedCache.Put(commonPid.Bytes(), int32(-10), 32) // should use the one from top-rated

		extraConnectedPid := core.PeerID("extra_connected_pid")
		args.ConnectionsProvider = &mock.ConnectionsProviderStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{topPid2, topPid3, badPid2, badPid3, commonPid, extraConnectedPid}
			},
		}

		monitor, err := NewPeersRatingMonitor(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(monitor))

		expectedMap := map[string]string{
			commonPid.Pretty():         "10",
			badPid2.Pretty():           "-50",
			badPid3.Pretty():           "-10",
			extraConnectedPid.Pretty(): unknownRating,
			topPid2.Pretty():           "50",
			topPid3.Pretty():           "0",
		}
		expectedStr, _ := json.Marshal(&expectedMap)
		ratings := monitor.GetConnectedPeersRatings()
		assert.Equal(t, string(expectedStr), ratings)
	})
}
