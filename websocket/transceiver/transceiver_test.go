package transceiver

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
	webSocket "github.com/multiversx/mx-chain-communication-go/websocket"
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsTransceiver {
	payloadConverter, _ := webSocket.NewWebSocketPayloadConverter(&testscommon.MarshallerMock{})
	return ArgsTransceiver{
		BlockingAckOnError: false,
		PayloadConverter:   payloadConverter,
		Log:                &testscommon.LoggerMock{},
		RetryDurationInSec: 1,
		AckTimeoutInSec:    2,
		WithAcknowledge:    false,
	}
}

func TestNewReceiver(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		args := createArgs()
		ws, err := NewTransceiver(args)
		require.NotNil(t, ws)
		require.Nil(t, err)
	})

	t.Run("empty logger, should return error", func(t *testing.T) {
		args := createArgs()
		args.Log = nil
		ws, err := NewTransceiver(args)
		require.Nil(t, ws)
		require.Equal(t, core.ErrNilLogger, err)
	})

	t.Run("nil payload converter, should return error", func(t *testing.T) {
		args := createArgs()
		args.PayloadConverter = nil
		ws, err := NewTransceiver(args)
		require.Nil(t, ws)
		require.Equal(t, data.ErrNilPayloadConverter, err)
	})

	t.Run("zero retry duration in seconds, should return error", func(t *testing.T) {
		args := createArgs()
		args.RetryDurationInSec = 0
		ws, err := NewTransceiver(args)
		require.Nil(t, ws)
		require.Equal(t, data.ErrZeroValueRetryDuration, err)
	})
	t.Run("zero acknowledge timeout, should return error", func(t *testing.T) {
		args := createArgs()
		args.AckTimeoutInSec = 0
		args.WithAcknowledge = true
		ws, err := NewTransceiver(args)
		require.Nil(t, ws)
		require.Equal(t, data.ErrZeroValueAckTimeout, err)
	})
}

func TestReceiver_ListenAndClose(t *testing.T) {
	args := createArgs()
	webSocketsReceiver, err := NewTransceiver(args)
	require.Nil(t, err)

	count := uint64(0)
	conn := &testscommon.WebsocketConnectionStub{
		ReadMessageCalled: func() (messageType int, payload []byte, err error) {
			time.Sleep(time.Second)
			if atomic.LoadUint64(&count) == 1 {
				return 0, nil, errors.New("closed")
			}
			return 0, nil, nil
		},
		CloseCalled: func() error {
			atomic.AddUint64(&count, 1)
			return nil
		},
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		webSocketsReceiver.Listen(conn)
		wg.Done()
		atomic.AddUint64(&count, 1)
	}()

	_ = webSocketsReceiver.Close()
	_ = conn.Close()
	wg.Wait()

	require.Equal(t, uint64(2), atomic.LoadUint64(&count))
}

func TestReceiver_ListenAndSendAck(t *testing.T) {
	args := createArgs()
	webSocketsReceiver, err := NewTransceiver(args)
	require.Nil(t, err)

	_ = webSocketsReceiver.SetPayloadHandler(&testscommon.PayloadHandlerStub{
		ProcessPayloadCalled: func(_ []byte, _ string, _ uint32) error {
			return nil
		},
	})

	mutex := sync.Mutex{}
	count := uint64(0)
	conn := &testscommon.WebsocketConnectionStub{
		ReadMessageCalled: func() (int, []byte, error) {
			mutex.Lock()
			time.Sleep(500 * time.Millisecond)
			if count == 2 {
				mutex.Unlock()
				return 0, nil, errors.New("closed")
			}
			count++
			mutex.Unlock()

			preparedPayload, _ := args.PayloadConverter.ConstructPayload(&data.WsMessage{
				Payload:         []byte("something"),
				Topic:           outport.TopicSaveAccounts,
				Counter:         10,
				WithAcknowledge: true,
				Type:            data.PayloadMessage,
			})
			return websocket.BinaryMessage, preparedPayload, nil
		},
		CloseCalled: func() error {
			return nil
		},
		WriteMessageCalled: func(messageType int, d []byte) error {
			return nil
		},
	}

	_ = webSocketsReceiver.Listen(conn)

	_ = webSocketsReceiver.Close()
	_ = conn.Close()

	require.GreaterOrEqual(t, count, uint64(2))
}

func TestSender_AddConnectionSendAndClose(t *testing.T) {
	args := createArgs()
	args.WithAcknowledge = true
	webSocketTransceiver, _ := NewTransceiver(args)

	write := false
	readAck := false
	conn1 := &testscommon.WebsocketConnectionStub{
		GetIDCalled: func() string {
			return "conn1"
		},
		WriteMessageCalled: func(messageType int, data []byte) error {
			write = true
			return nil
		},
		ReadMessageCalled: func() (messageType int, payload []byte, err error) {
			if readAck {
				wsMessage := &data.WsMessage{
					Counter: 1,
					Type:    data.AckMessage,
				}
				counterBytes, _ := args.PayloadConverter.ConstructPayload(wsMessage)
				return websocket.TextMessage, counterBytes, nil
			}

			readAck = true
			return websocket.BinaryMessage, []byte("0"), nil

		},
	}

	go func() {
		webSocketTransceiver.Listen(conn1)
	}()

	err := webSocketTransceiver.Send([]byte("something"), outport.TopicFinalizedBlock, conn1)
	require.Nil(t, err)
	require.True(t, write)
	require.True(t, readAck)

	err = webSocketTransceiver.Close()
	require.Nil(t, err)
}

func TestSender_AddConnectionSendAndWaitForAckClose(t *testing.T) {
	args := createArgs()
	args.WithAcknowledge = true
	webSocketTransceiver, _ := NewTransceiver(args)

	conn1 := &testscommon.WebsocketConnectionStub{
		GetIDCalled: func() string {
			return "conn1"
		},
		ReadMessageCalled: func() (messageType int, payload []byte, err error) {
			time.Sleep(50 * time.Millisecond)
			return websocket.BinaryMessage, []byte("0"), nil

		},
		CloseCalled: func() error {
			return nil
		},
	}

	called := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := webSocketTransceiver.Send([]byte("something"), outport.TopicSaveAccounts, conn1)
		require.Equal(t, data.ErrExpectedAckWasNotReceivedOnClose, err)
		called = true
		wg.Done()
	}()

	time.Sleep(100 * time.Millisecond)
	go func() {
		webSocketTransceiver.Listen(conn1)
	}()

	_ = webSocketTransceiver.Close()
	wg.Wait()
	require.True(t, called)
}

func TestWsTransceiverWaitForAck(t *testing.T) {
	args := createArgs()
	args.WithAcknowledge = true
	webSocketTransceiver, _ := NewTransceiver(args)

	ch := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := webSocketTransceiver.waitForAck(ch)
		require.Equal(t, data.ErrExpectedAckWasNotReceivedOnClose, err)
		wg.Done()
	}()

	time.Sleep(time.Second)
	err := webSocketTransceiver.Close()
	require.Nil(t, err)

	wg.Wait()
}

func TestWsTransceiver_SendMessageWaitAcKTimeout(t *testing.T) {
	args := createArgs()
	args.AckTimeoutInSec = 2
	args.WithAcknowledge = true

	webSocketTransceiver, _ := NewTransceiver(args)
	defer func() {
		_ = webSocketTransceiver.Close()
	}()

	conn := &testscommon.WebsocketConnectionStub{}

	err := webSocketTransceiver.Send([]byte("message"), outport.TopicSaveBlock, conn)
	require.Equal(t, data.ErrAckTimeout, err)
}

func TestWsTransceiver_ListenReturnsTrue(t *testing.T) {
	args := createArgs()
	args.AckTimeoutInSec = 2
	args.WithAcknowledge = true

	webSocketTransceiver, _ := NewTransceiver(args)
	defer func() {
		_ = webSocketTransceiver.Close()
	}()

	count := 0
	conn := &testscommon.WebsocketConnectionStub{
		ReadMessageCalled: func() (messageType int, payload []byte, err error) {
			if count == 2 {
				return 0, nil, errors.New("test error")
			}
			count++

			time.Sleep(100 * time.Millisecond)
			return 0, nil, err
		},
	}

	closed := webSocketTransceiver.Listen(conn)
	require.True(t, closed)
}
