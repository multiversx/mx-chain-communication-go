package libp2p_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
)

func TestNewDirectMsgThrottlerHandler(t *testing.T) {
	t.Parallel()

	t.Run("non-positive limit should error", func(t *testing.T) {
		t.Parallel()

		handler, err := libp2p.NewDirectMsgThrottlerHandler(libp2p.ArgDirectMsgThrottlerHandler{})
		assert.Equal(t, p2p.ErrInvalidValue, err)
		assert.Nil(t, handler)
	})
	t.Run("positive limit should work", func(t *testing.T) {
		t.Parallel()

		handler, err := libp2p.NewDirectMsgThrottlerHandler(libp2p.ArgDirectMsgThrottlerHandler{
			MaxGoroutinesPerPeer: 1,
		})
		require.NoError(t, err)
		assert.False(t, handler.IsInterfaceNil())
	})
}

func TestDirectMsgThrottlerHandler_Processing(t *testing.T) {
	t.Parallel()

	handler, err := libp2p.NewDirectMsgThrottlerHandler(libp2p.ArgDirectMsgThrottlerHandler{
		MaxGoroutinesPerPeer: 2,
	})
	require.NoError(t, err)

	pid1 := core.PeerID("peer-a")
	pid2 := core.PeerID("peer-b")

	assert.True(t, handler.TryStartProcessing(pid1))
	assert.True(t, handler.TryStartProcessing(pid1))
	assert.False(t, handler.TryStartProcessing(pid1))
	assert.True(t, handler.TryStartProcessing(pid2))

	handler.EndProcessing(pid1)
	assert.True(t, handler.TryStartProcessing(pid1))

	handler.EndProcessing(pid1)
	handler.EndProcessing(pid1)
	handler.EndProcessing(pid2)

	assert.True(t, handler.TryStartProcessing(pid1))
	assert.True(t, handler.TryStartProcessing(pid2))
}

func TestDirectMsgThrottlerHandler_ConcurrentAdmissionIsBounded(t *testing.T) {
	t.Parallel()

	const limit int32 = 10
	const numCallers = 1000

	handler, err := libp2p.NewDirectMsgThrottlerHandler(libp2p.ArgDirectMsgThrottlerHandler{
		MaxGoroutinesPerPeer: limit,
	})
	require.NoError(t, err)

	pid := core.PeerID("peer")
	start := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(numCallers)
	var accepted atomic.Int32

	for idx := 0; idx < numCallers; idx++ {
		go func() {
			defer wg.Done()
			<-start
			if handler.TryStartProcessing(pid) {
				accepted.Add(1)
			}
		}()
	}

	close(start)
	wg.Wait()

	assert.Equal(t, limit, accepted.Load())
}

func TestDirectMsgThrottlerHandler_EndWithoutAdmissionDoesNotAffectLimit(t *testing.T) {
	t.Parallel()

	handler, err := libp2p.NewDirectMsgThrottlerHandler(libp2p.ArgDirectMsgThrottlerHandler{
		MaxGoroutinesPerPeer: 1,
	})
	require.NoError(t, err)

	pid := core.PeerID("peer")
	handler.EndProcessing(pid)
	assert.True(t, handler.TryStartProcessing(pid))
	assert.False(t, handler.TryStartProcessing(pid))
}
