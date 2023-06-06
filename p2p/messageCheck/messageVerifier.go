package messagecheck

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsubPb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const loggerName = "p2p/messagecheck"

var log = logger.GetOrCreate(loggerName)

type messageVerifier struct {
	marshaller marshal.Marshalizer
	p2pSigner  p2pSigner
}

// ArgsMessageVerifier defines the arguments needed to create a messageVerifier
type ArgsMessageVerifier struct {
	Marshaller   marshal.Marshalizer
	P2PSigner    p2pSigner
	LoggerPrefix string
}

// NewMessageVerifier will create a new instance of messageVerifier
func NewMessageVerifier(args ArgsMessageVerifier) (*messageVerifier, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	if len(args.LoggerPrefix) > 0 {
		log = logger.GetOrCreate(args.LoggerPrefix + loggerName)
	}

	return &messageVerifier{
		marshaller: args.Marshaller,
		p2pSigner:  args.P2PSigner,
	}, nil
}

func checkArgs(args ArgsMessageVerifier) error {
	if check.IfNil(args.Marshaller) {
		return p2p.ErrNilMarshaller
	}
	if args.P2PSigner == nil {
		return p2p.ErrNilP2PSigner
	}

	return nil
}

// Verify will check the signature of a p2p message
func (m *messageVerifier) Verify(msg p2p.MessageP2P) error {
	if check.IfNil(msg) {
		return p2p.ErrNilMessage
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

func preparePubSubMessagePayload(msg p2p.MessageP2P) ([]byte, error) {
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

func convertP2PMessagetoPubSubMessage(msg p2p.MessageP2P) (*pubsubPb.Message, error) {
	if check.IfNil(msg) {
		return nil, p2p.ErrNilMessage
	}

	topic := msg.Topic()

	newMsg := &pubsubPb.Message{
		From:      msg.From(),
		Data:      msg.Payload(),
		Seqno:     msg.SeqNo(),
		Topic:     &topic,
		Signature: msg.Signature(),
		Key:       msg.Key(),
	}

	return newMsg, nil
}

func convertPubSubMessagestoP2PMessage(msg *pubsubPb.Message, marshaller marshal.Marshalizer) (p2p.MessageP2P, error) {
	if msg == nil {
		return nil, p2p.ErrNilMessage
	}

	pubsubMsg := &pubsub.Message{
		Message: &pubsubPb.Message{
			From:      msg.From,
			Data:      msg.Data,
			Seqno:     msg.Seqno,
			Topic:     msg.Topic,
			Signature: msg.Signature,
			Key:       msg.Key,
		},
		ReceivedFrom:  "",
		ValidatorData: nil,
	}

	return libp2p.NewMessage(pubsubMsg, marshaller, p2p.Broadcast)
}

// Serialize will serialize a list of p2p messages
func (m *messageVerifier) Serialize(messages []p2p.MessageP2P) ([]byte, error) {
	pubsubMessages := make([][]byte, 0)
	for _, msg := range messages {
		pubsubMsg, err := convertP2PMessagetoPubSubMessage(msg)
		if err != nil {
			log.Trace("convertP2PMessagetoPubSubMessage", "error", err.Error())
			continue
		}

		pubsubMsgBytes, err := pubsubMsg.Marshal()
		if err != nil {
			return nil, err
		}

		pubsubMessages = append(pubsubMessages, pubsubMsgBytes)
	}

	if len(pubsubMessages) == 0 {
		return make([]byte, 0), nil
	}

	b := &batch.Batch{
		Data: pubsubMessages,
	}
	messagesBytes, err := m.marshaller.Marshal(b)
	if err != nil {
		return nil, err
	}

	return messagesBytes, nil
}

// Deserialize will deserialize into a list of p2p messages
func (m *messageVerifier) Deserialize(messagesBytes []byte) ([]p2p.MessageP2P, error) {
	b := batch.Batch{}
	err := m.marshaller.Unmarshal(&b, messagesBytes)
	if err != nil {
		return nil, err
	}

	pubsubMessagesBytes := b.Data

	p2pMessages := make([]p2p.MessageP2P, 0)
	for _, pubsubMessageBytes := range pubsubMessagesBytes {
		var pubsubMsg pubsubPb.Message
		err = pubsubMsg.Unmarshal(pubsubMessageBytes)
		if err != nil {
			return nil, err
		}

		p2pMsg, err := convertPubSubMessagestoP2PMessage(&pubsubMsg, m.marshaller)
		if err != nil {
			log.Trace("convertPubSubMessagestoP2PMessage", "error", err.Error())
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
