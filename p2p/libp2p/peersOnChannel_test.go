package libp2p_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/mock"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/multiversx/mx-chain-core-go/core"
	coreAtomic "github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/stretchr/testify/assert"
)

func getPubSubs() map[p2p.NetworkType]libp2p.PubSub {
	return map[p2p.NetworkType]libp2p.PubSub{
		"main": &mock.PubSubStub{},
	}
}

func TestNewPeersOnChannel_EmptyPubSubsShouldErr(t *testing.T) {
	t.Parallel()

	poc, err := libp2p.NewPeersOnChannel(nil, &testscommon.NetworkTopicsHolderMock{}, 1, 1, &testscommon.LoggerStub{})

	assert.Nil(t, poc)
	assert.Equal(t, p2p.ErrNoPubSub, err)
}

func TestNewPeersOnChannel_NilNetworkTopicsHolderShouldErr(t *testing.T) {
	t.Parallel()

	poc, err := libp2p.NewPeersOnChannel(getPubSubs(), nil, 1, 1, &testscommon.LoggerStub{})

	assert.Nil(t, poc)
	assert.Equal(t, p2p.ErrNilNetworkTopicsHolder, err)
}

func TestNewPeersOnChannel_InvalidRefreshIntervalShouldErr(t *testing.T) {
	t.Parallel()

	poc, err := libp2p.NewPeersOnChannel(
		getPubSubs(),
		&testscommon.NetworkTopicsHolderMock{},
		0,
		1,
		&testscommon.LoggerStub{})

	assert.Nil(t, poc)
	assert.Equal(t, p2p.ErrInvalidDurationProvided, err)
}

func TestNewPeersOnChannel_InvalidTTLIntervalShouldErr(t *testing.T) {
	t.Parallel()

	poc, err := libp2p.NewPeersOnChannel(
		getPubSubs(),
		&testscommon.NetworkTopicsHolderMock{},
		1,
		0,
		&testscommon.LoggerStub{})

	assert.Nil(t, poc)
	assert.Equal(t, p2p.ErrInvalidDurationProvided, err)
}

func TestNewPeersOnChannel_NilLoggerShouldErr(t *testing.T) {
	t.Parallel()

	poc, err := libp2p.NewPeersOnChannel(
		getPubSubs(),
		&testscommon.NetworkTopicsHolderMock{},
		1,
		1,
		nil)

	assert.Nil(t, poc)
	assert.Equal(t, p2p.ErrNilLogger, err)
}

func TestNewPeersOnChannel_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	poc, err := libp2p.NewPeersOnChannel(
		getPubSubs(),
		&testscommon.NetworkTopicsHolderMock{},
		1,
		1,
		&testscommon.LoggerStub{})

	assert.NotNil(t, poc)
	assert.Nil(t, err)
}

func TestPeersOnChannel_ConnectedPeersOnChannelMissingTopicShouldTriggerFetchAndReturn(t *testing.T) {
	t.Parallel()

	retPeerIDs := []peer.ID{"peer1", "peer2"}
	wasFetchCalled := atomic.Value{}
	wasFetchCalled.Store(false)

	poc, _ := libp2p.NewPeersOnChannel(
		map[p2p.NetworkType]libp2p.PubSub{
			"main": &mock.PubSubStub{
				ListPeersCalled: func(topic string) []peer.ID {
					if topic == testTopic {
						wasFetchCalled.Store(true)
						return retPeerIDs
					}
					return nil
				},
			},
		},
		&testscommon.NetworkTopicsHolderMock{},
		time.Second,
		time.Second,
		&testscommon.LoggerStub{},
	)

	peers := poc.ConnectedPeersOnChannel(testTopic)

	assert.True(t, wasFetchCalled.Load().(bool))
	for idx, pid := range retPeerIDs {
		assert.Equal(t, []byte(pid), peers[idx].Bytes())
	}
}

func TestPeersOnChannel_ConnectedPeersOnChannelFindTopicShouldReturn(t *testing.T) {
	t.Parallel()

	retPeerIDs := []core.PeerID{"peer1", "peer2"}
	wasFetchCalled := atomic.Value{}
	wasFetchCalled.Store(false)

	poc, _ := libp2p.NewPeersOnChannel(
		map[p2p.NetworkType]libp2p.PubSub{
			"main": &mock.PubSubStub{
				ListPeersCalled: func(topic string) []peer.ID {
					wasFetchCalled.Store(true)
					return nil
				},
			},
		},
		&testscommon.NetworkTopicsHolderMock{
			GetNetworkTypeForTopicCalled: func(topic string) p2p.NetworkType {
				return ""
			},
		},
		time.Second,
		time.Second,
		&testscommon.LoggerStub{},
	)
	// manually put peers
	poc.SetPeersOnTopic(testTopic, time.Now(), retPeerIDs)

	peers := poc.ConnectedPeersOnChannel(testTopic)

	assert.False(t, wasFetchCalled.Load().(bool))
	for idx, pid := range retPeerIDs {
		assert.Equal(t, []byte(pid), peers[idx].Bytes())
	}
}

func TestPeersOnChannel_RefreshShouldBeDone(t *testing.T) {
	t.Parallel()

	retPeerIDs := []core.PeerID{"peer1", "peer2"}
	wasFetchCalled := coreAtomic.Flag{}
	wasFetchCalled.Reset()

	refreshInterval := time.Millisecond * 100
	ttlInterval := time.Duration(2)

	poc, _ := libp2p.NewPeersOnChannel(
		map[p2p.NetworkType]libp2p.PubSub{
			"main": &mock.PubSubStub{
				ListPeersCalled: func(topic string) []peer.ID {
					wasFetchCalled.SetValue(true)
					return nil
				},
			},
		},
		&testscommon.NetworkTopicsHolderMock{},
		refreshInterval,
		ttlInterval,
		&testscommon.LoggerStub{},
	)
	poc.SetTimeHandler(func() time.Time {
		return time.Unix(0, 4)
	})
	// manually put peers
	poc.SetPeersOnTopic(testTopic, time.Unix(0, 1), retPeerIDs)

	// wait for the go routine cycle finish up
	time.Sleep(time.Second)

	assert.True(t, wasFetchCalled.IsSet())
	assert.Empty(t, poc.GetPeers(testTopic))
}
