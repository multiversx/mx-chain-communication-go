package factory

import "github.com/multiversx/mx-chain-communication-go/websocket"

// FullDuplexHost defines what a full duplex host should be able to do
type FullDuplexHost interface {
	Send(payload []byte, topic string) error
	SetPayloadHandler(handler websocket.PayloadHandler) error
	Start()
	Close() error
	IsInterfaceNil() bool
}
