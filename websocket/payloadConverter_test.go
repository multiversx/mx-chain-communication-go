package websocket

import (
	"testing"

	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/require"
)

func TestNewWebSocketPayloadConverter(t *testing.T) {
	t.Parallel()

	payloadConverter, err := NewWebSocketPayloadConverter(nil)
	require.Nil(t, payloadConverter)
	require.Equal(t, data.ErrNilMarshaller, err)

	payloadConverter, _ = NewWebSocketPayloadConverter(&testscommon.MarshallerMock{})
	require.NotNil(t, payloadConverter)
	require.False(t, payloadConverter.IsInterfaceNil())
}

func TestWebSocketPayloadConverter_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	addrGroup, _ := NewWebSocketPayloadConverter(nil)
	require.True(t, addrGroup.IsInterfaceNil())

	addrGroup, _ = NewWebSocketPayloadConverter(&testscommon.MarshallerMock{})
	require.False(t, addrGroup.IsInterfaceNil())
}

func TestWebSocketsPayloadConverter_ConstructPayload(t *testing.T) {
	t.Parallel()

	payloadConverter, _ := NewWebSocketPayloadConverter(&testscommon.MarshallerMock{})

	wsMessage := &data.WsMessage{
		WithAcknowledge: true,
		Payload:         []byte("test"),
		Topic:           outport.TopicSaveAccounts,
		Counter:         10,
		Type:            data.PayloadMessage,
	}

	payload, err := payloadConverter.ConstructPayload(wsMessage)
	require.Nil(t, err)

	newWsMessage, err := payloadConverter.ExtractWsMessage(payload)
	require.Nil(t, err)
	require.Equal(t, wsMessage, newWsMessage)
}
