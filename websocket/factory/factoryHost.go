package factory

import (
	"github.com/multiversx/mx-chain-communication-go/websocket"
	"github.com/multiversx/mx-chain-communication-go/websocket/client"
	outportData "github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/multiversx/mx-chain-communication-go/websocket/server"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

// ArgsWebSocketHost holds all the arguments needed in order to create a FullDuplexHost
type ArgsWebSocketHost struct {
	WebSocketConfig outportData.WebSocketConfig
	Marshaller      marshal.Marshalizer
	Log             core.Logger
}

// CreateWebSocketHost will create and start a new instance of factory.FullDuplexHost
func CreateWebSocketHost(args ArgsWebSocketHost) (FullDuplexHost, error) {
	var host FullDuplexHost
	var err error
	if args.WebSocketConfig.IsServer {
		host, err = createWebSocketServer(args)
	} else {
		host, err = createWebSocketClient(args)
	}

	if err != nil {
		return nil, err
	}

	host.Start()

	return host, nil
}

func createWebSocketClient(args ArgsWebSocketHost) (FullDuplexHost, error) {
	payloadConverter, err := websocket.NewWebSocketPayloadConverter(args.Marshaller)
	if err != nil {
		return nil, err
	}

	return client.NewWebSocketClient(client.ArgsWebSocketClient{
		RetryDurationInSeconds: args.WebSocketConfig.RetryDurationInSec,
		WithAcknowledge:        args.WebSocketConfig.WithAcknowledge,
		URL:                    args.WebSocketConfig.URL,
		PayloadConverter:       payloadConverter,
		Log:                    args.Log,
		BlockingAckOnError:     args.WebSocketConfig.BlockingAckOnError,
	})
}

func createWebSocketServer(args ArgsWebSocketHost) (FullDuplexHost, error) {
	payloadConverter, err := websocket.NewWebSocketPayloadConverter(args.Marshaller)
	if err != nil {
		return nil, err
	}

	host, err := server.NewWebSocketServer(server.ArgsWebSocketServer{
		RetryDurationInSeconds: args.WebSocketConfig.RetryDurationInSec,
		WithAcknowledge:        args.WebSocketConfig.WithAcknowledge,
		URL:                    args.WebSocketConfig.URL,
		PayloadConverter:       payloadConverter,
		Log:                    args.Log,
		BlockingAckOnError:     args.WebSocketConfig.BlockingAckOnError,
	})
	if err != nil {
		return nil, err
	}

	return host, nil
}
