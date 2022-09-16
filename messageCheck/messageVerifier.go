package messagecheck

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go-p2p/common"
	"github.com/ElrondNetwork/elrond-go-p2p/message"
	pubsub "github.com/ElrondNetwork/go-libp2p-pubsub"
	pubsub_pb "github.com/ElrondNetwork/go-libp2p-pubsub/pb"
)

type messageVerifier struct {
	marshaller marshal.Marshalizer
	p2pSigner  p2pSigner
}

// ArgsMessageVerifier defines the arguments needed to create a messageVerifier
type ArgsMessageVerifier struct {
	Marshaller marshal.Marshalizer
	P2PSigner  p2pSigner
}

// NewMessageVerifier will create a new instance of messageVerifier
func NewMessageVerifier(args ArgsMessageVerifier) (*messageVerifier, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &messageVerifier{
		marshaller: args.Marshaller,
		p2pSigner:  args.P2PSigner,
	}, nil
}

func checkArgs(args ArgsMessageVerifier) error {
	if check.IfNil(args.Marshaller) {
		return common.ErrNilMarshalizer
	}
	if args.P2PSigner == nil {
		return common.ErrNilP2PSigner
	}

	return nil
}

// Verify will check the signature of a p2p message
func (m *messageVerifier) Verify(msg common.MessageP2P) error {
	if check.IfNil(msg) {
		return common.ErrNilMessage
	}

	payload, err := preparePubSubMessagePayload(msg)
	if err != nil {
		return err
	}

	err = m.p2pSigner.Verify(payload, msg.Peer(), msg.Signature())
	if err != nil {
		return err
	}

	return nil
}

func preparePubSubMessagePayload(msg common.MessageP2P) ([]byte, error) {
	pubsubMsg, err := convertP2PMessagetoPubSubMessage(msg)
	if err != nil {
		return nil, err
	}

	xm := *pubsubMsg
	xm.Signature = nil
	xm.Key = nil

	pubsubMsgBytes, err := xm.Marshal()
	if err != nil {
		return nil, err
	}

	pubsubMsgBytes = withSignPrefix(pubsubMsgBytes)

	return pubsubMsgBytes, nil
}

func withSignPrefix(bytes []byte) []byte {
	return append([]byte(pubsub.SignPrefix), bytes...)
}

func convertP2PMessagetoPubSubMessage(msg common.MessageP2P) (*pubsub_pb.Message, error) {
	if check.IfNil(msg) {
		return nil, common.ErrNilMessage
	}

	topic := msg.Topic()

	newMsg := &pubsub_pb.Message{
		From:      msg.From(),
		Data:      msg.Payload(),
		Seqno:     msg.SeqNo(),
		Topic:     &topic,
		Signature: msg.Signature(),
		Key:       msg.Key(),
	}

	return newMsg, nil
}

func convertPubSubMessagestoP2PMessage(msg *pubsub_pb.Message) (common.MessageP2P, error) {
	if msg == nil {
		return nil, common.ErrNilMessage
	}

	newMsg := &message.Message{
		FromField:      msg.From,
		PayloadField:   msg.Data,
		SeqNoField:     msg.Seqno,
		TopicField:     *msg.Topic,
		SignatureField: msg.Signature,
		KeyField:       msg.Key,
	}

	return newMsg, nil
}

// Serialize will serialize a list of p2p messages
func (m *messageVerifier) Serialize(messages []common.MessageP2P) ([]byte, error) {
	pubsubMessages := make([][]byte, 0, len(messages))
	for _, msg := range messages {
		pubsubMsg, err := convertP2PMessagetoPubSubMessage(msg)
		if err != nil {
			continue
		}

		pubsubMsgBytes, err := pubsubMsg.Marshal()
		if err != nil {
			return nil, err
		}

		pubsubMessages = append(pubsubMessages, pubsubMsgBytes)
	}

	messagesBytes, err := m.marshaller.Marshal(pubsubMessages)
	if err != nil {
		return nil, err
	}

	return messagesBytes, nil
}

// Deserialize will deserialize into a list of p2p messages
func (m *messageVerifier) Deserialize(messagesBytes []byte) ([]common.MessageP2P, error) {
	var pubsubMessagesBytes [][]byte
	err := m.marshaller.Unmarshal(&pubsubMessagesBytes, messagesBytes)
	if err != nil {
		return nil, err
	}

	p2pMessages := make([]common.MessageP2P, 0)
	for _, pubsubMessageBytes := range pubsubMessagesBytes {
		var pubsubMsg pubsub_pb.Message
		err = pubsubMsg.Unmarshal(pubsubMessageBytes)
		if err != nil {
			return nil, err
		}

		p2pMsg, err := convertPubSubMessagestoP2PMessage(&pubsubMsg)
		if err != nil {
			continue
		}

		p2pMessages = append(p2pMessages, p2pMsg)
	}

	return p2pMessages, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *messageVerifier) IsInterfaceNil() bool {
	return m == nil
}
