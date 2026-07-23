package libp2p_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/stretchr/testify/require"
)

const (
	mainNetwork         = p2p.NetworkType("main")
	transactionsNetwork = p2p.NetworkType("transactions")
)

func TestNewNetworkTopicsHolder(t *testing.T) {
	t.Parallel()

	holder := libp2p.NewNetworkTopicsHolder(&testscommon.LoggerStub{}, mainNetwork)
	require.NotNil(t, holder)
}

func TestNetworkTopicsHolder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	holder := libp2p.NewNetworkTopicsHolder(&testscommon.LoggerStub{}, mainNetwork)
	require.False(t, holder.IsInterfaceNil())
}

func TestNetworkTopicsHolder_AddTopicOnNetworkIfNeeded(t *testing.T) {
	t.Parallel()

	logger := &testscommon.LoggerStub{}
	holder := libp2p.NewNetworkTopicsHolder(logger, mainNetwork)

	holder.AddTopicOnNetworkIfNeeded(mainNetwork, "topic")
	nt := holder.GetNetworkTopics()
	require.Equal(t, mainNetwork, nt["topic"])

	// coverage only, should not override
	holder.AddTopicOnNetworkIfNeeded(mainNetwork, "topic")
}

func TestNetworkTopicsHolder_GetNetworkTypeForTopic(t *testing.T) {
	t.Parallel()

	logger := &testscommon.LoggerStub{
		DebugCalled: func(message string, args ...interface{}) {
			require.Contains(t, message, "p2p network not found")
		},
	}
	holder := libp2p.NewNetworkTopicsHolder(logger, mainNetwork)

	t.Run("non-existent topic should return main network", func(t *testing.T) {
		networkType := holder.GetNetworkTypeForTopic("missing_topic")
		require.Equal(t, mainNetwork, networkType)
	})

	t.Run("existent topic should return correct network", func(t *testing.T) {
		holder.AddTopicOnNetworkIfNeeded(mainNetwork, "topic")
		networkType := holder.GetNetworkTypeForTopic("topic")
		require.Equal(t, mainNetwork, networkType)
	})
}

func TestNetworkTopicsHolder_RemoveTopic(t *testing.T) {
	t.Parallel()

	logger := &testscommon.LoggerStub{}
	holder := libp2p.NewNetworkTopicsHolder(logger, mainNetwork)

	t.Run("should remove existing topic", func(t *testing.T) {
		holder.AddTopicOnNetworkIfNeeded(mainNetwork, "topic")
		holder.RemoveTopic("topic")
		nt := holder.GetNetworkTopics()
		require.Empty(t, nt)
	})

	t.Run("removing non-existent topic should not panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, "should have not panicked")
			}
		}()

		holder.RemoveTopic("missing_topic")
		nt := holder.GetNetworkTopics()
		require.Empty(t, nt)
	})
}

func TestNetworkTopicsHolder_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should have not panicked")
		}
	}()

	logger := &testscommon.LoggerStub{}
	holder := libp2p.NewNetworkTopicsHolder(logger, mainNetwork)
	wg := &sync.WaitGroup{}
	numGoroutines := 50
	wg.Add(numGoroutines)

	// Concurrent add, get, and remove operations
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			switch idx {
			case 0:
				holder.AddTopicOnNetworkIfNeeded(mainNetwork, fmt.Sprintf("topic%d", idx))
			case 1:
				_ = holder.GetNetworkTypeForTopic(fmt.Sprintf("topic%d", 3-idx))
			case 2:
				holder.RemoveTopic(fmt.Sprintf("topic%d", 3-idx))
			default:
				require.Fail(t, "should have not happened")
			}

			wg.Done()
		}(i % 3)
	}

	wg.Wait()
}
