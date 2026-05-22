package libp2p_test

import (
	"testing"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/mock"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
)

func createMockArgDirectMsgThrottlerHandler() libp2p.ArgDirectMsgThrottlerHandler {
	return libp2p.ArgDirectMsgThrottlerHandler{
		MaxGoroutinesPerPeer: 10,
		Network:              &mock.NetworkStub{},
		Logger:               &testscommon.LoggerStub{},
	}
}

func TestNewDirectMsgThrottlerHandler(t *testing.T) {
	t.Parallel()

	t.Run("zero MaxGoroutinesPerPeer should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDirectMsgThrottlerHandler()
		args.MaxGoroutinesPerPeer = 0
		handler, err := libp2p.NewDirectMsgThrottlerHandler(args)
		assert.Equal(t, p2p.ErrInvalidValue, err)
		assert.Nil(t, handler)
	})
	t.Run("nil Network should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDirectMsgThrottlerHandler()
		args.Network = nil
		handler, err := libp2p.NewDirectMsgThrottlerHandler(args)
		assert.Equal(t, p2p.ErrNilNetwork, err)
		assert.Nil(t, handler)
	})
	t.Run("nil Logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDirectMsgThrottlerHandler()
		args.Logger = nil
		handler, err := libp2p.NewDirectMsgThrottlerHandler(args)
		assert.Equal(t, p2p.ErrNilLogger, err)
		assert.Nil(t, handler)
	})
	t.Run("should work and register as Notifiee", func(t *testing.T) {
		t.Parallel()

		notifyCalled := false
		args := createMockArgDirectMsgThrottlerHandler()
		args.Network = &mock.NetworkStub{
			NotifyCalled: func(_ network.Notifiee) {
				notifyCalled = true
			},
		}
		handler, err := libp2p.NewDirectMsgThrottlerHandler(args)
		require.Nil(t, err)
		require.NotNil(t, handler)
		assert.True(t, notifyCalled)
	})
}

func TestDirectMsgThrottlerHandler_CanProcess(t *testing.T) {
	t.Parallel()

	t.Run("under limit returns true", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDirectMsgThrottlerHandler()
		args.MaxGoroutinesPerPeer = 5
		handler, err := libp2p.NewDirectMsgThrottlerHandler(args)
		require.Nil(t, err)

		pid := core.PeerID("peer1")
		assert.True(t, handler.CanProcess(pid))
	})
	t.Run("at limit returns false", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDirectMsgThrottlerHandler()
		args.MaxGoroutinesPerPeer = 3
		handler, err := libp2p.NewDirectMsgThrottlerHandler(args)
		require.Nil(t, err)

		pid := core.PeerID("peer-limited")
		for i := 0; i < 3; i++ {
			assert.True(t, handler.CanProcess(pid))
			handler.StartProcessing(pid)
		}
		assert.False(t, handler.CanProcess(pid))
	})
}

func TestDirectMsgThrottlerHandler_StartEndProcessing(t *testing.T) {
	t.Parallel()

	t.Run("start and end processing correctly tracks count", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDirectMsgThrottlerHandler()
		args.MaxGoroutinesPerPeer = 2
		handler, err := libp2p.NewDirectMsgThrottlerHandler(args)
		require.Nil(t, err)

		pid := core.PeerID("tracked-peer")

		assert.True(t, handler.CanProcess(pid))
		handler.StartProcessing(pid)

		assert.True(t, handler.CanProcess(pid))
		handler.StartProcessing(pid)

		assert.False(t, handler.CanProcess(pid))

		handler.EndProcessing(pid)
		assert.True(t, handler.CanProcess(pid))

		handler.EndProcessing(pid)
		assert.True(t, handler.CanProcess(pid))
	})
	t.Run("multiple peers are tracked independently", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDirectMsgThrottlerHandler()
		args.MaxGoroutinesPerPeer = 1
		handler, err := libp2p.NewDirectMsgThrottlerHandler(args)
		require.Nil(t, err)

		pid1 := core.PeerID("peer-a")
		pid2 := core.PeerID("peer-b")

		assert.True(t, handler.CanProcess(pid1))
		handler.StartProcessing(pid1)
		assert.False(t, handler.CanProcess(pid1))

		assert.True(t, handler.CanProcess(pid2))
		handler.StartProcessing(pid2)
		assert.False(t, handler.CanProcess(pid2))

		handler.EndProcessing(pid1)
		assert.True(t, handler.CanProcess(pid1))
		assert.False(t, handler.CanProcess(pid2))
	})
}

func TestDirectMsgThrottlerHandler_Disconnected(t *testing.T) {
	t.Parallel()

	handler := libp2p.NewDirectMsgThrottlerHandlerForNetwork(&mock.NetworkStub{}, &testscommon.LoggerStub{})
	require.NotNil(t, handler)

	pid := core.PeerID("peer-disco")

	for i := int32(0); i < libp2p.MaxGoroutinesPerPeerVal; i++ {
		handler.StartProcessing(pid)
	}
	assert.False(t, handler.CanProcess(pid))

	conn := &mock.ConnStub{
		RemotePeerCalled: func() peer.ID {
			return peer.ID(pid)
		},
	}
	handler.Disconnected(nil, conn)

	assert.True(t, handler.CanProcess(pid))

	handler.Disconnected(nil, nil) // coverage only
}

func TestDirectMsgThrottlerHandler_NotifieeNoOps(t *testing.T) {
	t.Parallel()

	handler := libp2p.NewDirectMsgThrottlerHandlerForNetwork(&mock.NetworkStub{}, &testscommon.LoggerStub{})
	require.NotNil(t, handler)

	handler.Listen(nil, multiaddr.Multiaddr(nil))
	handler.ListenClose(nil, multiaddr.Multiaddr(nil))
	handler.Connected(nil, nil)
}

func TestDirectMsgThrottlerHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	handler := libp2p.NewDirectMsgThrottlerHandlerForNetwork(&mock.NetworkStub{}, &testscommon.LoggerStub{})
	assert.False(t, handler.IsInterfaceNil())
}
