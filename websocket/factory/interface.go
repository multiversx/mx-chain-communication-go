package factory

import webSocket "github.com/multiversx/mx-chain-communication-go/websocket"

// FullDuplexHost defines what a full duplex host should be able to do
type FullDuplexHost interface {
	Send(payload []byte, topic string) error
	SetPayloadHandlerCreator(creator webSocket.PayloadHandlerCreator) error
	Close() error
	IsInterfaceNil() bool
}
