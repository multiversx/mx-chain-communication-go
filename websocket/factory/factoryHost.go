package factory

import (
	"github.com/multiversx/mx-chain-communication-go/websocket"
	"github.com/multiversx/mx-chain-communication-go/websocket/client"
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/multiversx/mx-chain-communication-go/websocket/server"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

// ArgsWebSocketHost holds all the arguments needed in order to create a FullDuplexHost
type ArgsWebSocketHost struct {
	WebSocketConfig data.WebSocketConfig
	Marshaller      marshal.Marshalizer
	Log             core.Logger
}

// CreateWebSocketHost will create and start a new instance of factory.FullDuplexHost
func CreateWebSocketHost(args ArgsWebSocketHost) (FullDuplexHost, error) {
	switch args.WebSocketConfig.Mode {
	case data.ModeServer:
		return createWebSocketServer(args)
	case data.ModeClient:
		return createWebSocketClient(args)
	default:
		return nil, data.ErrInvalidWebSocketHostMode
	}
}

func createWebSocketClient(args ArgsWebSocketHost) (FullDuplexHost, error) {
	payloadConverter, err := websocket.NewWebSocketPayloadConverter(args.Marshaller)
	if err != nil {
		return nil, err
	}

	return client.NewWebSocketClient(client.ArgsWebSocketClient{
		RetryDurationInSeconds:     args.WebSocketConfig.RetryDurationInSec,
		WithAcknowledge:            args.WebSocketConfig.WithAcknowledge,
		URL:                        args.WebSocketConfig.URL,
		PayloadConverter:           payloadConverter,
		Log:                        args.Log,
		BlockingAckOnError:         args.WebSocketConfig.BlockingAckOnError,
		BlockingSendIfNoConnection: args.WebSocketConfig.BlockingSendIfNoConnection,
	})
}

func createWebSocketServer(args ArgsWebSocketHost) (FullDuplexHost, error) {
	payloadConverter, err := websocket.NewWebSocketPayloadConverter(args.Marshaller)
	if err != nil {
		return nil, err
	}

	host, err := server.NewWebSocketServer(server.ArgsWebSocketServer{
		RetryDurationInSeconds:     args.WebSocketConfig.RetryDurationInSec,
		WithAcknowledge:            args.WebSocketConfig.WithAcknowledge,
		URL:                        args.WebSocketConfig.URL,
		PayloadConverter:           payloadConverter,
		Log:                        args.Log,
		BlockingAckOnError:         args.WebSocketConfig.BlockingAckOnError,
		BlockingSendIfNoConnection: args.WebSocketConfig.BlockingSendIfNoConnection,
	})
	if err != nil {
		return nil, err
	}

	return host, nil
}
