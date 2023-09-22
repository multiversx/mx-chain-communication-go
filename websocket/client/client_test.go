package client

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/multiversx/mx-chain-communication-go/websocket"
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsWebSocketClient {
	payloadConverter, _ := websocket.NewWebSocketPayloadConverter(&testscommon.MarshallerMock{})
	return ArgsWebSocketClient{
		RetryDurationInSeconds: 1,
		BlockingAckOnError:     false,
		WithAcknowledge:        false,
		URL:                    "url",
		PayloadConverter:       payloadConverter,
		Log:                    &testscommon.LoggerMock{},
		PayloadVersion:         1,
	}
}

func TestNewWebSocketServer(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		args := createArgs()
		ws, err := NewWebSocketClient(args)
		require.NotNil(t, ws)
		require.Nil(t, err)
		require.False(t, ws.IsInterfaceNil())
	})

	t.Run("empty url, should return error", func(t *testing.T) {
		args := createArgs()
		args.URL = ""
		ws, err := NewWebSocketClient(args)
		require.Nil(t, ws)
		require.Equal(t, err, data.ErrEmptyUrl)
	})

	t.Run("nil payload converter, should return error", func(t *testing.T) {
		args := createArgs()
		args.PayloadConverter = nil
		ws, err := NewWebSocketClient(args)
		require.Nil(t, ws)
		require.Equal(t, data.ErrNilPayloadConverter, err)
	})

	t.Run("zero retry duration in seconds, should return error", func(t *testing.T) {
		args := createArgs()
		args.RetryDurationInSeconds = 0
		ws, err := NewWebSocketClient(args)
		require.Nil(t, ws)
		require.Equal(t, data.ErrZeroValueRetryDuration, err)
	})
}

func TestClient_SendAndClose(t *testing.T) {
	args := createArgs()
	args.DropMessagesIfNoConnection = false
	ws, err := NewWebSocketClient(args)
	require.Nil(t, err)

	count := uint64(0)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = ws.Send([]byte("send"), outport.TopicSaveAccounts)
		require.Equal(t, "connection not open", err.Error())
		atomic.AddUint64(&count, 1)
	}()

	_ = ws.Close()
	wg.Wait()
	require.Equal(t, uint64(1), atomic.LoadUint64(&count))
}

func TestClient_Send(t *testing.T) {
	args := createArgs()
	args.DropMessagesIfNoConnection = false
	ws, err := NewWebSocketClient(args)
	require.Nil(t, err)

	count := uint64(0)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		err = ws.Send([]byte("test"), outport.TopicFinalizedBlock)
		require.Equal(t, "connection not open", err.Error())
		atomic.AddUint64(&count, 1)
	}()

	time.Sleep(2 * time.Second)
	_ = ws.Close()
	wg.Wait()

	require.Equal(t, uint64(1), atomic.LoadUint64(&count))
}

func TestClient_DropMessageIfNoConnection(t *testing.T) {
	args := createArgs()
	args.DropMessagesIfNoConnection = true
	ws, err := NewWebSocketClient(args)
	require.Nil(t, err)

	defer func() {
		_ = ws.Close()
	}()

	err = ws.Send([]byte("test"), "test")
	require.Nil(t, err)
}
