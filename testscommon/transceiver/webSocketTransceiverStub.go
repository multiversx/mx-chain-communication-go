package transceiver

import (
	"github.com/multiversx/mx-chain-communication-go/websocket"
)

// WebSocketTransceiverStub -
type WebSocketTransceiverStub struct {
	SendCalled              func(payload []byte, topic string, conn websocket.WSConClient) error
	CloseCalled             func() error
	SetPayloadHandlerCalled func(handler websocket.PayloadHandler) error
	ListenCalled            func(conn websocket.WSConClient) (closed bool)
}

// Send -
func (w *WebSocketTransceiverStub) Send(payload []byte, topic string, conn websocket.WSConClient) error {
	if w.SendCalled != nil {
		return w.SendCalled(payload, topic, conn)
	}
	return nil
}

// Close -
func (w *WebSocketTransceiverStub) Close() error {
	if w.CloseCalled != nil {
		return w.CloseCalled()
	}
	return nil
}

// SetPayloadHandler -
func (w *WebSocketTransceiverStub) SetPayloadHandler(handler websocket.PayloadHandler) error {
	if w.SetPayloadHandlerCalled != nil {
		return w.SetPayloadHandlerCalled(handler)
	}

	return nil
}

// Listen -
func (w *WebSocketTransceiverStub) Listen(conn websocket.WSConClient) (closed bool) {
	if w.ListenCalled != nil {
		return w.ListenCalled(conn)
	}
	return false
}
