package websocket

import (
	"context"
	"io"

	"github.com/multiversx/mx-chain-communication-go/websocket/data"
)

// PayloadHandler defines what a payload handler should be able to do
type PayloadHandler interface {
	ProcessPayload(payload []byte, topic string, version uint32) error
	Close() error
	IsInterfaceNil() bool
}

// PayloadConverter defines what a websocket payload converter should do
type PayloadConverter interface {
	ExtractWsMessage(payload []byte) (*data.WsMessage, error)
	ConstructPayload(wsMessage *data.WsMessage) ([]byte, error)
	IsInterfaceNil() bool
}

// WSConClient defines what a web-sockets connection client should be able to do
type WSConClient interface {
	io.Closer
	OpenConnection(url string) error
	IsOpen() bool
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (int, []byte, error)
	GetID() string
	IsInterfaceNil() bool
}

// HttpServerHandler defines the minimum behaviour of a http server
type HttpServerHandler interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}
