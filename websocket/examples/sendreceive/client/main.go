package main

import (
	"sync"

	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/multiversx/mx-chain-communication-go/testscommon/creator"
	"github.com/multiversx/mx-chain-communication-go/websocket"
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	factoryHost "github.com/multiversx/mx-chain-communication-go/websocket/factory"
	"github.com/multiversx/mx-chain-core-go/marshal/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var (
	marshaller, _ = factory.NewMarshalizer("json")
	log           = logger.GetOrCreate("client")
	url           = "localhost:12345"
)

func main() {
	args := factoryHost.ArgsWebSocketHost{
		WebSocketConfig: data.WebSocketConfig{
			URL:                        url,
			Mode:                       data.ModeClient,
			RetryDurationInSec:         1,
			WithAcknowledge:            true,
			BlockingAckOnError:         false,
			DropMessagesIfNoConnection: false,
		},
		Marshaller: marshaller,
		Log:        log,
	}

	wsClient, err := factoryHost.CreateWebSocketHost(args)
	if err != nil {
		log.Error("cannot create WebSocket client", "error", err)
		return
	}

	defer func() {
		_ = wsClient.Close()
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)

	_ = wsClient.SetPayloadHandlerCreator(&creator.PayloadHandlerCreatorStub{
		CreateCalled: func() (websocket.PayloadHandler, error) {
			return &testscommon.PayloadHandlerStub{
				ProcessPayloadCalled: func(payload []byte, topic string) error {
					log.Info("received", "topic", topic, "payload", string(payload))
					wg.Done()
					return nil
				},
			}, nil
		},
	})

	wg.Wait()
}
