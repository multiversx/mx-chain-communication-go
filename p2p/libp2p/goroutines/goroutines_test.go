package goroutines

import (
	"bytes"
	"fmt"
	"runtime"
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
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTopic = "test"
const timeoutWaitResponses = time.Second * 2
const directSenderMarker = "libp2p.(*directSender).directStreamHandler"

var log = logger.GetOrCreate("p2p/libp2p/goroutines")
var keyGen = signing.NewKeyGenerator(secp256k1.NewSecp256k1())

func createDefaultConfig() config.P2PConfig {
	return config.P2PConfig{
		Node: config.NodeConfig{
			Transports: config.TransportConfig{
				TCP: config.TCPProtocolConfig{
					ListenAddress: p2p.LocalHostListenAddrWithIp4AndTcp,
				},
			},
			Port: "0",
		},
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:                          true,
			Type:                             "optimized",
			RefreshIntervalInSec:             1,
			RoutingTableRefreshIntervalInSec: 1,
			ProtocolID:                       "/erd/kad/1.0.0",
			InitialPeerList:                  nil,
			BucketSize:                       100,
		},
	}
}

func closeMessengers(messengers ...p2p.Messenger) {
	for _, mes := range messengers {
		_ = mes.Close()
	}
}

func prepareMessengerForMatchDataReceive(messenger p2p.Messenger, matchData []byte, wg *sync.WaitGroup, checkSigSize func(sigSize int) bool) {
	_ = messenger.CreateTopic(testTopic, false)

	_ = messenger.RegisterMessageProcessor(testTopic, "identifier",
		&mock.MessageProcessorStub{
			ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) error {
				if !bytes.Equal(matchData, message.Data()) {
					return nil
				}
				if !checkSigSize(len(message.Signature())) {
					return nil
				}

				// do not print the message.Data() or matchData as the test TestLibp2pMessenger_BroadcastDataBetween2PeersWithLargeMsgShouldWork
				// will cause disruption on the GitHub action page

				wg.Done()

				return nil
			},
		})
}

func waitDoneWithTimeout(t *testing.T, chanDone chan bool, timeout time.Duration) {
	select {
	case <-chanDone:
		return
	case <-time.After(timeout):
		assert.Fail(t, "timeout reached")
	}
}

func TestDisconnectWillCloseGoRoutines(t *testing.T) {
	// do not run the tests from this package in parallel, or otherwise, they will alter the results of each other

	msg := []byte("test message")

	p2pConfig := createDefaultConfig()
	p2pConfig.Sharding = config.ShardingConfig{
		TargetPeerCount:         100,
		MaxIntraShardValidators: 40,
		MaxCrossShardValidators: 40,
		MaxIntraShardObservers:  1,
		MaxCrossShardObservers:  1,
		MaxSeeders:              1,
		Type:                    p2p.ListsSharder,
	}
	p2pConfig.Node.ThresholdMinConnectedPeers = 1

	args := libp2p.ArgsNetworkMessenger{
		P2pConfig:             p2pConfig,
		PreferredPeersHolder:  &mock.PeersHolderStub{},
		Marshaller:            &testscommon.MarshallerMock{},
		SyncTimer:             &mock.SyncTimerStub{},
		PeersRatingHandler:    &mock.PeersRatingHandlerStub{},
		ConnectionWatcherType: p2p.ConnectionWatcherTypePrint,
		P2pPrivateKey:         mock.NewPrivateKeyMock(),
		P2pSingleSigner:       &singlesig.Secp256k1Signer{},
		P2pKeyGenerator:       keyGen,
		Logger:                &testscommon.LoggerStub{},
	}
	args.P2pPrivateKey, _ = keyGen.GeneratePair()

	messenger1, err := libp2p.NewTestNetMessenger(args)
	require.Nil(t, err)

	args.P2pPrivateKey, _ = keyGen.GeneratePair()
	messenger2, err := libp2p.NewTestNetMessenger(args)
	require.Nil(t, err)
	adr2 := messenger2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	err = messenger1.ConnectToPeer(adr2)
	require.Nil(t, err)

	wg := &sync.WaitGroup{}
	chanDone := make(chan bool)
	wg.Add(1)

	go func() {
		wg.Wait()
		chanDone <- true
	}()

	minimumSigSize := 70
	err = messenger1.CreateTopic(testTopic, false)
	require.Nil(t, err)
	prepareMessengerForMatchDataReceive(
		messenger2,
		msg,
		wg,
		func(sigSize int) bool {
			return sigSize >= minimumSigSize // message has a non-empty signature
		},
	)

	log.Info("Delaying as to allow peers to announce themselves on the opened topic...")
	time.Sleep(time.Second)

	log.Info("sending message", "from", messenger1.ID().Pretty(), "to", messenger2.ID().Pretty())

	err = messenger1.SendToConnectedPeer("test", msg, messenger2.ID())
	assert.Nil(t, err)

	waitDoneWithTimeout(t, chanDone, timeoutWaitResponses)
	assert.Equal(t, 2, getDirectSenderRunningGoRoutines()) // direct sender go routines running

	err = messenger1.ClosePeer(messenger2.ID())
	assert.Nil(t, err)

	log.Info("Delaying as to allow peers to close everything...")
	time.Sleep(time.Second)

	assert.Equal(t, 0, getDirectSenderRunningGoRoutines()) // direct sender go routines should have ended

	closeMessengers(messenger1, messenger2)
}

func getDirectSenderRunningGoRoutines() int {
	buffSize := 10 * 1024 * 1024 // 10MB
	buff := make([]byte, buffSize)
	n := runtime.Stack(buff, true)
	goRoutinesDump := string(buff[:n])

	return strings.Count(goRoutinesDump, directSenderMarker)
}
