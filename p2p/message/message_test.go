package message_test

import (
	"testing"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/message"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestMessage_AllFieldsShouldWork(t *testing.T) {
	t.Parallel()

	from := []byte("from")
	data := []byte("data")
	seqNo := []byte("seq no")
	topic := "topic"
	sig := []byte("sig")
	key := []byte("key")
	peer := core.PeerID("peer")
	msgType := p2p.Direct

	msg := &message.Message{
		FromField:            from,
		DataField:            data,
		SeqNoField:           seqNo,
		TopicField:           topic,
		SignatureField:       sig,
		KeyField:             key,
		PeerField:            peer,
		BroadcastMethodField: msgType,
	}

	assert.False(t, check.IfNil(msg))
	assert.Equal(t, from, msg.From())
	assert.Equal(t, data, msg.Data())
	assert.Equal(t, seqNo, msg.SeqNo())
	assert.Equal(t, topic, msg.Topic())
	assert.Equal(t, sig, msg.Signature())
	assert.Equal(t, key, msg.Key())
	assert.Equal(t, peer, msg.Peer())
	assert.Equal(t, msgType, msg.BroadcastMethod())
}
