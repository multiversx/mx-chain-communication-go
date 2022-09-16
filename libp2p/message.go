package libp2p

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-p2p/common"
	"github.com/ElrondNetwork/elrond-go-p2p/data"
	"github.com/ElrondNetwork/elrond-go-p2p/message"
	pubsub "github.com/ElrondNetwork/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-core/peer"
)

const currentTopicMessageVersion = uint32(1)

// NewMessage returns a new instance of a Message object
func NewMessage(msg *pubsub.Message, marshalizer common.Marshalizer) (*message.Message, error) {
	if check.IfNil(marshalizer) {
		return nil, common.ErrNilMarshalizer
	}
	if msg == nil {
		return nil, common.ErrNilMessage
	}
	if msg.Topic == nil {
		return nil, common.ErrNilTopic
	}

	newMsg := &message.Message{
		FromField:      msg.From,
		PayloadField:   msg.Data,
		SeqNoField:     msg.Seqno,
		TopicField:     *msg.Topic,
		SignatureField: msg.Signature,
		KeyField:       msg.Key,
	}

	topicMessage := &data.TopicMessage{}
	err := marshalizer.Unmarshal(topicMessage, msg.Data)
	if err != nil {
		return nil, fmt.Errorf("%w error: %s", common.ErrMessageUnmarshalError, err.Error())
	}

	// TODO change this area when new versions of the message will need to be implemented
	if topicMessage.Version != currentTopicMessageVersion {
		return nil, fmt.Errorf("%w, supported %d, got %d",
			common.ErrUnsupportedMessageVersion, currentTopicMessageVersion, topicMessage.Version)
	}

	if len(topicMessage.SignatureOnPid)+len(topicMessage.Pk) > 0 {
		return nil, fmt.Errorf("%w for topicMessage.SignatureOnPid and topicMessage.Pk",
			common.ErrUnsupportedFields)
	}

	newMsg.DataField = topicMessage.Payload
	newMsg.TimestampField = topicMessage.Timestamp

	id, err := peer.IDFromBytes(newMsg.From())
	if err != nil {
		return nil, err
	}

	newMsg.PeerField = core.PeerID(id)
	return newMsg, nil
}
