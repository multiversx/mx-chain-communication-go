package rating

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/mock"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
)

func createMockArgPeersRatingMonitor() ArgPeersRatingMonitor {
	return ArgPeersRatingMonitor{
		TopRatedCache: &mock.CacherStub{},
		BadRatedCache: &mock.CacherStub{},
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
		assert.Nil(t, monitor)
	})
	t.Run("nil bad rated cache should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeersRatingMonitor()
		args.BadRatedCache = nil

		monitor, err := NewPeersRatingMonitor(args)
		assert.True(t, errors.Is(err, p2p.ErrNilCacher))
		assert.True(t, strings.Contains(err.Error(), "BadRatedCache"))
		assert.Nil(t, monitor)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := ArgPeersRatingMonitor{
			TopRatedCache: mock.NewCacherMock(),
			BadRatedCache: mock.NewCacherMock(),
		}

		monitor, err := NewPeersRatingMonitor(args)
		assert.Nil(t, err)
		assert.NotNil(t, monitor)
	})
}

func TestPeersRatingMonitor_GetConnectedPeersRatings(t *testing.T) {
	t.Parallel()

	t.Run("nil connections provider should error", func(t *testing.T) {
		t.Parallel()

		args := ArgPeersRatingMonitor{
			TopRatedCache: mock.NewCacherMock(),
			BadRatedCache: mock.NewCacherMock(),
		}
		monitor, err := NewPeersRatingMonitor(args)
		assert.Nil(t, err)

		ratings, err := monitor.GetConnectedPeersRatings(nil)
		assert.Equal(t, p2p.ErrNilConnectionsHandler, err)
		assert.Empty(t, ratings)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := ArgPeersRatingMonitor{
			TopRatedCache: mock.NewCacherMock(),
			BadRatedCache: mock.NewCacherMock(),
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
		connectionsProvider := &testscommon.ConnectionsHandlerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{topPid2, topPid3, badPid2, badPid3, commonPid, extraConnectedPid}
			},
		}

		monitor, err := NewPeersRatingMonitor(args)
		assert.Nil(t, err)

		expectedMap := map[string]string{
			commonPid.Pretty():         "10",
			badPid2.Pretty():           "-50",
			badPid3.Pretty():           "-10",
			extraConnectedPid.Pretty(): unknownRating,
			topPid2.Pretty():           "50",
			topPid3.Pretty():           "0",
		}
		expectedStr, _ := json.Marshal(&expectedMap)
		ratings, err := monitor.GetConnectedPeersRatings(connectionsProvider)
		assert.NoError(t, err)
		assert.Equal(t, string(expectedStr), ratings)
	})
}

func TestPeersRatingMonitor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var monitor *peersRatingMonitor
	assert.True(t, monitor.IsInterfaceNil())

	monitor, _ = NewPeersRatingMonitor(ArgPeersRatingMonitor{
		TopRatedCache: mock.NewCacherMock(),
		BadRatedCache: mock.NewCacherMock(),
	})
	assert.False(t, monitor.IsInterfaceNil())
}
