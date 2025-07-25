package libp2p_test

import (
	"bytes"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/config"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/mock"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
)

func createMessenger() p2p.Messenger {
	args := libp2p.ArgsNetworkMessenger{
		Marshaller: &testscommon.ProtoMarshallerMock{},
		P2pConfig: config.P2PConfig{
			Node: config.NodeConfig{
				Port:       "0",
				Transports: createTestTCPTransportConfig(),
				ResourceLimiter: config.ResourceLimiterConfig{
					Type: p2p.DefaultAutoscaleResourceLimiter,
				},
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
		Logger:                &testscommon.LoggerStub{},
	}

	libP2PMes, err := libp2p.NewNetworkMessenger(args)
	if err != nil {
		fmt.Println(err.Error())
	}

	return libP2PMes
}

// TestIssueEN898_StreamResetError emphasizes what happens if direct sender writes to a stream that has been reset
// Testing is done by writing a large buffer that will cause the recipient to reset its inbound stream
// Sender will then be notified that the stream writing did not succeed but it will only log the error
// Next message that the sender tries to send will cause a new error to be logged and no data to be sent
// The fix consists in the full stream closing when an error occurs during writing.
func TestIssueEN898_StreamResetError(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	mes1 := createMessenger()
	mes2 := createMessenger()

	defer func() {
		_ = mes1.Close()
		_ = mes2.Close()
	}()

	_ = mes1.ConnectToPeer(getConnectableAddress(mes2))

	topic := "test topic"

	size4MB := 1 << 22
	size4kB := 1 << 12

	// a 4MB slice containing character A
	largePacket := bytes.Repeat([]byte{65}, size4MB)
	// a 4kB slice containing character B
	smallPacket := bytes.Repeat([]byte{66}, size4kB)

	largePacketReceived := &atomic.Value{}
	largePacketReceived.Store(false)

	smallPacketReceived := &atomic.Value{}
	smallPacketReceived.Store(false)

	_ = mes2.CreateTopic(topic, false)
	_ = mes2.RegisterMessageProcessor(topic, "identifier", &mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, _ core.PeerID, _ p2p.MessageHandler) ([]byte, error) {
			if bytes.Equal(message.Data(), largePacket) {
				largePacketReceived.Store(true)
			}

			if bytes.Equal(message.Data(), smallPacket) {
				smallPacketReceived.Store(true)
			}

			return nil, nil
		},
	})

	fmt.Println("sending the large packet...")
	_ = mes1.SendToConnectedPeer(topic, largePacket, mes2.ID())

	time.Sleep(time.Second)

	fmt.Println("sending the small packet...")
	_ = mes1.SendToConnectedPeer(topic, smallPacket, mes2.ID())

	time.Sleep(time.Second)

	assert.False(t, largePacketReceived.Load().(bool))
	assert.True(t, smallPacketReceived.Load().(bool))
}
