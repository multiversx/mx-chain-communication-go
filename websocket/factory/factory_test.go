package factory

import (
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsWebSocketDriverFactory {
	return ArgsWebSocketDriverFactory{
		WebSocketConfig: data.WebSocketConfig{
			URL:                "localhost:1234",
			WithAcknowledge:    false,
			IsServer:           false,
			RetryDurationInSec: 1,
			BlockingAckOnError: false,
		},
		Marshaller: &testscommon.MarshalizerMock{},
		Log:        &testscommon.LoggerMock{},
	}
}

func TestNewWebSocketDriver(t *testing.T) {
	t.Parallel()

	args := createArgs()
	driver, err := NewWebSocketDriver(args)
	require.Nil(t, err)
	require.NotNil(t, driver)
	require.Equal(t, "*driver.webSocketDriver", fmt.Sprintf("%T", driver))

	err = driver.Close()
	require.Equal(t, errors.New("connection not open"), err)
}

func TestCreateClient(t *testing.T) {
	t.Parallel()

	args := createArgs()
	webSocketsClient, err := createWebSocketClient(args)
	require.Nil(t, err)
	require.Equal(t, "*client.client", fmt.Sprintf("%T", webSocketsClient))
}

func TestCreateServer(t *testing.T) {
	t.Parallel()

	args := createArgs()
	webSocketsClient, err := createWebSocketServer(args)
	require.Nil(t, err)
	require.Equal(t, "*server.server", fmt.Sprintf("%T", webSocketsClient))
}
