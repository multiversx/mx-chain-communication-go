package libp2p_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	ggio "github.com/gogo/protobuf/io"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/data"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/mock"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

const timeout = time.Second * 5
const testMaxSize = 1 << 21

var blankMessageHandler = &mock.MessageProcessorStub{
	ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
		return nil
	},
}

func generateHostStub() *mock.ConnectableHostStub {
	return &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
	}
}

func createConnStub(stream network.Stream, id peer.ID, sk libp2pCrypto.PrivKey, remotePeer peer.ID) *mock.ConnStub {
	return &mock.ConnStub{
		GetStreamsCalled: func() []network.Stream {
			if stream == nil {
				return make([]network.Stream, 0)
			}

			return []network.Stream{stream}
		},
		LocalPeerCalled: func() peer.ID {
			return id
		},
		LocalPrivateKeyCalled: func() libp2pCrypto.PrivKey {
			return sk
		},
		RemotePeerCalled: func() peer.ID {
			return remotePeer
		},
	}
}

func createLibP2PCredentialsDirectSender() (peer.ID, libp2pCrypto.PrivKey) {
	prvKey, _, _ := libp2pCrypto.GenerateSecp256k1Key(rand.Reader)
	id, _ := peer.IDFromPublicKey(prvKey.GetPublic())

	return id, prvKey
}

func TestNewDirectSender(t *testing.T) {
	t.Parallel()

	t.Run("nil context", func(t *testing.T) {
		t.Parallel()

		var ctx context.Context = nil
		ds, err := libp2p.NewDirectSender(
			ctx,
			&mock.ConnectableHostStub{},
			&mock.P2PSignerStub{},
			&testscommon.MarshallerMock{},
			&testscommon.LoggerStub{},
		)

		assert.True(t, check.IfNil(ds))
		assert.Equal(t, p2p.ErrNilContext, err)
	})
	t.Run("nil host", func(t *testing.T) {
		t.Parallel()

		ds, err := libp2p.NewDirectSender(
			context.Background(),
			nil,
			&mock.P2PSignerStub{},
			&testscommon.MarshallerMock{},
			&testscommon.LoggerStub{},
		)

		assert.True(t, check.IfNil(ds))
		assert.Equal(t, p2p.ErrNilHost, err)
	})
	t.Run("nil message handler", func(t *testing.T) {
		t.Parallel()

		ds, err := libp2p.NewDirectSender(
			context.Background(),
			generateHostStub(),
			&mock.P2PSignerStub{},
			&testscommon.MarshallerMock{},
			&testscommon.LoggerStub{},
		)
		assert.False(t, check.IfNil(ds))
		assert.Nil(t, err)

		err = ds.RegisterDirectMessageProcessor(nil)
		assert.Equal(t, p2p.ErrNilDirectSendMessageHandler, err)
	})
	t.Run("nil signer", func(t *testing.T) {
		t.Parallel()

		ds, err := libp2p.NewDirectSender(
			context.Background(),
			generateHostStub(),
			nil,
			&testscommon.MarshallerMock{},
			&testscommon.LoggerStub{},
		)

		assert.True(t, check.IfNil(ds))
		assert.Equal(t, p2p.ErrNilP2PSigner, err)
	})
	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		ds, err := libp2p.NewDirectSender(
			context.Background(),
			generateHostStub(),
			&mock.P2PSignerStub{},
			nil,
			&testscommon.LoggerStub{},
		)

		assert.True(t, check.IfNil(ds))
		assert.Equal(t, p2p.ErrNilMarshaller, err)
	})
	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()

		ds, err := libp2p.NewDirectSender(
			context.Background(),
			generateHostStub(),
			&mock.P2PSignerStub{},
			&testscommon.MarshallerMock{},
			nil,
		)

		assert.True(t, check.IfNil(ds))
		assert.Equal(t, p2p.ErrNilLogger, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ds, err := libp2p.NewDirectSender(
			context.Background(),
			generateHostStub(),
			&mock.P2PSignerStub{},
			&testscommon.MarshallerMock{},
			&testscommon.LoggerStub{},
		)

		assert.False(t, check.IfNil(ds))
		assert.Nil(t, err)
	})
}

func TestNewDirectSender_OkValsShouldCallSetStreamHandlerWithCorrectValues(t *testing.T) {
	t.Parallel()

	var pidCalled protocol.ID
	var handlerCalled network.StreamHandler

	hs := &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {
			pidCalled = pid
			handlerCalled = handler
		},
	}

	_, _ = libp2p.NewDirectSender(
		context.Background(),
		hs,
		&mock.P2PSignerStub{},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)

	assert.NotNil(t, handlerCalled)
	assert.Equal(t, libp2p.DirectSendID, pidCalled)
}

// ------- ProcessReceivedDirectMessage

func TestDirectSender_ProcessReceivedDirectMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		&mock.P2PSignerStub{},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)
	_ = ds.RegisterDirectMessageProcessor(blankMessageHandler)

	err := ds.ProcessReceivedDirectMessage(nil, "peer id")

	assert.Equal(t, p2p.ErrNilMessage, err)
}

func TestDirectSender_ProcessReceivedDirectMessageNilTopicIdsShouldErr(t *testing.T) {
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		&mock.P2PSignerStub{},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)
	_ = ds.RegisterDirectMessageProcessor(blankMessageHandler)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	msg.Topic = nil

	err := ds.ProcessReceivedDirectMessage(msg, id)

	assert.Equal(t, p2p.ErrNilTopic, err)
}

func TestDirectSender_ProcessReceivedDirectMessageKeyFieldIsNotNilShouldErr(t *testing.T) {
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		&mock.P2PSignerStub{},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)
	_ = ds.RegisterDirectMessageProcessor(blankMessageHandler)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	t.Run("Key contains a non-empty byte slice", func(t *testing.T) {
		msg.Key = []byte("random key")

		err := ds.ProcessReceivedDirectMessage(msg, id)
		assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for Key field as the node accepts only nil on this field"))
	})
	t.Run("Key contains an empty byte slice", func(t *testing.T) {
		msg.Key = make([]byte, 0)

		err := ds.ProcessReceivedDirectMessage(msg, id)
		assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for Key field as the node accepts only nil on this field"))
	})
}

func TestDirectSender_ProcessReceivedDirectMessageAbnormalSeqNoFieldShouldErr(t *testing.T) {
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		&mock.P2PSignerStub{},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)
	_ = ds.RegisterDirectMessageProcessor(blankMessageHandler)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = bytes.Repeat([]byte{0x00}, libp2p.SequenceNumberSize+1)
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	err := ds.ProcessReceivedDirectMessage(msg, id)
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
	assert.True(t, strings.Contains(err.Error(), "for SeqNo field as the node accepts only a maximum"))
}

func TestDirectSender_ProcessReceivedDirectMessageAlreadySeenMsgShouldErr(t *testing.T) {
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		&mock.P2PSignerStub{},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)
	_ = ds.RegisterDirectMessageProcessor(blankMessageHandler)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	msgId := string(msg.GetFrom()) + string(msg.GetSeqno())
	ds.SeenMessages().Add(msgId)

	err := ds.ProcessReceivedDirectMessage(msg, id)

	assert.Equal(t, p2p.ErrAlreadySeenMessage, err)
}

func TestDirectSender_ProcessReceivedDirectMessageShouldWork(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshallerMock{}
	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		&mock.P2PSignerStub{},
		marshaller,
		&testscommon.LoggerStub{},
	)
	_ = ds.RegisterDirectMessageProcessor(blankMessageHandler)

	id, _ := createLibP2PCredentialsDirectSender()

	innerMessage := &data.TopicMessage{
		Payload:   []byte("data"),
		Timestamp: time.Now().Unix(),
		Version:   libp2p.CurrentTopicMessageVersion,
	}
	buff, _ := marshaller.Marshal(innerMessage)
	msg := &pb.Message{}
	msg.Data = buff
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	t.Run("Seqno contains bytes", func(t *testing.T) {
		msg.Seqno = []byte("111")
		err := ds.ProcessReceivedDirectMessage(msg, id)
		assert.Nil(t, err)
	})
	t.Run("empty Seqno", func(t *testing.T) {
		msg.Seqno = make([]byte, 0)
		err := ds.ProcessReceivedDirectMessage(msg, id)
		assert.Nil(t, err)
	})
	t.Run("max Seqno", func(t *testing.T) {
		msg.Seqno = bytes.Repeat([]byte{0xFF}, libp2p.SequenceNumberSize)
		err := ds.ProcessReceivedDirectMessage(msg, id)
		assert.Nil(t, err)
	})
	t.Run("min Seqno", func(t *testing.T) {
		msg.Seqno = bytes.Repeat([]byte{0x00}, libp2p.SequenceNumberSize)
		err := ds.ProcessReceivedDirectMessage(msg, id)
		assert.Nil(t, err)
	})
}

func TestDirectSender_ProcessReceivedDirectMessageShouldCallMessageHandler(t *testing.T) {
	t.Parallel()

	wasCalled := false

	marshaller := &testscommon.MarshallerMock{}
	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		&mock.P2PSignerStub{},
		marshaller,
		&testscommon.LoggerStub{},
	)
	_ = ds.RegisterDirectMessageProcessor(&mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			wasCalled = true
			return nil
		},
	})
	id, _ := createLibP2PCredentialsDirectSender()

	innerMessage := &data.TopicMessage{
		Payload:   []byte("data"),
		Timestamp: time.Now().Unix(),
		Version:   libp2p.CurrentTopicMessageVersion,
	}
	buff, _ := marshaller.Marshal(innerMessage)
	msg := &pb.Message{}
	msg.Data = buff
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	_ = ds.ProcessReceivedDirectMessage(msg, id)

	assert.True(t, wasCalled)
}

func TestDirectSender_ProcessReceivedDirectMessageShouldReturnHandlersError(t *testing.T) {
	t.Parallel()

	checkErr := errors.New("checking error")

	marshaller := &testscommon.MarshallerMock{}
	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		&mock.P2PSignerStub{},
		marshaller,
		&testscommon.LoggerStub{},
	)
	_ = ds.RegisterDirectMessageProcessor(&mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			return checkErr
		},
	})

	id, _ := createLibP2PCredentialsDirectSender()

	innerMessage := &data.TopicMessage{
		Payload:   []byte("data"),
		Timestamp: time.Now().Unix(),
		Version:   libp2p.CurrentTopicMessageVersion,
	}
	buff, _ := marshaller.Marshal(innerMessage)
	msg := &pb.Message{}
	msg.Data = buff
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	err := ds.ProcessReceivedDirectMessage(msg, id)

	assert.Equal(t, checkErr, err)
}

// ------- SendDirectToConnectedPeer

func TestDirectSender_SendDirectToConnectedPeerBufferToLargeShouldErr(t *testing.T) {
	t.Parallel()

	netw := &mock.NetworkStub{}

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")

	stream := mock.NewStreamMock()
	_ = stream.SetProtocol(libp2p.DirectSendID)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []network.Conn {
		return []network.Conn{cs}
	}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		&mock.ConnectableHostStub{
			SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
			NetworkCalled: func() network.Network {
				return netw
			},
		},
		&mock.P2PSignerStub{},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)

	messageTooLarge := bytes.Repeat([]byte{65}, libp2p.MaxSendBuffSize)

	err := ds.Send("topic", messageTooLarge, core.PeerID(cs.RemotePeer()))

	assert.True(t, errors.Is(err, p2p.ErrMessageTooLarge))
}

func TestDirectSender_SendDirectToConnectedPeerNotConnectedPeerShouldErr(t *testing.T) {
	t.Parallel()

	netw := &mock.NetworkStub{
		ConnsToPeerCalled: func(p peer.ID) []network.Conn {
			return make([]network.Conn, 0)
		},
	}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		&mock.ConnectableHostStub{
			SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
			NetworkCalled: func() network.Network {
				return netw
			},
		},
		&mock.P2PSignerStub{},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)

	err := ds.Send("topic", []byte("data"), "not connected peer")

	assert.Equal(t, p2p.ErrPeerNotDirectlyConnected, err)
}

func TestDirectSender_SendDirectToConnectedPeerNewStreamErrorsShouldErr(t *testing.T) {
	t.Parallel()

	netw := &mock.NetworkStub{}

	hs := &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
		NetworkCalled: func() network.Network {
			return netw
		},
	}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		hs,
		&mock.P2PSignerStub{},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")
	errNewStream := errors.New("new stream error")

	cs := createConnStub(nil, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []network.Conn {
		return []network.Conn{cs}
	}

	hs.NewStreamCalled = func(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
		return nil, errNewStream
	}

	topic := "topic"
	err := ds.Send(topic, providedData, core.PeerID(cs.RemotePeer()))

	assert.Equal(t, errNewStream, err)
}

func TestDirectSender_SendDirectToConnectedPeerSignFails(t *testing.T) {
	t.Parallel()

	netw := &mock.NetworkStub{}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		&mock.ConnectableHostStub{
			SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
			NetworkCalled: func() network.Network {
				return netw
			},
		},
		&mock.P2PSignerStub{
			SignCalled: func(payload []byte) ([]byte, error) {
				return nil, expectedErr
			},
		},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")

	stream := mock.NewStreamMock()
	_ = stream.SetProtocol(libp2p.DirectSendID)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []network.Conn {
		return []network.Conn{cs}
	}

	topic := "topic"
	err := ds.Send(topic, providedData, core.PeerID(cs.RemotePeer()))

	assert.Equal(t, expectedErr, err)
}

func TestDirectSender_SendDirectToConnectedPeerExistingStreamShouldSendToStream(t *testing.T) {
	t.Parallel()

	netw := &mock.NetworkStub{}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		&mock.ConnectableHostStub{
			SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
			NetworkCalled: func() network.Network {
				return netw
			},
		},
		&mock.P2PSignerStub{},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")

	stream := mock.NewStreamMock()
	err := stream.SetProtocol(libp2p.DirectSendID)
	assert.Nil(t, err)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []network.Conn {
		return []network.Conn{cs}
	}

	receivedMsg := &pb.Message{}
	chanDone := make(chan bool)

	go func(s network.Stream) {
		reader := ggio.NewDelimitedReader(s, testMaxSize)
		for {
			errRead := reader.ReadMsg(receivedMsg)
			if errRead != nil {
				fmt.Println(errRead.Error())
				return
			}

			chanDone <- true
		}
	}(stream)

	topic := "topic"
	err = ds.Send(topic, providedData, core.PeerID(cs.RemotePeer()))
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(timeout):
		assert.Fail(t, "timeout getting data from stream")
		return
	}

	assert.Nil(t, err)
	assert.Equal(t, providedData, receivedMsg.Data)
	assert.Equal(t, topic, *receivedMsg.Topic)
}

func TestDirectSender_SendDirectToConnectedPeerNewStreamShouldSendToStream(t *testing.T) {
	t.Parallel()

	netw := &mock.NetworkStub{}

	hs := &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
		NetworkCalled: func() network.Network {
			return netw
		},
	}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		hs,
		&mock.P2PSignerStub{},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")

	stream := mock.NewStreamMock()
	_ = stream.SetProtocol(libp2p.DirectSendID)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []network.Conn {
		return []network.Conn{cs}
	}

	hs.NewStreamCalled = func(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
		if p == remotePeer && pids[0] == libp2p.DirectSendID {
			return stream, nil
		}
		return nil, errors.New("wrong parameters")
	}

	receivedMsg := &pb.Message{}
	chanDone := make(chan bool)

	go func(s network.Stream) {
		reader := ggio.NewDelimitedReader(s, testMaxSize)
		for {
			err := reader.ReadMsg(receivedMsg)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			chanDone <- true
		}
	}(stream)

	topic := "topic"
	err := ds.Send(topic, providedData, core.PeerID(cs.RemotePeer()))

	select {
	case <-chanDone:
	case <-time.After(timeout):
		assert.Fail(t, "timeout getting data from stream")
		return
	}

	assert.Nil(t, err)
	assert.Equal(t, providedData, receivedMsg.Data)
	assert.Equal(t, topic, *receivedMsg.Topic)
}

// ------- received messages tests

func TestDirectSender_ReceivedSentMessageShouldCallMessageHandlerTestFullCycle(t *testing.T) {
	t.Parallel()

	var streamHandler network.StreamHandler
	netw := &mock.NetworkStub{}

	hs := &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {
			streamHandler = handler
		},
		NetworkCalled: func() network.Network {
			return netw
		},
	}

	var receivedMsg p2p.MessageP2P
	chanDone := make(chan bool)

	marshaller := &testscommon.MarshallerMock{}
	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		hs,
		&mock.P2PSignerStub{},
		marshaller,
		&testscommon.LoggerStub{},
	)
	_ = ds.RegisterDirectMessageProcessor(&mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			receivedMsg = message
			chanDone <- true
			return nil
		},
	})

	id, sk := createLibP2PCredentialsDirectSender()
	realPID, _ := core.NewPeerID("QmY33RXFSbFFpxD2ZfamQvXGULFUsxAYSR2VkTXVewuMNh")
	remotePeer := peer.ID(realPID)

	stream := mock.NewStreamMock()
	stream.SetConn(
		&mock.ConnStub{
			RemotePeerCalled: func() peer.ID {
				return remotePeer
			},
		})
	_ = stream.SetProtocol(libp2p.DirectSendID)

	streamHandler(stream)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []network.Conn {
		return []network.Conn{cs}
	}
	cs.LocalPeerCalled = func() peer.ID {
		return cs.RemotePeer()
	}

	innerMessage := &data.TopicMessage{
		Payload:   []byte("data"),
		Timestamp: time.Now().Unix(),
		Version:   libp2p.CurrentTopicMessageVersion,
	}
	buff, _ := marshaller.Marshal(innerMessage)
	topic := "topic"
	_ = ds.Send(topic, buff, core.PeerID(cs.RemotePeer()))

	select {
	case <-chanDone:
	case <-time.After(timeout):
		assert.Fail(t, "timeout")
		return
	}

	assert.NotNil(t, receivedMsg)
	assert.Equal(t, buff, receivedMsg.Payload())
	assert.Equal(t, topic, receivedMsg.Topic())
}

func TestDirectSender_ProcessReceivedDirectMessageButHandlerNotSetShouldErr(t *testing.T) {
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		&mock.P2PSignerStub{},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)

	err := ds.ProcessReceivedDirectMessage(nil, "peer")

	assert.True(t, errors.Is(err, p2p.ErrNilDirectSendMessageHandler))
}

func TestDirectSender_ProcessReceivedDirectMessageFromMismatchesFromConnectedPeerShouldErr(t *testing.T) {
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		&mock.P2PSignerStub{},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)
	_ = ds.RegisterDirectMessageProcessor(blankMessageHandler)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	err := ds.ProcessReceivedDirectMessage(msg, "not the same peer id")

	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestDirectSender_ProcessReceivedDirectMessageSignatureFails(t *testing.T) {
	t.Parallel()

	verifyCalled := false
	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		&mock.P2PSignerStub{
			VerifyCalled: func(payload []byte, pid core.PeerID, signature []byte) error {
				verifyCalled = true
				return expectedErr
			},
		},
		&testscommon.MarshallerMock{},
		&testscommon.LoggerStub{},
	)
	_ = ds.RegisterDirectMessageProcessor(blankMessageHandler)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	msg.Signature = []byte("signature")
	topic := "topic"
	msg.Topic = &topic

	err := ds.ProcessReceivedDirectMessage(msg, peer.ID(msg.From))

	assert.True(t, errors.Is(err, expectedErr))
	assert.True(t, verifyCalled)
}

func TestDirectSender_ProcessReceivedDirectMessageNewMessageFails(t *testing.T) {
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		&mock.P2PSignerStub{},
		&testscommon.MarshallerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		},
		&testscommon.LoggerStub{},
	)
	_ = ds.RegisterDirectMessageProcessor(blankMessageHandler)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	msg.Signature = []byte("signature")
	topic := "topic"
	msg.Topic = &topic

	err := ds.ProcessReceivedDirectMessage(msg, peer.ID(msg.From))

	assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
}
