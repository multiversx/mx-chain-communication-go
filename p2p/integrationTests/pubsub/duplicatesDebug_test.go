package peerDisconnecting

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/integrationTests"
)

type discardedDebuggerStub struct {
	mut           sync.Mutex
	numDuplicates map[string]int
	sizes         map[string]uint64
	numIgnored    map[string]int
	ignoredSizes  map[string]uint64
}

func newDiscardedDebuggerStub() *discardedDebuggerStub {
	return &discardedDebuggerStub{
		numDuplicates: make(map[string]int),
		sizes:         make(map[string]uint64),
		numIgnored:    make(map[string]int),
		ignoredSizes:  make(map[string]uint64),
	}
}

func (stub *discardedDebuggerStub) AddIncomingMessage(_ string, _ uint64, _ bool) {}
func (stub *discardedDebuggerStub) AddOutgoingMessage(_ string, _ uint64, _ bool) {}
func (stub *discardedDebuggerStub) Close() error                                  { return nil }
func (stub *discardedDebuggerStub) IsInterfaceNil() bool                          { return stub == nil }

func (stub *discardedDebuggerStub) AddDuplicateMessage(topic string, size uint64) {
	stub.mut.Lock()
	defer stub.mut.Unlock()

	stub.numDuplicates[topic]++
	stub.sizes[topic] += size
}

func (stub *discardedDebuggerStub) AddIgnoredMessage(topic string, size uint64) {
	stub.mut.Lock()
	defer stub.mut.Unlock()

	stub.numIgnored[topic]++
	stub.ignoredSizes[topic] += size
}

func (stub *discardedDebuggerStub) get(topic string) (int, uint64) {
	stub.mut.Lock()
	defer stub.mut.Unlock()

	return stub.numDuplicates[topic], stub.sizes[topic]
}

func (stub *discardedDebuggerStub) getIgnored(topic string) (int, uint64) {
	stub.mut.Lock()
	defer stub.mut.Unlock()

	return stub.numIgnored[topic], stub.ignoredSizes[topic]
}

// TestDuplicatedMessagesAreRecordedByTheDebugger checks the whole chain: pubsub raw tracer -> messenger -> debugger.
// In a fully connected mesh of 3 peers, the 2 non-publishing peers forward the message to each other, so both
// record exactly one duplicate per broadcast.
func TestDuplicatedMessagesAreRecordedByTheDebugger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfPeers := 3
	peers := make([]p2p.Messenger, numOfPeers)
	for i := 0; i < numOfPeers; i++ {
		peers[i] = integrationTests.CreateMessengerWithNoDiscovery()
	}

	defer func() {
		for _, peerInstance := range peers {
			if peerInstance != nil {
				_ = peerInstance.Close()
			}
		}
	}()

	for i := 0; i < numOfPeers; i++ {
		for j := i + 1; j < numOfPeers; j++ {
			err := peers[i].ConnectToPeer(peers[j].Addresses()[0])
			require.Nil(t, err)
		}
	}

	testTopic := "test"
	for _, peerInstance := range peers {
		err := peerInstance.CreateTopic(testTopic, true)
		require.Nil(t, err)

		err = peerInstance.RegisterMessageProcessor(testTopic, "test", &messageProcessorStub{
			ProcessReceivedMessageCalled: func(_ p2p.MessageP2P, _ p2p.MessageHandler) ([]byte, error) {
				return []byte{}, nil
			},
		})
		require.Nil(t, err)
	}

	debugger := newDiscardedDebuggerStub()
	err := peers[0].SetDebugger(debugger)
	require.Nil(t, err)

	// let the gossipsub mesh form
	time.Sleep(time.Second * 3)

	numBroadcasts := 5
	payload := []byte("this is a test message used to check the duplicates accounting")
	for i := 0; i < numBroadcasts; i++ {
		peers[1].Broadcast(testTopic, payload)
		time.Sleep(time.Millisecond * 200)
	}

	require.Eventually(t, func() bool {
		num, _ := debugger.get(testTopic)
		return num > 0
	}, time.Second*10, time.Millisecond*100, "no duplicate was recorded")

	num, size := debugger.get(testTopic)
	fmt.Printf("recorded %d duplicates, %d bytes on topic %s\n", num, size, testTopic)
	require.True(t, size >= uint64(len(payload)))
}

var _ p2p.DiscardedMessagesDebugger = (*discardedDebuggerStub)(nil)

// TestIgnoredMessagesAreRecordedByTheDebugger covers the equivalent messages case: distinct pubsub messages
// (different originators and sequence numbers, so the pubsub deduplication does not catch them) that the node
// maps to the same message id. Only the first one is accepted, the rest are ignored and not propagated.
func TestIgnoredMessagesAreRecordedByTheDebugger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfPeers := 2
	peers := make([]p2p.Messenger, numOfPeers)
	for i := 0; i < numOfPeers; i++ {
		peers[i] = integrationTests.CreateMessengerWithNoDiscovery()
	}

	defer func() {
		for _, peerInstance := range peers {
			if peerInstance != nil {
				_ = peerInstance.Close()
			}
		}
	}()

	err := peers[0].ConnectToPeer(peers[1].Addresses()[0])
	require.Nil(t, err)

	testTopic := "test"
	equivalentMessageID := []byte("same message id for every received message")
	for idx, peerInstance := range peers {
		err = peerInstance.CreateTopic(testTopic, true)
		require.Nil(t, err)

		returnedMessageID := make([]byte, 0)
		if idx == 0 {
			returnedMessageID = equivalentMessageID
		}

		err = peerInstance.RegisterMessageProcessor(testTopic, "test", &messageProcessorStub{
			ProcessReceivedMessageCalled: func(_ p2p.MessageP2P, _ p2p.MessageHandler) ([]byte, error) {
				return returnedMessageID, nil
			},
		})
		require.Nil(t, err)
	}

	debugger := newDiscardedDebuggerStub()
	err = peers[0].SetDebugger(debugger)
	require.Nil(t, err)

	// let the gossipsub mesh form
	time.Sleep(time.Second * 3)

	numBroadcasts := 5
	for i := 0; i < numBroadcasts; i++ {
		peers[1].Broadcast(testTopic, []byte(fmt.Sprintf("distinct payload number %d", i)))
		time.Sleep(time.Millisecond * 200)
	}

	require.Eventually(t, func() bool {
		num, _ := debugger.getIgnored(testTopic)
		return num == numBroadcasts-1
	}, time.Second*10, time.Millisecond*100, "expected all but the first message to be ignored")

	numIgnored, ignoredSize := debugger.getIgnored(testTopic)
	numDuplicates, _ := debugger.get(testTopic)
	fmt.Printf("recorded %d ignored (%d bytes) and %d duplicates on topic %s\n",
		numIgnored, ignoredSize, numDuplicates, testTopic)

	// the payloads are distinct messages, so the pubsub deduplication must not have caught any of them
	require.Zero(t, numDuplicates)
	require.True(t, ignoredSize > 0)
}
