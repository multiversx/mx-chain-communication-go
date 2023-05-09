package integrationTests

import (
	"github.com/multiversx/mx-chain-communication-go/websocket"
	"github.com/multiversx/mx-chain-communication-go/websocket/client"
	"github.com/multiversx/mx-chain-communication-go/websocket/server"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal/factory"
)

const retryDurationInSeconds = 1

var (
	marshaller, _       = factory.NewMarshalizer("gogo protobuf")
	payloadConverter, _ = websocket.NewWebSocketPayloadConverter(marshaller)
)

func createClient(url string, log core.Logger) (websocket.HostWebSocket, error) {
	return client.NewWebSocketClient(client.ArgsWebSocketClient{
		RetryDurationInSeconds: retryDurationInSeconds,
		WithAcknowledge:        true,
		URL:                    url,
		PayloadConverter:       payloadConverter,
		Log:                    log,
	})
}

func createServer(url string, log core.Logger) (websocket.HostWebSocket, error) {
	return server.NewWebSocketServer(server.ArgsWebSocketServer{
		RetryDurationInSeconds: retryDurationInSeconds,
		WithAcknowledge:        true,
		URL:                    url,
		PayloadConverter:       payloadConverter,
		Log:                    log,
	})
}
