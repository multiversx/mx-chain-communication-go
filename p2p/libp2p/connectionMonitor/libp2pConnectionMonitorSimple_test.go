package connectionMonitor_test

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p/connectionMonitor"
	"github.com/multiversx/mx-chain-communication-go/p2p/mock"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const durationTimeoutWaiting = time.Second * 3
const delayDisconnected = time.Millisecond * 100

func createMockArgsConnectionMonitorSimple() connectionMonitor.ArgsConnectionMonitorSimple {
	return connectionMonitor.ArgsConnectionMonitorSimple{
		Reconnecter:                &mock.ReconnecterStub{},
		ThresholdMinConnectedPeers: 3,
		Sharder:                    &mock.KadSharderStub{},
		PreferredPeersHolder:       &mock.PeersHolderStub{},
		ConnectionsWatcher:         &mock.ConnectionsWatcherStub{},
		Network:                    &mock.NetworkStub{},
	}
}

func TestNewLibp2pConnectionMonitorSimple(t *testing.T) {
	t.Parallel()

	t.Run("nil reconnecter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsConnectionMonitorSimple()
		args.Reconnecter = nil
		lcms, err := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)

		assert.Equal(t, p2p.ErrNilReconnecter, err)
		assert.True(t, check.IfNil(lcms))
	})
	t.Run("nil sharder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsConnectionMonitorSimple()
		args.Sharder = nil
		lcms, err := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)

		assert.Equal(t, p2p.ErrNilSharder, err)
		assert.True(t, check.IfNil(lcms))
	})
	t.Run("nil preferred peers holder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsConnectionMonitorSimple()
		args.PreferredPeersHolder = nil
		lcms, err := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)

		assert.Equal(t, p2p.ErrNilPreferredPeersHolder, err)
		assert.True(t, check.IfNil(lcms))
	})
	t.Run("nil connections watcher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsConnectionMonitorSimple()
		args.ConnectionsWatcher = nil
		lcms, err := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)

		assert.Equal(t, p2p.ErrNilConnectionsWatcher, err)
		assert.True(t, check.IfNil(lcms))
	})
	t.Run("nil network should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsConnectionMonitorSimple()
		args.Network = nil
		lcms, err := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)

		assert.Equal(t, p2p.ErrNilNetwork, err)
		assert.True(t, check.IfNil(lcms))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsConnectionMonitorSimple()
		lcms, err := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(lcms))
		assert.Nil(t, lcms.Close())
	})
}

func TestNewLibp2pConnectionMonitorSimple_OnDisconnectedUnderThresholdShouldCallReconnect(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Parallel()

	chReconnectCalled := make(chan struct{}, 1)

	rs := &mock.ReconnecterStub{
		ReconnectToNetworkCalled: func(ctx context.Context) {
			chReconnectCalled <- struct{}{}
		},
	}

	ns := mock.NetworkStub{
		PeersCall: func() []peer.ID {
			// only one connection which is under the threshold
			return []peer.ID{"mock"}
		},
	}

	args := createMockArgsConnectionMonitorSimple()
	args.Reconnecter = rs
	lcms, _ := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)

	go func() {
		time.Sleep(delayDisconnected)
		lcms.Disconnected(&ns, nil)
	}()
	select {
	case <-chReconnectCalled:
	case <-time.After(durationTimeoutWaiting):
		assert.Fail(t, "timeout waiting to call reconnect")
	}

	// second disconnection should not call reconnect due to flag
	go func() {
		time.Sleep(delayDisconnected)
		lcms.Disconnected(&ns, nil)
	}()
	select {
	case <-chReconnectCalled:
		assert.Fail(t, "should not have called reconnect")
	case <-time.After(durationTimeoutWaiting):
	}

	// flag should be reset after 5 seconds
	go func() {
		time.Sleep(time.Second * 5)
		lcms.Disconnected(&ns, nil)
	}()
	select {
	case <-chReconnectCalled:
	case <-time.After(durationTimeoutWaiting * 2):
		assert.Fail(t, "timeout waiting to call reconnect")
	}
}

func TestLibp2pConnectionMonitorSimple_ConnectedButDeniedShouldCloseConnection(t *testing.T) {
	t.Parallel()

	args := createMockArgsConnectionMonitorSimple()
	peerDenialEvaluator := &mock.PeerDenialEvaluatorStub{
		IsDeniedCalled: func(pid core.PeerID) bool {
			return true
		},
	}
	lcms, _ := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)
	_ = lcms.SetPeerDenialEvaluator(peerDenialEvaluator)

	wasCalled := false
	lcms.Connected(
		&mock.NetworkStub{},
		&mock.ConnStub{
			RemotePeerCalled: func() peer.ID {
				return "pid"
			},
			CloseCalled: func() error {
				wasCalled = true
				return nil
			},
		},
	)

	assert.True(t, wasCalled)
}

func TestLibp2pConnectionMonitorSimple_ConnectedWithSharderShouldCallEvictAndClosePeer(t *testing.T) {
	t.Parallel()

	evictedPid := []peer.ID{"evicted"}
	numComputeWasCalled := 0
	numClosedWasCalled := 0
	args := createMockArgsConnectionMonitorSimple()
	args.Sharder = &mock.KadSharderStub{
		ComputeEvictListCalled: func(pidList []peer.ID) []peer.ID {
			numComputeWasCalled++
			return evictedPid
		},
	}
	knownConnectionCalled := false
	args.ConnectionsWatcher = &mock.ConnectionsWatcherStub{
		NewKnownConnectionCalled: func(pid core.PeerID, connection string) {
			knownConnectionCalled = true
		},
	}
	putConnectionAddressCalled := false
	args.PreferredPeersHolder = &mock.PeersHolderStub{
		PutConnectionAddressCalled: func(peerID core.PeerID, addressSlice string) {
			putConnectionAddressCalled = true
		},
	}
	lcms, _ := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)

	lcms.Connected(
		&mock.NetworkStub{
			ClosePeerCall: func(id peer.ID) error {
				numClosedWasCalled++
				return nil
			},
			PeersCall: func() []peer.ID {
				return nil
			},
		},
		&mock.ConnStub{
			RemotePeerCalled: func() peer.ID {
				return evictedPid[0]
			},
		},
	)

	assert.Equal(t, 1, numClosedWasCalled)
	assert.Equal(t, 1, numComputeWasCalled)
	assert.True(t, knownConnectionCalled)
	assert.True(t, putConnectionAddressCalled)
}

func TestNewLibp2pConnectionMonitorSimple_DisconnectedShouldRemovePeerFromPreferredPeers(t *testing.T) {
	t.Parallel()

	prefPeerID := "preferred peer 0"
	chRemoveCalled := make(chan struct{}, 1)

	ns := mock.NetworkStub{
		PeersCall: func() []peer.ID {
			// only one connection which is under the threshold
			return []peer.ID{"mock"}
		},
	}

	removeCalled := false
	prefPeersHolder := &mock.PeersHolderStub{
		RemoveCalled: func(peerID core.PeerID) {
			removeCalled = true
			require.Equal(t, core.PeerID(prefPeerID), peerID)
			chRemoveCalled <- struct{}{}
		},
	}

	args := createMockArgsConnectionMonitorSimple()
	args.PreferredPeersHolder = prefPeersHolder
	lcms, _ := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)
	lcms.Disconnected(&ns, &mock.ConnStub{
		IDCalled: func() string {
			return prefPeerID
		},
	})

	require.True(t, removeCalled)
	select {
	case <-chRemoveCalled:
	case <-time.After(durationTimeoutWaiting):
		assert.Fail(t, "timeout waiting to call reconnect")
	}
}

func TestLibp2pConnectionMonitorSimple_EmptyFuncsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have panic")
		}
	}()

	netw := &mock.NetworkStub{
		PeersCall: func() []peer.ID {
			return make([]peer.ID, 0)
		},
	}

	args := createMockArgsConnectionMonitorSimple()
	lcms, _ := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)

	lcms.Disconnected(netw, nil)
	lcms.Listen(netw, nil)
	lcms.ListenClose(netw, nil)
}

func TestLibp2pConnectionMonitorSimple_SetThresholdMinConnectedPeers(t *testing.T) {
	t.Parallel()

	args := createMockArgsConnectionMonitorSimple()
	lcms, _ := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)

	thr := 10
	lcms.SetThresholdMinConnectedPeers(thr, &mock.NetworkStub{})
	thrSet := lcms.ThresholdMinConnectedPeers()

	assert.Equal(t, thr, thrSet)
}

func TestLibp2pConnectionMonitorSimple_SetThresholdMinConnectedPeersNilNetwShouldDoNothing(t *testing.T) {
	t.Parallel()

	minConnPeers := uint32(3)
	args := createMockArgsConnectionMonitorSimple()
	args.ThresholdMinConnectedPeers = minConnPeers
	lcms, _ := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)

	thr := 10
	lcms.SetThresholdMinConnectedPeers(thr, nil)
	thrSet := lcms.ThresholdMinConnectedPeers()

	assert.Equal(t, uint32(thrSet), minConnPeers)
}

func TestLibp2pConnectionMonitorSimple_SetPeerDenialEvaluator(t *testing.T) {
	t.Parallel()

	t.Run("nil handler should error", func(t *testing.T) {
		t.Parallel()

		lcms, _ := connectionMonitor.NewLibp2pConnectionMonitorSimple(createMockArgsConnectionMonitorSimple())

		err := lcms.SetPeerDenialEvaluator(nil)
		assert.Equal(t, p2p.ErrNilPeerDenialEvaluator, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		lcms, _ := connectionMonitor.NewLibp2pConnectionMonitorSimple(createMockArgsConnectionMonitorSimple())

		providedPeerDenialEvaluator := &mock.PeerDenialEvaluatorStub{}
		err := lcms.SetPeerDenialEvaluator(providedPeerDenialEvaluator)
		assert.Nil(t, err)
		assert.Equal(t, providedPeerDenialEvaluator, lcms.PeerDenialEvaluator())
	})
}

func TestNewLibp2pConnectionMonitorSimple_checkConnectionsBlocking(t *testing.T) {
	t.Parallel()

	args := createMockArgsConnectionMonitorSimple()
	ch := make(chan struct{}, 1)
	args.Network = &mock.NetworkStub{
		PeersCall: func() []peer.ID {
			return []peer.ID{"pid1"}
		},
		ClosePeerCall: func(id peer.ID) error {
			ch <- struct{}{}
			return nil
		},
	}
	peerDenialEvaluator := &mock.PeerDenialEvaluatorStub{
		IsDeniedCalled: func(pid core.PeerID) bool {
			return pid == "pid1"
		},
	}
	lcms, _ := connectionMonitor.NewLibp2pConnectionMonitorSimple(args)
	_ = lcms.SetPeerDenialEvaluator(peerDenialEvaluator)

	select {
	case <-ch:
	case <-time.After(durationTimeoutWaiting):
		assert.Fail(t, "timeout")
	}
}
