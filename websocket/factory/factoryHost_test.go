package factory

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsWebSocketHost {
	return ArgsWebSocketHost{
		WebSocketConfig: data.WebSocketConfig{
			URL:                "localhost:1234",
			WithAcknowledge:    false,
			Mode:               data.ModeClient,
			RetryDurationInSec: 1,
			BlockingAckOnError: false,
		},
		Marshaller: &testscommon.MarshallerMock{},
		Log:        &testscommon.LoggerMock{},
	}
}

func TestCreateClient(t *testing.T) {
	t.Parallel()

	args := createArgs()
	webSocketsClient, err := CreateWebSocketHost(args)
	require.Nil(t, err)
	require.Equal(t, "*client.client", fmt.Sprintf("%T", webSocketsClient))
}

func TestCreateServer(t *testing.T) {
	t.Parallel()

	args := createArgs()
	args.WebSocketConfig.Mode = data.ModeServer
	webSocketsClient, err := CreateWebSocketHost(args)
	require.Nil(t, err)
	require.Equal(t, "*server.server", fmt.Sprintf("%T", webSocketsClient))
}
