package data

import "errors"

// ErrServerIsClosed represents the error thrown by the server's ListenAndServe() function when the server is closed
var ErrServerIsClosed = errors.New("http: Server closed")

// ErrNilMarshaller signals that a nil marshaller has been provided
var ErrNilMarshaller = errors.New("nil marshaller")

// ErrNilPayloadProcessor signals that a nil payload processor has been provided
var ErrNilPayloadProcessor = errors.New("nil payload processor provided")

// ErrNilPayloadConverter signals that a nil payload converter has been provided
var ErrNilPayloadConverter = errors.New("nil payload converter provided")

// ErrEmptyUrl signals that an empty websocket url has been provided
var ErrEmptyUrl = errors.New("empty websocket url provided")

// ErrZeroValueRetryDuration signals that a zero value for retry duration has been provided
var ErrZeroValueRetryDuration = errors.New("zero value provided for retry duration")

// ErrConnectionAlreadyOpen signal that the WebSocket connection was already open
var ErrConnectionAlreadyOpen = errors.New("connection already open")

// ErrConnectionNotOpen signal that the WebSocket connection is not open
var ErrConnectionNotOpen = errors.New("connection not open")

// ErrExpectedAckWasNotReceivedOnClose signals that the acknowledgment message was not received at close
var ErrExpectedAckWasNotReceivedOnClose = errors.New("expected ack message was not received on close")

// ErrInvalidWebSocketHostMode signals that the provided mode is invalid
var ErrInvalidWebSocketHostMode = errors.New("invalid web socket host mode")

// ErrNoClientsConnected is an error signal indicating the absence of any connected clients
var ErrNoClientsConnected = errors.New("no client connected")

// ErrNilPayloadHandlerCreator signals that a nil payload handler creator has been provided
var ErrNilPayloadHandlerCreator = errors.New("nil payload handler creator provided")
