package libp2p_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/config"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/mock"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/stretchr/testify/require"
)

func createTestArgs() libp2p.ArgsPubSubsHolder {
	return libp2p.ArgsPubSubsHolder{
		Log: &testscommon.LoggerStub{},
		Host: &mock.ConnectableHostStub{
			PeerstoreCalled: func() peerstore.Peerstore {
				return &mock.PeerstoreStub{
					PeersCalled: func() peer.IDSlice {
						return peer.IDSlice{}
					},
				}
			},
		},
		MainNetwork:         mainNetwork,
		MessageSigning:      false,
		NetworkTopicsHolder: &testscommon.NetworkTopicsHolderMock{},
		P2pConfig: config.P2PConfig{
			SubNetworks: config.SubNetworksConfig{
				Networks: []config.SubNetworkConfig{
					{
						Name:        string(transactionsNetwork),
						ProtocolIDs: []string{"mvx-transactions"},
						PubSub: config.PubSubConfig{
							OptimalPeersNum: 2,
							MaximumPeersNum: 3,
							MinimumPeersNum: 1,
						},
					},
				},
			},
		},
	}
}

func TestNewPubSubsHolder(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should return error", func(t *testing.T) {
		t.Parallel()

		args := createTestArgs()
		args.Log = nil
		holder, err := libp2p.NewPubSubsHolder(args)
		require.Equal(t, p2p.ErrNilLogger, err)
		require.Nil(t, holder)
	})
	t.Run("nil host should return error", func(t *testing.T) {
		t.Parallel()

		args := createTestArgs()
		args.Host = nil
		holder, err := libp2p.NewPubSubsHolder(args)
		require.Equal(t, p2p.ErrNilHost, err)
		require.Nil(t, holder)
	})
	t.Run("nil network topics holder should return error", func(t *testing.T) {
		t.Parallel()

		args := createTestArgs()
		args.NetworkTopicsHolder = nil
		holder, err := libp2p.NewPubSubsHolder(args)
		require.Equal(t, p2p.ErrNilNetworkTopicsHolder, err)
		require.Nil(t, holder)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createTestArgs()
		holder, err := libp2p.NewPubSubsHolder(args)
		require.NoError(t, err)
		require.NotNil(t, holder)
	})
}

func TestPubSubsHolder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	args := createTestArgs()
	args.Log = nil
	holder, _ := libp2p.NewPubSubsHolder(args)
	require.True(t, holder.IsInterfaceNil())

	holder, _ = libp2p.NewPubSubsHolder(createTestArgs())
	require.False(t, holder.IsInterfaceNil())
}

func TestPubSubsHolder_GetPubSub(t *testing.T) {
	t.Parallel()

	t.Run("should return pubsub for existing topic", func(t *testing.T) {
		t.Parallel()

		args := createTestArgs()
		args.NetworkTopicsHolder = &testscommon.NetworkTopicsHolderMock{
			GetNetworkTypeForTopicCalled: func(topic string) p2p.NetworkType {
				require.Equal(t, providedTopic, topic)
				return mainNetwork
			},
		}
		holder, err := libp2p.NewPubSubsHolder(args)
		require.NoError(t, err)

		pubSub, found := holder.GetPubSub(providedTopic)
		require.True(t, found)
		require.NotNil(t, pubSub)
	})
	t.Run("should return false for non-existing topic", func(t *testing.T) {
		t.Parallel()

		args := createTestArgs()
		args.NetworkTopicsHolder = &testscommon.NetworkTopicsHolderMock{
			GetNetworkTypeForTopicCalled: func(topic string) p2p.NetworkType {
				require.Equal(t, providedTopic, topic)
				return "invalidNetwork"
			},
		}
		holder, err := libp2p.NewPubSubsHolder(args)
		require.NoError(t, err)
		pubSub, found := holder.GetPubSub(providedTopic)
		require.False(t, found)
		require.Nil(t, pubSub)
	})
}

func TestPubSubsHolder_Close(t *testing.T) {
	t.Parallel()

	holder, err := libp2p.NewPubSubsHolder(createTestArgs())
	require.NoError(t, err)

	err = holder.Close()
	require.NoError(t, err)
}

func TestPubSubsHolder_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should have not panicked")
		}
	}()

	args := createTestArgs()
	cnt := uint32(0)
	args.NetworkTopicsHolder = &testscommon.NetworkTopicsHolderMock{
		GetNetworkTypeForTopicCalled: func(topic string) p2p.NetworkType {
			atomic.AddUint32(&cnt, 1)
			if atomic.LoadUint32(&cnt)%2 == 0 {
				return transactionsNetwork
			}

			return mainNetwork
		},
	}
	holder, err := libp2p.NewPubSubsHolder(args)
	require.NoError(t, err)

	numGoroutines := 100
	wg := sync.WaitGroup{}
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_, _ = holder.GetPubSub(providedTopic)
		}()
	}

	wg.Wait()
}
