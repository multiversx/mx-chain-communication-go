package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/multiversx/mx-chain-communication-go/testscommon"
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
			AcknowledgeTimeoutInSec:    10,
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

	err = wsClient.SetPayloadHandler(&testscommon.PayloadHandlerStub{
		ProcessPayloadCalled: func(payload []byte, topic string, version string) error {
			log.Info("received", "topic", topic, "payload", string(payload), "version", version)
			return nil
		},
	})
	log.LogIfError(err)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-interrupt
	err = wsClient.Close()
	log.LogIfError(err)
}
