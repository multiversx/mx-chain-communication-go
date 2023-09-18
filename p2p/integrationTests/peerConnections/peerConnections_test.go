package peerConnections

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/config"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/mock"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/secp256k1"
	"github.com/multiversx/mx-chain-crypto-go/signing/secp256k1/singlesig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var keyGen = signing.NewKeyGenerator(secp256k1.NewSecp256k1())

func createBaseArgs() libp2p.ArgsNetworkMessenger {
	prvKey, _ := keyGen.GeneratePair()

	return libp2p.ArgsNetworkMessenger{
		Marshaller: &testscommon.MarshallerMock{},
		P2pConfig: config.P2PConfig{
			Node: config.NodeConfig{
				Port: "0", // auto-select port
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
		SyncTimer:             &mock.SyncTimerStub{},
		PreferredPeersHolder:  &mock.PeersHolderStub{},
		PeersRatingHandler:    &mock.PeersRatingHandlerStub{},
		ConnectionWatcherType: p2p.ConnectionWatcherTypePrint,
		P2pPrivateKey:         prvKey,
		P2pSingleSigner:       &singlesig.Secp256k1Signer{},
		P2pKeyGenerator:       keyGen,
		Logger:                &testscommon.LoggerStub{},
	}
}

func TestPeerConnectionsOnAllSupportedProtocolsShouldExchangeData(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	messengers := make([]p2p.Messenger, 0)

	seederArgs := createBaseArgs()
	seederArgs.P2pConfig.Node.Transports = config.TransportConfig{
		TCP: config.TCPProtocolConfig{
			ListenAddress: p2p.LocalHostListenAddrWithIp4AndTcp,
		},
		QUICAddress:         "/ip4/127.0.0.1/udp/%d/quic-v1",
		WebSocketAddress:    "/ip4/127.0.0.1/tcp/%d/ws",
		WebTransportAddress: "/ip4/127.0.0.1/udp/%d/quic-v1/webtransport",
	}
	seeder, err := libp2p.NewNetworkMessenger(seederArgs)
	require.Nil(t, err)
	messengers = append(messengers, seeder)

	tcpPeerArgs := createBaseArgs()
	tcpPeerArgs.P2pConfig.Node.Transports = config.TransportConfig{
		TCP: config.TCPProtocolConfig{
			ListenAddress: p2p.LocalHostListenAddrWithIp4AndTcp,
		},
	}
	tcpPeer, err := libp2p.NewNetworkMessenger(tcpPeerArgs)
	require.Nil(t, err)
	addressToConnect := getAddressMatching(seeder.Addresses(), "/tcp/", "/ws/")
	err = tcpPeer.ConnectToPeer(addressToConnect)
	require.Nil(t, err)
	messengers = append(messengers, tcpPeer)

	quicPeerArgs := createBaseArgs()
	quicPeerArgs.P2pConfig.Node.Transports = config.TransportConfig{
		QUICAddress: "/ip4/127.0.0.1/udp/%d/quic-v1",
	}
	quicPeer, err := libp2p.NewNetworkMessenger(quicPeerArgs)
	require.Nil(t, err)
	addressToConnect = getAddressMatching(seeder.Addresses(), "/quic-v1/", "webtransport")
	err = quicPeer.ConnectToPeer(addressToConnect)
	require.Nil(t, err)
	messengers = append(messengers, quicPeer)

	wsPeerArgs := createBaseArgs()
	wsPeerArgs.P2pConfig.Node.Transports = config.TransportConfig{
		WebSocketAddress: "/ip4/127.0.0.1/tcp/%d/ws",
	}
	wsPeer, err := libp2p.NewNetworkMessenger(wsPeerArgs)
	require.Nil(t, err)
	addressToConnect = getAddressMatching(seeder.Addresses(), "/ws/", "")
	err = wsPeer.ConnectToPeer(addressToConnect)
	require.Nil(t, err)
	messengers = append(messengers, wsPeer)

	webTransportPeerArgs := createBaseArgs()
	webTransportPeerArgs.P2pConfig.Node.Transports = config.TransportConfig{
		WebTransportAddress: "/ip4/127.0.0.1/udp/%d/quic-v1/webtransport",
	}
	webTransportPeer, err := libp2p.NewNetworkMessenger(webTransportPeerArgs)
	require.Nil(t, err)
	addressToConnect = getAddressMatching(seeder.Addresses(), "webtransport", "")
	err = webTransportPeer.ConnectToPeer(addressToConnect)
	require.Nil(t, err)

	// create a common topic on all messengers
	commonTopic := "test"
	for _, mes := range messengers {
		err = mes.CreateTopic(commonTopic, true)
		require.Nil(t, err)
	}

	// setup interceptors
	mutMessages := sync.Mutex{}
	messages := make(map[string]map[string]int)

	for _, mes := range messengers {
		err = mes.RegisterMessageProcessor(commonTopic, "", createInterceptor(mes.ID().Pretty(), messages, &mutMessages))
		require.Nil(t, err)
	}

	time.Sleep(time.Second * 2) // allow topic setup

	// all messengers broadcast an unique message
	for idx, mes := range messengers {
		mes.Broadcast(commonTopic, []byte(fmt.Sprintf("message %d", idx)))
	}

	time.Sleep(time.Second * 2) // allow data to be passed among peers

	mutMessages.Lock()
	assert.Equal(t, len(messengers), len(messages)) // all hosts should have created an entry in the map (key == ID)
	for _, numMessagesMap := range messages {
		assert.Equal(t, len(messengers), len(numMessagesMap)) // on each host, should have received the required number of messages (key == message xxx)
		for _, numInt := range numMessagesMap {
			assert.Equal(t, 1, numInt) // each message should have been received exactly once
		}
	}

	mutMessages.Unlock()

	for _, mes := range messengers {
		_ = mes.Close()
	}
}

func getAddressMatching(addresses []string, including string, excluding string) string {
	for _, addr := range addresses {
		if len(including) > 0 {
			if !strings.Contains(addr, including) {
				continue
			}
		}
		if len(excluding) > 0 {
			if strings.Contains(addr, excluding) {
				continue
			}
		}

		return addr
	}

	return ""
}

func createInterceptor(hostName string, dataMap map[string]map[string]int, mut *sync.Mutex) p2p.MessageProcessor {
	return &mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) error {
			mut.Lock()
			numMessagesMap := dataMap[hostName]
			if numMessagesMap == nil {
				numMessagesMap = make(map[string]int)
				dataMap[hostName] = numMessagesMap
			}

			numMessagesMap[string(message.Data())]++
			mut.Unlock()

			return nil
		},
	}
}
