package libp2p

import (
	"testing"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsubPb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type plainDebuggerStub struct{}

func (stub *plainDebuggerStub) AddIncomingMessage(_ string, _ uint64, _ bool) {}
func (stub *plainDebuggerStub) AddOutgoingMessage(_ string, _ uint64, _ bool) {}
func (stub *plainDebuggerStub) Close() error                                  { return nil }
func (stub *plainDebuggerStub) IsInterfaceNil() bool                          { return stub == nil }

type discardedDebuggerStub struct {
	plainDebuggerStub
	addDuplicateMessageCalled func(topic string, size uint64)
	addIgnoredMessageCalled   func(topic string, size uint64)
}

func (stub *discardedDebuggerStub) AddDuplicateMessage(topic string, size uint64) {
	if stub.addDuplicateMessageCalled != nil {
		stub.addDuplicateMessageCalled(topic, size)
	}
}

func (stub *discardedDebuggerStub) AddIgnoredMessage(topic string, size uint64) {
	if stub.addIgnoredMessageCalled != nil {
		stub.addIgnoredMessageCalled(topic, size)
	}
}

func (stub *discardedDebuggerStub) IsInterfaceNil() bool { return stub == nil }

// discardedOnlyDebuggerStub records the discarded messages but not the RPC traffic
type discardedOnlyDebuggerStub struct {
	discardedDebuggerStub
}

type fullDebuggerStub struct {
	discardedDebuggerStub
	isRecording           bool
	addRPCPublishedCalled func(topic string, size uint64, isIncoming bool)
	addRPCControlCalled   func(topic string, size uint64, isIncoming bool)
}

func (stub *fullDebuggerStub) IsRecording() bool { return stub.isRecording }

func (stub *fullDebuggerStub) AddRPCPublishedMessage(topic string, size uint64, isIncoming bool) {
	if stub.addRPCPublishedCalled != nil {
		stub.addRPCPublishedCalled(topic, size, isIncoming)
	}
}

func (stub *fullDebuggerStub) AddRPCControlMessage(topic string, size uint64, isIncoming bool) {
	if stub.addRPCControlCalled != nil {
		stub.addRPCControlCalled(topic, size, isIncoming)
	}
}

func (stub *fullDebuggerStub) IsInterfaceNil() bool { return stub == nil }

func createRPC(topic string, data []byte, numIhaveIDs int, numIwantIDs int) *pubsub.RPC {
	rpc := &pubsub.RPC{}
	rpc.Publish = []*pubsubPb.Message{{Topic: &topic, Data: data}}

	ihaveIDs := make([]string, numIhaveIDs)
	for i := range ihaveIDs {
		ihaveIDs[i] = "ihave message id"
	}
	iwantIDs := make([]string, numIwantIDs)
	for i := range iwantIDs {
		iwantIDs[i] = "iwant message id"
	}

	rpc.Control = &pubsubPb.ControlMessage{
		Ihave: []*pubsubPb.ControlIHave{{TopicID: &topic, MessageIDs: ihaveIDs}},
		Iwant: []*pubsubPb.ControlIWant{{MessageIDs: iwantIDs}},
		Graft: []*pubsubPb.ControlGraft{{TopicID: &topic}},
	}

	return rpc
}

func createPubsubMessage(topic string, data []byte) *pubsub.Message {
	return &pubsub.Message{
		Message: &pubsubPb.Message{
			Topic: &topic,
			Data:  data,
		},
	}
}

func TestPubsubTracer_ImplementsRawTracer(t *testing.T) {
	t.Parallel()

	var _ pubsub.RawTracer = newPubsubTracer()
}

func TestPubsubTracer_DuplicateMessageShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		assert.Nil(t, r)
	}()

	var nilTracer *pubsubTracer
	nilTracer.setDebugger(&plainDebuggerStub{})
	nilTracer.DuplicateMessage(createPubsubMessage("topic", []byte("data")))
	nilTracer.RejectMessage(createPubsubMessage("topic", []byte("data")), pubsub.RejectValidationIgnored)

	tracer := newPubsubTracer()
	tracer.DuplicateMessage(createPubsubMessage("topic", []byte("data")))
	tracer.DuplicateMessage(nil)
	tracer.RejectMessage(createPubsubMessage("topic", []byte("data")), pubsub.RejectValidationIgnored)
	tracer.RejectMessage(nil, pubsub.RejectValidationIgnored)
}

func TestPubsubTracer_RejectMessage(t *testing.T) {
	t.Parallel()

	createTracer := func(numCalls *int, recordedTopic *string, recordedSize *uint64) *pubsubTracer {
		tracer := newPubsubTracer()
		tracer.setDebugger(&fullDebuggerStub{
			discardedDebuggerStub: discardedDebuggerStub{
				addIgnoredMessageCalled: func(topic string, size uint64) {
					*numCalls++
					*recordedTopic = topic
					*recordedSize = size
				},
				addDuplicateMessageCalled: func(_ string, _ uint64) {
					require.Fail(t, "should not have recorded a duplicate")
				},
			},
		})

		return tracer
	}

	t.Run("ignored messages are recorded", func(t *testing.T) {
		t.Parallel()

		numCalls, recordedTopic, recordedSize := 0, "", uint64(0)
		tracer := createTracer(&numCalls, &recordedTopic, &recordedSize)

		data := []byte("message data")
		tracer.RejectMessage(createPubsubMessage("testTopic", data), pubsub.RejectValidationIgnored)

		assert.Equal(t, 1, numCalls)
		assert.Equal(t, "testTopic", recordedTopic)
		assert.Equal(t, uint64(len(data)), recordedSize)
	})
	t.Run("other reject reasons are not recorded", func(t *testing.T) {
		t.Parallel()

		numCalls, recordedTopic, recordedSize := 0, "", uint64(0)
		tracer := createTracer(&numCalls, &recordedTopic, &recordedSize)

		reasons := []string{
			pubsub.RejectValidationFailed,
			pubsub.RejectValidationThrottled,
			pubsub.RejectValidationQueueFull,
			pubsub.RejectSelfOrigin,
		}
		for _, reason := range reasons {
			tracer.RejectMessage(createPubsubMessage("testTopic", []byte("data")), reason)
		}

		assert.Zero(t, numCalls)
	})
}

func TestPubsubTracer_setDebugger(t *testing.T) {
	t.Parallel()

	t.Run("debugger not recording discarded messages is not kept", func(t *testing.T) {
		t.Parallel()

		tracer := newPubsubTracer()

		assert.False(t, tracer.setDebugger(&plainDebuggerStub{}))
		assert.Nil(t, tracer.debugger)
	})
	t.Run("nil discarded messages debugger is not kept", func(t *testing.T) {
		t.Parallel()

		var nilDebugger *discardedDebuggerStub

		tracer := newPubsubTracer()

		assert.False(t, tracer.setDebugger(nilDebugger))
		assert.Nil(t, tracer.debugger)
	})
	t.Run("debugger recording everything is kept", func(t *testing.T) {
		t.Parallel()

		tracer := newPubsubTracer()

		assert.True(t, tracer.setDebugger(&fullDebuggerStub{}))
		assert.NotNil(t, tracer.debugger)
		assert.NotNil(t, tracer.rpcDebugger)
	})
	t.Run("debugger without RPC recording keeps only the discarded part", func(t *testing.T) {
		t.Parallel()

		tracer := newPubsubTracer()

		assert.False(t, tracer.setDebugger(&discardedOnlyDebuggerStub{}))
		assert.NotNil(t, tracer.debugger)
		assert.Nil(t, tracer.rpcDebugger)
	})
}

func TestPubsubTracer_DuplicateMessageShouldForward(t *testing.T) {
	t.Parallel()

	recordedTopic := ""
	recordedSize := uint64(0)
	numCalls := 0
	debugger := &fullDebuggerStub{
		discardedDebuggerStub: discardedDebuggerStub{
			addDuplicateMessageCalled: func(topic string, size uint64) {
				recordedTopic = topic
				recordedSize = size
				numCalls++
			},
		},
	}

	tracer := newPubsubTracer()
	tracer.setDebugger(debugger)

	data := []byte("message data")
	tracer.DuplicateMessage(createPubsubMessage("testTopic", data))

	assert.Equal(t, 1, numCalls)
	assert.Equal(t, "testTopic", recordedTopic)
	assert.Equal(t, uint64(len(data)), recordedSize)
}

func TestPubsubTracer_recordRPC(t *testing.T) {
	t.Parallel()

	type record struct {
		topic      string
		size       uint64
		isIncoming bool
	}

	createTracer := func(isRecording bool, published *[]record, control *[]record) *pubsubTracer {
		tracer := newPubsubTracer()
		tracer.setDebugger(&fullDebuggerStub{
			isRecording: isRecording,
			addRPCPublishedCalled: func(topic string, size uint64, isIncoming bool) {
				*published = append(*published, record{topic, size, isIncoming})
			},
			addRPCControlCalled: func(topic string, size uint64, isIncoming bool) {
				*control = append(*control, record{topic, size, isIncoming})
			},
		})

		return tracer
	}

	t.Run("nothing is recorded while not recording", func(t *testing.T) {
		t.Parallel()

		published, control := make([]record, 0), make([]record, 0)
		tracer := createTracer(false, &published, &control)

		tracer.RecvRPC(createRPC("testTopic", []byte("data"), 2, 1))
		tracer.SendRPC(createRPC("testTopic", []byte("data"), 2, 1), "pid")

		assert.Empty(t, published)
		assert.Empty(t, control)
	})
	t.Run("incoming RPC is split by topic and kind", func(t *testing.T) {
		t.Parallel()

		published, control := make([]record, 0), make([]record, 0)
		tracer := createTracer(true, &published, &control)

		tracer.RecvRPC(createRPC("testTopic", []byte("data"), 2, 1))

		require.Len(t, published, 1)
		assert.Equal(t, "testTopic", published[0].topic)
		assert.True(t, published[0].isIncoming)
		assert.Greater(t, published[0].size, uint64(0))

		// ihave and graft carry the topic, iwant does not
		require.Len(t, control, 3)
		assert.Equal(t, "testTopic", control[0].topic)
		assert.Equal(t, "testTopic", control[1].topic)
		assert.Equal(t, controlWithoutTopic, control[2].topic)
		for _, c := range control {
			assert.True(t, c.isIncoming)
			assert.Greater(t, c.size, uint64(0))
		}
	})
	t.Run("outgoing RPC is marked as outgoing", func(t *testing.T) {
		t.Parallel()

		published, control := make([]record, 0), make([]record, 0)
		tracer := createTracer(true, &published, &control)

		tracer.SendRPC(createRPC("testTopic", []byte("data"), 2, 1), "pid")

		require.Len(t, published, 1)
		assert.False(t, published[0].isIncoming)
		require.NotEmpty(t, control)
		assert.False(t, control[0].isIncoming)
	})
	t.Run("larger ihave lists produce larger recorded sizes", func(t *testing.T) {
		t.Parallel()

		small, big := make([]record, 0), make([]record, 0)
		unused := make([]record, 0)

		createTracer(true, &unused, &small).RecvRPC(createRPC("testTopic", []byte("data"), 2, 0))
		createTracer(true, &unused, &big).RecvRPC(createRPC("testTopic", []byte("data"), 200, 0))

		require.NotEmpty(t, small)
		require.NotEmpty(t, big)
		assert.Greater(t, big[0].size, small[0].size)
	})
	t.Run("should not panic on nil or empty RPCs", func(t *testing.T) {
		t.Parallel()

		defer func() {
			assert.Nil(t, recover())
		}()

		published, control := make([]record, 0), make([]record, 0)
		tracer := createTracer(true, &published, &control)

		tracer.RecvRPC(nil)
		tracer.SendRPC(nil, "pid")
		tracer.RecvRPC(&pubsub.RPC{})

		var nilTracer *pubsubTracer
		nilTracer.RecvRPC(createRPC("testTopic", []byte("data"), 1, 1))
	})
}
