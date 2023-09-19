package server

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/multiversx/mx-chain-communication-go/websocket"
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsWebSocketServer {
	payloadConverter, _ := websocket.NewWebSocketPayloadConverter(&testscommon.MarshallerMock{})
	return ArgsWebSocketServer{
		RetryDurationInSeconds: 1,
		BlockingAckOnError:     false,
		WithAcknowledge:        false,
		URL:                    "url",
		PayloadConverter:       payloadConverter,
		Log:                    &testscommon.LoggerMock{},
		PayloadVersion:         "1.0",
	}
}

func TestNewWebSocketsServer(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		args := createArgs()
		ws, err := NewWebSocketServer(args)
		require.NotNil(t, ws)
		require.Nil(t, err)
		require.False(t, ws.IsInterfaceNil())
	})

	t.Run("empty url, should return error", func(t *testing.T) {
		args := createArgs()
		args.URL = ""
		ws, err := NewWebSocketServer(args)
		require.Nil(t, ws)
		require.Equal(t, data.ErrEmptyUrl, err)
	})

	t.Run("nil payload converter, should return error", func(t *testing.T) {
		args := createArgs()
		args.PayloadConverter = nil
		ws, err := NewWebSocketServer(args)
		require.Nil(t, ws)
		require.Equal(t, data.ErrNilPayloadConverter, err)
	})

	t.Run("zero retry duration in seconds, should return error", func(t *testing.T) {
		args := createArgs()
		args.RetryDurationInSeconds = 0
		ws, err := NewWebSocketServer(args)
		require.Nil(t, ws)
		require.Equal(t, data.ErrZeroValueRetryDuration, err)
	})
}

func TestServer_ListenAndClose(t *testing.T) {
	count := uint64(0)
	wg := sync.WaitGroup{}
	wg.Add(1)

	args := createArgs()
	args.URL = "localhost:9211"
	args.Log = &testscommon.LoggerStub{
		InfoCalled: func(message string, args ...interface{}) {
			if message == "server was closed" {
				wg.Done()
				atomic.AddUint64(&count, 1)
			}
		},
	}
	wsServer, _ := NewWebSocketServer(args)

	_ = wsServer.Close()
	wg.Wait()
	require.Equal(t, uint64(1), atomic.LoadUint64(&count))
}

func TestServer_ListenAndRegisterPayloadHandlerAndClose(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	args := createArgs()
	args.Log = &testscommon.LoggerStub{
		InfoCalled: func(message string, args ...interface{}) {
			if message == "server was closed" {
				wg.Done()
			}
		},
	}

	args.URL = "localhost:9211"
	wsServer, _ := NewWebSocketServer(args)

	_ = wsServer.SetPayloadHandler(&testscommon.PayloadHandlerStub{})
	wsServer.connectionHandler(&testscommon.WebsocketConnectionStub{
		ReadMessageCalled: func() (messageType int, payload []byte, err error) {
			return 0, nil, errors.New("local error")
		},
	})

	_ = wsServer.Close()
	wg.Wait()
}

func TestServer_DropMessageInCaseOfNoConnection(t *testing.T) {
	args := createArgs()
	args.DropMessagesIfNoConnection = true
	args.URL = "localhost:9211"
	wsServer, _ := NewWebSocketServer(args)

	defer func() {
		_ = wsServer.Close()
	}()

	err := wsServer.Send([]byte("test"), "test")
	require.Nil(t, err)
}

func TestServer_SendReturnsErrorIfNoConnection(t *testing.T) {
	args := createArgs()
	args.DropMessagesIfNoConnection = false
	args.URL = "localhost:9211"
	wsServer, _ := NewWebSocketServer(args)

	defer func() {
		_ = wsServer.Close()
	}()

	err := wsServer.Send([]byte("test"), "test")
	require.Equal(t, data.ErrNoClientsConnected, err)
}
