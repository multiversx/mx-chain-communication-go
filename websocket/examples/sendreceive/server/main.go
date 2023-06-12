package main

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	factoryHost "github.com/multiversx/mx-chain-communication-go/websocket/factory"
	"github.com/multiversx/mx-chain-core-go/marshal/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var (
	marshaller, _ = factory.NewMarshalizer("json")
	log           = logger.GetOrCreate("server")
	url           = "localhost:12345"
)

func main() {
	args := factoryHost.ArgsWebSocketHost{
		WebSocketConfig: data.WebSocketConfig{
			URL:                        url,
			Mode:                       data.ModeServer,
			RetryDurationInSec:         1,
			WithAcknowledge:            true,
			BlockingAckOnError:         false,
			DropMessagesIfNoConnection: false,
			AcknowledgeTimeoutInSec:    10,
		},
		Marshaller: marshaller,
		Log:        log,
	}

	wsServer, err := factoryHost.CreateWebSocketHost(args)
	if err != nil {
		log.Error("cannot create WebSocket server", "error", err)
		return
	}

	count := 0
	for {
		message := fmt.Sprintf("messagae %d", count)
		err = wsServer.Send([]byte(message), "test")
		if err == nil {
			count++
			log.Info("message sent to clients")
		}
		time.Sleep(100 * time.Millisecond)
	}
}
