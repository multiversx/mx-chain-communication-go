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
		tracer.setDebugger(&discardedDebuggerStub{
			addIgnoredMessageCalled: func(topic string, size uint64) {
				*numCalls++
				*recordedTopic = topic
				*recordedSize = size
			},
			addDuplicateMessageCalled: func(_ string, _ uint64) {
				require.Fail(t, "should not have recorded a duplicate")
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
	t.Run("discarded messages debugger is kept", func(t *testing.T) {
		t.Parallel()

		tracer := newPubsubTracer()

		assert.True(t, tracer.setDebugger(&discardedDebuggerStub{}))
		assert.NotNil(t, tracer.debugger)
	})
}

func TestPubsubTracer_DuplicateMessageShouldForward(t *testing.T) {
	t.Parallel()

	recordedTopic := ""
	recordedSize := uint64(0)
	numCalls := 0
	debugger := &discardedDebuggerStub{
		addDuplicateMessageCalled: func(topic string, size uint64) {
			recordedTopic = topic
			recordedSize = size
			numCalls++
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
