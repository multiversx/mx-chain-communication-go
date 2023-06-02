package facade

import (
	"testing"

	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/config"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/mock"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/stretchr/testify/require"
)

func createMockMessengerArgs() libp2p.ArgsNetworkMessenger {
	return libp2p.ArgsNetworkMessenger{
		Marshaller:    &testscommon.ProtoMarshallerMock{},
		ListenAddress: libp2p.TestListenAddrWithIp4AndTcp,
		P2pConfig: config.P2PConfig{
			Node: config.NodeConfig{
				Port: "0",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
		SyncTimer:             &libp2p.LocalSyncTimer{},
		PreferredPeersHolder:  &mock.PeersHolderStub{},
		PeersRatingHandler:    &mock.PeersRatingHandlerStub{},
		ConnectionWatcherType: p2p.ConnectionWatcherTypePrint,
		P2pPrivateKey:         mock.NewPrivateKeyMock(),
		P2pSingleSigner:       &mock.SingleSignerStub{},
		P2pKeyGenerator:       &mock.KeyGenStub{},
		MessengerType:         p2p.RegularMessenger,
	}
}

func TestNewNetworkMessengersFacade(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		netw := mocknet.New()
		m1, err := libp2p.NewMockMessenger(createMockMessengerArgs(), netw)
		require.NoError(t, err)

		facade, err := NewNetworkMessengersFacade(m1)
		require.NoError(t, err)
		require.NotNil(t, facade)
	})
}
