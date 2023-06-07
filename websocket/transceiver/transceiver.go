package transceiver

import (
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	webSocket "github.com/multiversx/mx-chain-communication-go/websocket"
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/closing"
)

// ArgsTransceiver holds the arguments that are needed for a transceiver
type ArgsTransceiver struct {
	PayloadConverter   webSocket.PayloadConverter
	Log                core.Logger
	RetryDurationInSec int
	AckTimeoutInSec    int
	BlockingAckOnError bool
	WithAcknowledge    bool
}

type wsTransceiver struct {
	payloadParser      webSocket.PayloadConverter
	payloadHandler     webSocket.PayloadHandler
	mutPayloadHandler  sync.RWMutex
	log                core.Logger
	safeCloser         core.SafeCloser
	retryDuration      time.Duration
	ackTimeout         time.Duration
	mapAck             map[uint64]chan struct{}
	mutMapAck          sync.Mutex
	counter            uint64
	blockingAckOnError bool
	withAcknowledge    bool
}

// NewTransceiver will create a new instance of transceiver
func NewTransceiver(args ArgsTransceiver) (*wsTransceiver, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &wsTransceiver{
		log:                args.Log,
		retryDuration:      time.Duration(args.RetryDurationInSec) * time.Second,
		ackTimeout:         time.Duration(args.AckTimeoutInSec) * time.Second,
		blockingAckOnError: args.BlockingAckOnError,
		safeCloser:         closing.NewSafeChanCloser(),
		payloadHandler:     webSocket.NewNilPayloadHandler(),
		payloadParser:      args.PayloadConverter,
		withAcknowledge:    args.WithAcknowledge,
		mapAck:             make(map[uint64]chan struct{}),
	}, nil
}

func checkArgs(args ArgsTransceiver) error {
	if check.IfNil(args.Log) {
		return core.ErrNilLogger
	}
	if check.IfNil(args.PayloadConverter) {
		return data.ErrNilPayloadConverter
	}
	if args.RetryDurationInSec == 0 {
		return data.ErrZeroValueRetryDuration
	}
	if args.WithAcknowledge == true && args.AckTimeoutInSec == 0 {
		return data.ErrZeroValueAckTimeout
	}
	return nil
}

// SetPayloadHandler will set the payload handler
func (wt *wsTransceiver) SetPayloadHandler(handler webSocket.PayloadHandler) error {
	if check.IfNil(handler) {
		return data.ErrNilPayloadProcessor
	}

	wt.mutPayloadHandler.Lock()
	defer wt.mutPayloadHandler.Unlock()

	wt.payloadHandler = handler
	return nil
}

// Listen will listen for messages from the provided connection
func (wt *wsTransceiver) Listen(connection webSocket.WSConClient) (closed bool) {
	timer := time.NewTimer(wt.retryDuration)
	defer timer.Stop()

	for {
		_, message, err := connection.ReadMessage()
		if err == nil {
			wt.verifyPayloadAndSendAckIfNeeded(connection, message)
			continue
		}

		_, isConnectionClosed := err.(*websocket.CloseError)
		if !strings.Contains(err.Error(), data.ClosedConnectionMessage) && !isConnectionClosed {
			wt.log.Warn("wt.Listen()-> connection problem", "error", err.Error())
		}
		if isConnectionClosed {
			wt.log.Info("received connection close")
			return true
		}

		timer.Reset(wt.retryDuration)

		select {
		case <-wt.safeCloser.ChanClose():
			return
		case <-timer.C:
		}
	}
}

func (wt *wsTransceiver) verifyPayloadAndSendAckIfNeeded(connection webSocket.WSConClient, payload []byte) {
	if len(payload) == 0 {
		wt.log.Debug("wt.verifyPayloadAndSendAckIfNeeded(): empty payload")
		return
	}

	wsMessage, err := wt.payloadParser.ExtractWsMessage(payload)
	if err != nil {
		wt.log.Warn("wt.verifyPayloadAndSendAckIfNeeded: cannot extract payload data", "error", err.Error())
		return
	}

	if wsMessage.Type == data.AckMessage {
		wt.handleAckMessage(wsMessage.Counter)
		return
	}

	if wsMessage.Type != data.PayloadMessage {
		wt.log.Debug("received an unknown message type", "message type received", wsMessage.Type)
		return
	}

	err = wt.payloadHandler.ProcessPayload(wsMessage.Payload, wsMessage.Topic)
	if err != nil && wt.blockingAckOnError {
		wt.log.Warn("wt.payloadHandler.ProcessPayload: cannot handle payload", "error", err)
		return
	}

	wt.sendAckIfNeeded(connection, wsMessage)
}

func (wt *wsTransceiver) handleAckMessage(counter uint64) {
	wt.mutMapAck.Lock()
	defer wt.mutMapAck.Unlock()

	ch, found := wt.mapAck[counter]
	if !found {
		wt.log.Warn("wsTransceiver.handleAckMessage invalid counter received", "received", counter)
		return
	}

	close(ch)
	delete(wt.mapAck, counter)
}

func (wt *wsTransceiver) sendAckIfNeeded(connection webSocket.WSConClient, wsMessage *data.WsMessage) {
	if !wsMessage.WithAcknowledge {
		return
	}

	timer := time.NewTimer(wt.retryDuration)
	defer timer.Stop()

	ackWsMessage := &data.WsMessage{
		Counter: wsMessage.Counter,
		Type:    data.AckMessage,
	}
	wsMessageBytes, errConstruct := wt.payloadParser.ConstructPayload(ackWsMessage)
	if errConstruct != nil {
		wt.log.Warn("sendAckIfNeeded.ConstructPayload: cannot prepare message", "error", errConstruct)
		return
	}

	for {
		timer.Reset(wt.retryDuration)

		err := connection.WriteMessage(websocket.BinaryMessage, wsMessageBytes)
		if err == nil {
			return
		}

		if !strings.Contains(err.Error(), data.ErrConnectionNotOpen.Error()) {
			wt.log.Error("could not write acknowledge message", "error", err.Error(), "retrying in", wt.retryDuration)
		}

		wt.log.Debug("wt.sendAckIfNeeded(): cannot write ack", "error", err)

		select {
		case <-timer.C:
		case <-wt.safeCloser.ChanClose():
			return
		}
	}
}

// Send will prepare and send the provided WsSendArgs
func (wt *wsTransceiver) Send(payload []byte, topic string, connection webSocket.WSConClient) error {
	ch, localCounter := wt.prepareChanAndCounter()
	wsMessage := &data.WsMessage{
		WithAcknowledge: wt.withAcknowledge,
		Counter:         localCounter,
		Type:            data.PayloadMessage,
		Payload:         payload,
		Topic:           topic,
	}
	newPayload, err := wt.payloadParser.ConstructPayload(wsMessage)
	if err != nil {
		return err
	}

	return wt.sendPayload(newPayload, connection, ch)
}

func (wt *wsTransceiver) prepareChanAndCounter() (chan struct{}, uint64) {
	wt.mutMapAck.Lock()
	wt.counter++
	localCounter := wt.counter

	ch := make(chan struct{})
	if wt.withAcknowledge {
		wt.mapAck[localCounter] = ch
	}
	wt.mutMapAck.Unlock()

	return ch, localCounter
}

func (wt *wsTransceiver) sendPayload(payload []byte, connection webSocket.WSConClient, ch chan struct{}) error {
	errSend := connection.WriteMessage(websocket.BinaryMessage, payload)
	if errSend != nil {
		return errSend
	}

	if !wt.withAcknowledge {
		return nil
	}

	return wt.waitForAck(ch)
}

func (wt *wsTransceiver) waitForAck(ch chan struct{}) error {
	timer := time.NewTimer(wt.ackTimeout)
	defer timer.Stop()

	select {
	case <-ch:
		return nil
	case <-timer.C:
		return data.ErrAckTimeout
	case <-wt.safeCloser.ChanClose():
		return data.ErrExpectedAckWasNotReceivedOnClose
	}
}

// Close will close the underlying ws connection
func (wt *wsTransceiver) Close() error {
	defer wt.safeCloser.Close()

	err := wt.payloadHandler.Close()
	if err != nil {
		wt.log.Debug("cannot close the payload handler", "error", err)
	}

	return err
}
