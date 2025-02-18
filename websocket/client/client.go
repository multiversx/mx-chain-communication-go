package client

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/multiversx/mx-chain-communication-go/websocket"
	"github.com/multiversx/mx-chain-communication-go/websocket/connection"
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/multiversx/mx-chain-communication-go/websocket/transceiver"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/closing"
)

// ArgsWebSocketClient holds the arguments needed for creating a client
type ArgsWebSocketClient struct {
	RetryDurationInSeconds     int
	AckTimeoutInSeconds        int
	WithAcknowledge            bool
	BlockingAckOnError         bool
	DropMessagesIfNoConnection bool
	URL                        string
	PayloadConverter           websocket.PayloadConverter
	Log                        core.Logger
	PayloadVersion             uint32
}

type client struct {
	url                        string
	retryDuration              time.Duration
	safeCloser                 core.SafeCloser
	log                        core.Logger
	wsConn                     websocket.WSConClient
	transceiver                Transceiver
	dropMessagesIfNoConnection bool
}

// NewWebSocketClient will create a new instance of WebSocket client
func NewWebSocketClient(args ArgsWebSocketClient) (*client, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	argsTransceiver := transceiver.ArgsTransceiver{
		PayloadConverter:   args.PayloadConverter,
		Log:                args.Log,
		RetryDurationInSec: args.RetryDurationInSeconds,
		AckTimeoutInSec:    args.AckTimeoutInSeconds,
		BlockingAckOnError: args.BlockingAckOnError,
		WithAcknowledge:    args.WithAcknowledge,
		PayloadVersion:     args.PayloadVersion,
	}
	wsTransceiver, err := transceiver.NewTransceiver(argsTransceiver)
	if err != nil {
		return nil, err
	}

	wsUrl, err := url.Parse(args.URL)
	if err != nil || (wsUrl.Scheme != "ws" && wsUrl.Scheme != "wss") {
		return nil, fmt.Errorf("invalid WebSocket URL: %v", err)
	}

	wsUrl.Path = data.WSRoute

	wsClient := &client{
		url:                        wsUrl.String(),
		wsConn:                     connection.NewWSConnClient(),
		retryDuration:              time.Duration(args.RetryDurationInSeconds) * time.Second,
		safeCloser:                 closing.NewSafeChanCloser(),
		transceiver:                wsTransceiver,
		log:                        args.Log,
		dropMessagesIfNoConnection: args.DropMessagesIfNoConnection,
	}

	wsClient.start()

	return wsClient, nil
}

func checkArgs(args ArgsWebSocketClient) error {
	if check.IfNil(args.Log) {
		return core.ErrNilLogger
	}
	if check.IfNil(args.PayloadConverter) {
		return data.ErrNilPayloadConverter
	}
	if args.URL == "" {
		return data.ErrEmptyUrl
	}
	if args.RetryDurationInSeconds == 0 {
		return data.ErrZeroValueRetryDuration
	}
	return nil
}

func (c *client) start() {
	go func() {
		timer := time.NewTimer(c.retryDuration)
		defer timer.Stop()
		for {
			err := c.wsConn.OpenConnection(c.url)
			if err != nil && !errors.Is(err, data.ErrConnectionAlreadyOpen) {
				c.log.Warn(fmt.Sprintf("c.openConnection(), retrying in %v...", c.retryDuration), "error", err)
			}

			timer.Reset(c.retryDuration)

			select {
			case <-timer.C:
			case <-c.safeCloser.ChanClose():
				return
			}
		}
	}()

	go func() {
		timer := time.NewTimer(c.retryDuration)
		defer timer.Stop()
		for {
			closed := c.transceiver.Listen(c.wsConn)
			if closed {
				err := c.wsConn.Close()
				c.log.Debug("try to close the connection", "close error", err)
			}

			timer.Reset(c.retryDuration)

			select {
			case <-c.safeCloser.ChanClose():
				return
			case <-timer.C:
			}
		}
	}()
}

// Send will send the provided payload from args
func (c *client) Send(payload []byte, topic string) error {
	dropMessage := c.dropMessagesIfNoConnection && !c.wsConn.IsOpen()
	if dropMessage {
		return nil
	}

	return c.transceiver.Send(payload, topic, c.wsConn)
}

// SetPayloadHandler set the payload handler
func (c *client) SetPayloadHandler(handler websocket.PayloadHandler) error {
	return c.transceiver.SetPayloadHandler(handler)
}

// Close will close the component
func (c *client) Close() error {
	defer c.safeCloser.Close()

	var lastErr error

	c.log.Info("closing client...")
	err := c.transceiver.Close()
	if err != nil {
		c.log.Warn("client.Close() transceiver", "error", err)
		lastErr = err
	}

	err = c.wsConn.Close()
	if err != nil {
		c.log.Warn("client.Close() cannot close connection", "error", err)
		lastErr = err
	}

	return lastErr
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *client) IsInterfaceNil() bool {
	return c == nil
}
