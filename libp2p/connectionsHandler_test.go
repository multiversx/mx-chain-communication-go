package libp2p_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	p2p "github.com/ElrondNetwork/elrond-go-p2p"
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-p2p/mock"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")

func createMockArgConnectionsHandler() libp2p.ArgConnectionsHandler {
	return libp2p.ArgConnectionsHandler{
		P2pHost:              &mock.ConnectableHostStub{},
		PeersOnChannel:       &mock.PeersOnChannelStub{},
		PeerShardResolver:    &mock.PeerShardResolverStub{},
		Sharder:              &mock.SharderStub{},
		PreferredPeersHolder: &mock.PeersHolderStub{},
		ConnMonitor:          &mock.ConnectionMonitorStub{},
	}
}

func TestNewConnectionsHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil P2pHost should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgConnectionsHandler()
		args.P2pHost = nil
		ch, err := libp2p.NewConnectionsHandler(args)
		assert.Equal(t, p2p.ErrNilHost, err)
		assert.True(t, check.IfNil(ch))
	})
	t.Run("nil PeersOnChannel should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgConnectionsHandler()
		args.PeersOnChannel = nil
		ch, err := libp2p.NewConnectionsHandler(args)
		assert.Equal(t, p2p.ErrNilPeersOnChannel, err)
		assert.True(t, check.IfNil(ch))
	})
	t.Run("nil PeerShardResolver should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgConnectionsHandler()
		args.PeerShardResolver = nil
		ch, err := libp2p.NewConnectionsHandler(args)
		assert.Equal(t, p2p.ErrNilPeerShardResolver, err)
		assert.True(t, check.IfNil(ch))
	})
	t.Run("nil Sharder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgConnectionsHandler()
		args.Sharder = nil
		ch, err := libp2p.NewConnectionsHandler(args)
		assert.Equal(t, p2p.ErrNilSharder, err)
		assert.True(t, check.IfNil(ch))
	})
	t.Run("nil PreferredPeersHolder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgConnectionsHandler()
		args.PreferredPeersHolder = nil
		ch, err := libp2p.NewConnectionsHandler(args)
		assert.Equal(t, p2p.ErrNilPreferredPeersHolder, err)
		assert.True(t, check.IfNil(ch))
	})
	t.Run("nil ConnMonitor should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgConnectionsHandler()
		args.ConnMonitor = nil
		ch, err := libp2p.NewConnectionsHandler(args)
		assert.Equal(t, p2p.ErrNilConnectionMonitor, err)
		assert.True(t, check.IfNil(ch))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ch, err := libp2p.NewConnectionsHandler(createMockArgConnectionsHandler())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(ch))
	})
}

func TestConnectionsHandler_Peers(t *testing.T) {
	t.Parallel()

	providedPeers := []peer.ID{"peer1", "peer2"}
	expectedPIDs := []core.PeerID{"peer1", "peer2"}
	args := createMockArgConnectionsHandler()
	args.P2pHost = &mock.ConnectableHostStub{
		PeerstoreCalled: func() peerstore.Peerstore {
			return &mock.PeerstoreStub{
				PeersCalled: func() peer.IDSlice {
					return providedPeers
				},
			}
		},
	}
	ch, _ := libp2p.NewConnectionsHandler(args)
	assert.False(t, check.IfNil(ch))
	peers := ch.Peers()
	assert.Equal(t, expectedPIDs, peers)
}

func TestConnectionsHandler_Addresses(t *testing.T) {
	t.Parallel()

	providedAddr1, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9999/p2p/16Uiu2HAkw5SNNtSvH1zJiQ6Gc3WoGNSxiyNueRKe6fuAuh57G3Bk")
	providedAddr2, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9999/p2p/16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ")
	providedAddrs := []multiaddr.Multiaddr{providedAddr1, providedAddr2}
	args := createMockArgConnectionsHandler()
	args.P2pHost = &mock.ConnectableHostStub{
		AddrsCalled: func() []multiaddr.Multiaddr {
			return providedAddrs
		},
	}
	ch, _ := libp2p.NewConnectionsHandler(args)
	assert.False(t, check.IfNil(ch))
	addrs := ch.Addresses()
	assert.Equal(t, 2, len(addrs))
	assert.True(t, strings.Contains(addrs[0], providedAddr1.String()))
	assert.True(t, strings.Contains(addrs[1], providedAddr2.String()))
}

func TestConnectionsHandler_ConnectToPeer(t *testing.T) {
	t.Parallel()

	args := createMockArgConnectionsHandler()
	wasCalled := false
	args.P2pHost = &mock.ConnectableHostStub{
		ConnectToPeerCalled: func(ctx context.Context, address string) error {
			wasCalled = true
			return nil
		},
	}
	ch, _ := libp2p.NewConnectionsHandler(args)
	assert.False(t, check.IfNil(ch))
	err := ch.ConnectToPeer("address")
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestConnectionsHandler_WaitForConnections(t *testing.T) {
	t.Parallel()

	t.Run("zero min num of peers should early exit", func(t *testing.T) {
		t.Parallel()

		ch, _ := libp2p.NewConnectionsHandler(createMockArgConnectionsHandler())
		assert.False(t, check.IfNil(ch))
		maxWaitingTime := time.Millisecond * 100
		start := time.Now()
		ch.WaitForConnections(maxWaitingTime, 0)
		stop := time.Now()
		assert.True(t, stop.Sub(start) <= time.Millisecond*150)
	})
	t.Run("should work but not reach the target", func(t *testing.T) {
		t.Parallel()

		ch, _ := libp2p.NewConnectionsHandler(createMockArgConnectionsHandler())
		assert.False(t, check.IfNil(ch))
		maxWaitingTime := time.Millisecond * 100
		start := time.Now()
		ch.WaitForConnections(maxWaitingTime, 1)
		stop := time.Now()
		assert.True(t, stop.Sub(start) <= time.Millisecond*150)
	})
	t.Run("should work and reach the target", func(t *testing.T) {
		t.Parallel()

		providedConnectedPID := peer.ID("connectedPID1")
		args := createMockArgConnectionsHandler()
		args.P2pHost = &mock.ConnectableHostStub{
			NetworkCalled: func() network.Network {
				return &mock.NetworkStub{
					ConnectednessCalled: func(pid peer.ID) network.Connectedness {
						return network.Connected
					},
					ConnsCalled: func() []network.Conn {
						return []network.Conn{
							&mock.ConnStub{
								RemotePeerCalled: func() peer.ID {
									return providedConnectedPID
								},
							},
						}
					},
				}
			},
		}
		ch, _ := libp2p.NewConnectionsHandler(args)
		assert.False(t, check.IfNil(ch))
		maxWaitingTime := time.Second * 3
		start := time.Now()
		ch.WaitForConnections(maxWaitingTime, 1)
		stop := time.Now()
		assert.True(t, stop.Sub(start) < time.Second*2) // should finish way before maxWaitingTime
	})
}

func TestConnectionsHandler_IsConnected(t *testing.T) {
	t.Parallel()

	args := createMockArgConnectionsHandler()
	providedConnectedPID := peer.ID("connectedPID")
	providedNotConnectedPID := peer.ID("notConnectedPID")
	args.P2pHost = &mock.ConnectableHostStub{
		NetworkCalled: func() network.Network {
			return &mock.NetworkStub{
				ConnectednessCalled: func(pid peer.ID) network.Connectedness {
					if pid == providedConnectedPID {
						return network.Connected
					}
					return network.NotConnected
				},
			}
		},
	}
	ch, _ := libp2p.NewConnectionsHandler(args)
	assert.False(t, check.IfNil(ch))
	assert.True(t, ch.IsConnected(core.PeerID(providedConnectedPID)))
	assert.False(t, ch.IsConnected(core.PeerID(providedNotConnectedPID)))
}

func TestConnectionsHandler_ConnectedPeers(t *testing.T) {
	t.Parallel()

	args := createMockArgConnectionsHandler()
	providedConnectedPID1 := peer.ID("connectedPID1")
	providedConnectedPID2 := peer.ID("connectedPID2")
	providedNotConnectedPID := peer.ID("notConnectedPID")
	args.P2pHost = &mock.ConnectableHostStub{
		NetworkCalled: func() network.Network {
			return &mock.NetworkStub{
				ConnectednessCalled: func(pid peer.ID) network.Connectedness {
					if pid == providedConnectedPID1 || pid == providedConnectedPID2 {
						return network.Connected
					}
					return network.NotConnected
				},
				ConnsCalled: func() []network.Conn {
					return []network.Conn{
						&mock.ConnStub{
							RemotePeerCalled: func() peer.ID {
								return providedConnectedPID1
							},
						},
						&mock.ConnStub{
							RemotePeerCalled: func() peer.ID {
								return providedConnectedPID2
							},
						},
						&mock.ConnStub{
							RemotePeerCalled: func() peer.ID {
								return providedNotConnectedPID
							},
						},
					}
				},
			}
		},
	}
	ch, _ := libp2p.NewConnectionsHandler(args)
	assert.False(t, check.IfNil(ch))
	expectedPeers := []core.PeerID{core.PeerID(providedConnectedPID1), core.PeerID(providedConnectedPID2)}
	connectedPeers := ch.ConnectedPeers()
	assert.Equal(t, expectedPeers, connectedPeers)
}

func TestConnectionsHandler_ConnectedAddresses(t *testing.T) {
	t.Parallel()

	args := createMockArgConnectionsHandler()
	providedPID1 := peer.ID("connectedPID1")
	providedPID2 := peer.ID("connectedPID2")
	providedAddr := "/ip4/127.0.0.1/tcp/9999/p2p/16Uiu2HAkw5SNNtSvH1zJiQ6Gc3WoGNSxiyNueRKe6fuAuh57G3Bk"
	args.P2pHost = &mock.ConnectableHostStub{
		NetworkCalled: func() network.Network {
			return &mock.NetworkStub{
				ConnsCalled: func() []network.Conn {
					return []network.Conn{
						&mock.ConnStub{
							RemoteMultiaddrCalled: func() multiaddr.Multiaddr {
								return multiaddr.StringCast(providedAddr)
							},
							RemotePeerCalled: func() peer.ID {
								return providedPID1
							},
						},
						&mock.ConnStub{
							RemoteMultiaddrCalled: func() multiaddr.Multiaddr {
								return multiaddr.StringCast(providedAddr)
							},
							RemotePeerCalled: func() peer.ID {
								return providedPID2
							},
						},
					}
				},
			}
		},
	}
	ch, _ := libp2p.NewConnectionsHandler(args)
	assert.False(t, check.IfNil(ch))
	expectedAddrs := []string{providedAddr + "/p2p/" + providedPID1.String(), providedAddr + "/p2p/" + providedPID2.String()}
	connectedAddrs := ch.ConnectedAddresses()
	assert.Equal(t, expectedAddrs, connectedAddrs)
}

func TestConnectionsHandler_PeerAddresses(t *testing.T) {
	t.Parallel()

	providedPID := peer.ID("provided pid")
	providedAddr := "/ip4/127.0.0.1/tcp/9999/p2p/16Uiu2HAkw5SNNtSvH1zJiQ6Gc3WoGNSxiyNueRKe6fuAuh57G3Bk"
	providedPSAddr := "/ip4/127.0.0.1/tcp/10000/p2p/16Uiu2HAkw5SNNtSvH1zJiQ6Gc3WoGNSxiyNueRKe6fuAuh57G3Bk"
	otherPID := peer.ID("other pid")
	args := createMockArgConnectionsHandler()
	args.P2pHost = &mock.ConnectableHostStub{
		NetworkCalled: func() network.Network {
			return &mock.NetworkStub{
				ConnsCalled: func() []network.Conn {
					return []network.Conn{
						&mock.ConnStub{
							RemotePeerCalled: func() peer.ID {
								return otherPID
							},
						},
						&mock.ConnStub{
							RemoteMultiaddrCalled: func() multiaddr.Multiaddr {
								return multiaddr.StringCast(providedAddr)
							},
							RemotePeerCalled: func() peer.ID {
								return providedPID
							},
						},
					}
				},
			}
		},
		PeerstoreCalled: func() peerstore.Peerstore {
			return &mock.PeerstoreStub{
				AddrsCalled: func(pid peer.ID) []multiaddr.Multiaddr {
					assert.Equal(t, providedPID, pid)
					return []multiaddr.Multiaddr{multiaddr.StringCast(providedPSAddr)}
				},
			}
		},
	}
	ch, _ := libp2p.NewConnectionsHandler(args)
	assert.False(t, check.IfNil(ch))
	expectedAddrs := []string{providedAddr, providedPSAddr}
	addrs := ch.PeerAddresses(core.PeerID(providedPID))
	assert.Equal(t, expectedAddrs, addrs)
}

func TestConnectionsHandler_ConnectedPeersOnTopic(t *testing.T) {
	t.Parallel()

	args := createMockArgConnectionsHandler()
	wasCalled := false
	args.PeersOnChannel = &mock.PeersOnChannelStub{
		ConnectedPeersOnChannelCalled: func(topic string) []core.PeerID {
			wasCalled = true
			return nil
		},
	}
	ch, _ := libp2p.NewConnectionsHandler(args)
	assert.False(t, check.IfNil(ch))
	err := ch.ConnectedPeersOnTopic("topic")
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestConnectionsHandler_ConnectedFullHistoryPeersOnTopic(t *testing.T) {
	t.Parallel()

	providedConnectedPIDs := []core.PeerID{core.PeerID("connected pid1"), core.PeerID("connected pid2")}
	args := createMockArgConnectionsHandler()
	args.PeersOnChannel = &mock.PeersOnChannelStub{
		ConnectedPeersOnChannelCalled: func(topic string) []core.PeerID {
			return providedConnectedPIDs
		},
	}
	args.PeerShardResolver = &mock.PeerShardResolverStub{
		GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
			return core.P2PPeerInfo{
				PeerSubType: core.FullHistoryObserver,
			}
		},
	}
	ch, _ := libp2p.NewConnectionsHandler(args)
	assert.False(t, check.IfNil(ch))
	connectedFH := ch.ConnectedFullHistoryPeersOnTopic("topic")
	assert.Equal(t, providedConnectedPIDs, connectedFH)
}

func TestConnectionsHandler_SetPeerShardResolver(t *testing.T) {
	t.Parallel()

	t.Run("nil resolver should error", func(t *testing.T) {
		t.Parallel()

		ch, _ := libp2p.NewConnectionsHandler(createMockArgConnectionsHandler())
		assert.False(t, check.IfNil(ch))
		err := ch.SetPeerShardResolver(nil)
		assert.Equal(t, p2p.ErrNilPeerShardResolver, err)
	})
	t.Run("set resolver on sharder returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgConnectionsHandler()
		args.Sharder = &mock.SharderStub{
			SetPeerShardResolverCalled: func(psp p2p.PeerShardResolver) error {
				return expectedErr
			},
		}
		ch, _ := libp2p.NewConnectionsHandler(args)
		assert.False(t, check.IfNil(ch))
		err := ch.SetPeerShardResolver(&mock.PeerShardResolverStub{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedResolver := &mock.PeerShardResolverStub{}
		args := createMockArgConnectionsHandler()
		args.Sharder = &mock.SharderStub{
			SetPeerShardResolverCalled: func(psp p2p.PeerShardResolver) error {
				assert.Equal(t, providedResolver, psp)
				return nil
			},
		}
		ch, _ := libp2p.NewConnectionsHandler(args)
		assert.False(t, check.IfNil(ch))
		err := ch.SetPeerShardResolver(providedResolver)
		assert.Nil(t, err)
	})
}

func TestConnectionsHandler_GetConnectedPeersInfo(t *testing.T) {
	t.Parallel()

	providedSelf := peer.ID("self")
	providedSeeder := peer.ID("seeder")
	providedUnknown := peer.ID("unknown")
	providedValidatorIntra := peer.ID("validator intra")
	providedValidatorCross := peer.ID("validator cross")
	providedObserverIntra := peer.ID("observer intra")
	providedObserverCross := peer.ID("observer cross")
	providedFullHistory := peer.ID("full history")
	providedConnectedPIDs := []peer.ID{providedSeeder, providedUnknown, providedValidatorIntra, providedValidatorCross, providedObserverIntra, providedObserverCross, providedFullHistory}
	args := createMockArgConnectionsHandler()
	args.P2pHost = &mock.ConnectableHostStub{
		NetworkCalled: func() network.Network {
			return &mock.NetworkStub{
				PeersCall: func() []peer.ID {
					return providedConnectedPIDs
				},
				ConnsToPeerCalled: func(pid peer.ID) []network.Conn {
					return []network.Conn{
						&mock.ConnStub{
							RemoteMultiaddrCalled: func() multiaddr.Multiaddr {
								return &mock.MultiaddrStub{
									StringCalled: func() string {
										return "multiaddr"
									},
								}
							},
						},
					}
				},
			}
		},
		IDCalled: func() peer.ID {
			return providedSelf
		},
	}
	args.PeerShardResolver = &mock.PeerShardResolverStub{
		GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
			switch peer.ID(pid) {
			case providedSelf:
				return core.P2PPeerInfo{
					ShardID: 0,
				}
			case providedSeeder:
				return core.P2PPeerInfo{}
			case providedUnknown:
				return core.P2PPeerInfo{
					PeerType: core.UnknownPeer,
				}
			case providedValidatorIntra:
				return core.P2PPeerInfo{
					ShardID:  0,
					PeerType: core.ValidatorPeer,
				}
			case providedValidatorCross:
				return core.P2PPeerInfo{
					ShardID:  1,
					PeerType: core.ValidatorPeer,
				}
			case providedObserverIntra:
				return core.P2PPeerInfo{
					ShardID:  0,
					PeerType: core.ObserverPeer,
				}
			case providedObserverCross:
				return core.P2PPeerInfo{
					ShardID:  1,
					PeerType: core.ObserverPeer,
				}
			case providedFullHistory:
				return core.P2PPeerInfo{
					PeerType:    core.ObserverPeer,
					PeerSubType: core.FullHistoryObserver,
				}
			default:
				assert.Fail(t, "should never be called")
				return core.P2PPeerInfo{}
			}
		},
	}
	args.Sharder = &mock.SharderStub{
		IsSeederCalled: func(pid core.PeerID) bool {
			return pid == core.PeerID(providedSeeder)
		},
	}
	args.PreferredPeersHolder = &mock.PeersHolderStub{
		ContainsCalled: func(peerID core.PeerID) bool {
			return peerID == core.PeerID(providedValidatorIntra)
		},
	}
	ch, _ := libp2p.NewConnectionsHandler(args)
	assert.False(t, check.IfNil(ch))
	expectedResult := &p2p.ConnectedPeersInfo{
		SelfShardID:              0,
		UnknownPeers:             []string{getExpectedConn(providedUnknown)},
		Seeders:                  []string{getExpectedConn(providedSeeder)},
		IntraShardValidators:     map[uint32][]string{0: {getExpectedConn(providedValidatorIntra)}},
		IntraShardObservers:      map[uint32][]string{0: {getExpectedConn(providedObserverIntra)}},
		CrossShardValidators:     map[uint32][]string{1: {getExpectedConn(providedValidatorCross)}},
		CrossShardObservers:      map[uint32][]string{1: {getExpectedConn(providedObserverCross)}},
		FullHistoryObservers:     map[uint32][]string{0: {getExpectedConn(providedFullHistory)}},
		NumValidatorsOnShard:     map[uint32]int{0: 1, 1: 1},
		NumObserversOnShard:      map[uint32]int{0: 2, 1: 1},
		NumPreferredPeersOnShard: map[uint32]int{0: 1},
		NumIntraShardValidators:  1,
		NumIntraShardObservers:   1,
		NumCrossShardValidators:  1,
		NumCrossShardObservers:   1,
		NumFullHistoryObservers:  1,
	}
	connPeersInfo := ch.GetConnectedPeersInfo()
	assert.Equal(t, expectedResult, connPeersInfo)
}

func getExpectedConn(pid peer.ID) string {
	return "multiaddr/p2p/" + pid.String()
}

func TestConnectionsHandler_IsConnectedToTheNetwork(t *testing.T) {
	t.Parallel()

	args := createMockArgConnectionsHandler()
	args.ConnMonitor = &mock.ConnectionMonitorStub{
		IsConnectedToTheNetworkCalled: func(netw network.Network) bool {
			return true
		},
	}
	ch, _ := libp2p.NewConnectionsHandler(args)
	assert.False(t, check.IfNil(ch))
	assert.True(t, ch.IsConnectedToTheNetwork())
}

func TestConnectionsHandler_SetThresholdMinConnectedPeers(t *testing.T) {
	t.Parallel()

	t.Run("negative threshold should error", func(t *testing.T) {
		t.Parallel()

		ch, _ := libp2p.NewConnectionsHandler(createMockArgConnectionsHandler())
		assert.False(t, check.IfNil(ch))
		err := ch.SetThresholdMinConnectedPeers(-1)
		assert.Equal(t, p2p.ErrInvalidValue, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedThreshold := 100
		args := createMockArgConnectionsHandler()
		wasCalled := false
		args.ConnMonitor = &mock.ConnectionMonitorStub{
			SetThresholdMinConnectedPeersCalled: func(thresholdMinConnectedPeers int, netw network.Network) {
				assert.Equal(t, providedThreshold, thresholdMinConnectedPeers)
				wasCalled = true
			},
			ThresholdMinConnectedPeersCalled: func() int {
				return providedThreshold
			},
		}
		ch, _ := libp2p.NewConnectionsHandler(args)
		assert.False(t, check.IfNil(ch))
		err := ch.SetThresholdMinConnectedPeers(providedThreshold)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
		threshold := ch.ThresholdMinConnectedPeers()
		assert.Equal(t, providedThreshold, threshold)
	})
}

func TestConnectionsHandler_Close(t *testing.T) {
	t.Parallel()

	t.Run("peers on channel close returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgConnectionsHandler()
		args.PeersOnChannel = &mock.PeersOnChannelStub{
			CloseCalled: func() error {
				return expectedErr
			},
		}
		ch, _ := libp2p.NewConnectionsHandler(args)
		assert.False(t, check.IfNil(ch))
		err := ch.Close()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("connection monitor close returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgConnectionsHandler()
		args.ConnMonitor = &mock.ConnectionMonitorStub{
			CloseCalled: func() error {
				return expectedErr
			},
		}
		ch, _ := libp2p.NewConnectionsHandler(args)
		assert.False(t, check.IfNil(ch))
		err := ch.Close()
		assert.Equal(t, expectedErr, err)
	})
}
