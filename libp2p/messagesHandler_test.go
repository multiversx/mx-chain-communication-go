package libp2p_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	atomicCore "github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/secp256k1"
	p2p "github.com/ElrondNetwork/elrond-go-p2p"
	"github.com/ElrondNetwork/elrond-go-p2p/data"
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p"
	p2pCrypto "github.com/ElrondNetwork/elrond-go-p2p/libp2p/crypto"
	"github.com/ElrondNetwork/elrond-go-p2p/mock"
	"github.com/libp2p/go-libp2p-pubsub"
	pubsubPb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
)

var (
	expectedError      = errors.New("expected error")
	providedTopic      = "topic"
	providedIdentifier = "identifier"
	providedPid        = core.PeerID("pid")
	providedData       = []byte("data")
	providedChannel    = "channel"
)

func createMockArgMessagesHandler() libp2p.ArgMessagesHandler {
	return libp2p.ArgMessagesHandler{
		PubSub:       &mock.PubSubStub{},
		DirectSender: &mock.DirectSenderStub{},
		Throttler:    &mock.ThrottlerStub{},
		OutgoingCLB: &mock.ChannelLoadBalancerStub{
			CollectOneElementFromChannelsCalled: func() *libp2p.SendableData {
				return &libp2p.SendableData{}
			},
		},
		TopicsHandler: &mock.TopicsHandlerStub{},
		Marshaller:    &mock.ProtoMarshallerMock{},
		ConnMonitorWrapper: &mock.ConnectionMonitorWrapperStub{
			PeerDenialEvaluatorCalled: func() p2p.PeerDenialEvaluator {
				return &mock.PeerDenialEvaluatorStub{}
			},
		},
		PeersRatingHandler: &mock.PeersRatingHandlerStub{},
		Debugger:           &mock.DebuggerStub{},
		SyncTimer:          &libp2p.LocalSyncTimer{},
		IDProvider: &mock.IDProviderStub{
			IDCalled: func() peer.ID {
				return peer.ID(providedPid)
			},
		},
	}
}

func TestNewMessagesHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil PubSub should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.PubSub = nil
		mh, err := libp2p.NewMessagesHandler(args)
		assert.Equal(t, p2p.ErrNilPubSub, err)
		assert.True(t, check.IfNil(mh))
	})
	t.Run("nil DirectSender should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.DirectSender = nil
		mh, err := libp2p.NewMessagesHandler(args)
		assert.Equal(t, p2p.ErrNilDirectSender, err)
		assert.True(t, check.IfNil(mh))
	})
	t.Run("nil Throttler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.Throttler = nil
		mh, err := libp2p.NewMessagesHandler(args)
		assert.Equal(t, p2p.ErrNilThrottler, err)
		assert.True(t, check.IfNil(mh))
	})
	t.Run("nil OutgoingCLB should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.OutgoingCLB = nil
		mh, err := libp2p.NewMessagesHandler(args)
		assert.Equal(t, p2p.ErrNilChannelLoadBalancer, err)
		assert.True(t, check.IfNil(mh))
	})
	t.Run("nil TopicsHandler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.TopicsHandler = nil
		mh, err := libp2p.NewMessagesHandler(args)
		assert.Equal(t, p2p.ErrNilTopicsHandler, err)
		assert.True(t, check.IfNil(mh))
	})
	t.Run("nil Marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.Marshaller = nil
		mh, err := libp2p.NewMessagesHandler(args)
		assert.Equal(t, p2p.ErrNilMarshaller, err)
		assert.True(t, check.IfNil(mh))
	})
	t.Run("nil ConnMonitorWrapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.ConnMonitorWrapper = nil
		mh, err := libp2p.NewMessagesHandler(args)
		assert.Equal(t, p2p.ErrNilConnectionMonitorWrapper, err)
		assert.True(t, check.IfNil(mh))
	})
	t.Run("nil PeersRatingHandler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.PeersRatingHandler = nil
		mh, err := libp2p.NewMessagesHandler(args)
		assert.Equal(t, p2p.ErrNilPeersRatingHandler, err)
		assert.True(t, check.IfNil(mh))
	})
	t.Run("nil Debugger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.Debugger = nil
		mh, err := libp2p.NewMessagesHandler(args)
		assert.Equal(t, p2p.ErrNilDebugger, err)
		assert.True(t, check.IfNil(mh))
	})
	t.Run("nil SyncTimer should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.SyncTimer = nil
		mh, err := libp2p.NewMessagesHandler(args)
		assert.Equal(t, p2p.ErrNilSyncTimer, err)
		assert.True(t, check.IfNil(mh))
	})
	t.Run("nil IDProvider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.IDProvider = nil
		mh, err := libp2p.NewMessagesHandler(args)
		assert.Equal(t, p2p.ErrNilIDProvider, err)
		assert.True(t, check.IfNil(mh))
	})
	t.Run("RegisterMessageHandler fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.DirectSender = &mock.DirectSenderStub{
			RegisterMessageHandlerCalled: func(handler func(msg *pubsub.Message, fromConnectedPeer core.PeerID) error) error {
				return expectedError
			},
		}
		mh, err := libp2p.NewMessagesHandler(args)
		assert.Equal(t, expectedError, err)
		assert.True(t, check.IfNil(mh))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		mh, err := libp2p.NewMessagesHandler(createMockArgMessagesHandler())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(mh))
		assert.Nil(t, mh.Close())
	})
	t.Run("process loop", func(t *testing.T) {
		t.Parallel()

		t.Run("CollectOneElementFromChannels always returns nil", func(t *testing.T) {
			t.Parallel()

			args := createMockArgMessagesHandler()
			args.OutgoingCLB = &mock.ChannelLoadBalancerStub{
				CollectOneElementFromChannelsCalled: func() *libp2p.SendableData {
					return nil
				},
			}
			wasCalled := atomicCore.Flag{}
			args.TopicsHandler = &mock.TopicsHandlerStub{
				GetTopicCalled: func(topic string) libp2p.PubSubTopic {
					wasCalled.SetValue(true)
					return &pubsub.Topic{}
				},
			}
			mh, _ := libp2p.NewMessagesHandler(args)
			assert.False(t, check.IfNil(mh))
			time.Sleep(time.Millisecond * 3)
			assert.False(t, wasCalled.IsSet())
			assert.Nil(t, mh.Close())
		})
		t.Run("topic not registered should", func(t *testing.T) {
			t.Parallel()

			providedSendableData := &libp2p.SendableData{
				Buff:  []byte("provided buff"),
				Topic: "provided topic",
			}
			args := createMockArgMessagesHandler()
			args.OutgoingCLB = &mock.ChannelLoadBalancerStub{
				CollectOneElementFromChannelsCalled: func() *libp2p.SendableData {
					return providedSendableData
				},
			}
			wasGetTopicCalled := atomicCore.Flag{}
			args.TopicsHandler = &mock.TopicsHandlerStub{
				GetTopicCalled: func(topic string) libp2p.PubSubTopic {
					assert.Equal(t, providedSendableData.Topic, topic)
					wasGetTopicCalled.SetValue(true)
					return nil
				},
			}
			wasMarshalCalled := atomicCore.Flag{}
			args.Marshaller = &mock.MarshallerStub{
				MarshalCalled: func(obj interface{}) ([]byte, error) {
					wasMarshalCalled.SetValue(true)
					return nil, nil
				},
			}
			mh, _ := libp2p.NewMessagesHandler(args)
			assert.False(t, check.IfNil(mh))
			time.Sleep(time.Millisecond * 3)
			assert.True(t, wasGetTopicCalled.IsSet())
			assert.False(t, wasMarshalCalled.IsSet())
			assert.Nil(t, mh.Close())
		})
		t.Run("marshal of data fails", func(t *testing.T) {
			t.Parallel()

			args := createMockArgMessagesHandler()
			wasPublishCalled := atomicCore.Flag{}
			args.TopicsHandler = &mock.TopicsHandlerStub{
				GetTopicCalled: func(topic string) libp2p.PubSubTopic {
					return &mock.PubSubTopicStub{
						PublishCalled: func(ctx context.Context, data []byte, opts ...pubsub.PubOpt) error {
							wasPublishCalled.SetValue(true)
							return nil
						},
					}
				},
			}
			wasMarshalCalled := atomicCore.Flag{}
			args.Marshaller = &mock.MarshallerStub{
				MarshalCalled: func(obj interface{}) ([]byte, error) {
					wasMarshalCalled.SetValue(true)
					return nil, expectedError
				},
			}
			mh, _ := libp2p.NewMessagesHandler(args)
			assert.False(t, check.IfNil(mh))
			time.Sleep(time.Millisecond * 3)
			assert.True(t, wasMarshalCalled.IsSet())
			assert.False(t, wasPublishCalled.IsSet())
			assert.Nil(t, mh.Close())
		})
		t.Run("should work and publish", func(t *testing.T) {
			t.Parallel()

			keyGen := signing.NewKeyGenerator(secp256k1.NewSecp256k1())
			privateKey, _ := keyGen.GeneratePair()
			p2pPrivKey, _ := p2pCrypto.ConvertPrivateKeyToLibp2pPrivateKey(privateKey)
			providedSendableData := &libp2p.SendableData{
				Buff:  []byte("provided buff"),
				Topic: "provided topic",
				Sk:    p2pPrivKey,
			}
			providedMarshalledData := []byte("provided marshalled data")
			args := createMockArgMessagesHandler()
			args.OutgoingCLB = &mock.ChannelLoadBalancerStub{
				CollectOneElementFromChannelsCalled: func() *libp2p.SendableData {
					return providedSendableData
				},
			}
			wasPublishCalled := atomicCore.Flag{}
			args.TopicsHandler = &mock.TopicsHandlerStub{
				GetTopicCalled: func(topic string) libp2p.PubSubTopic {
					return &mock.PubSubTopicStub{
						PublishCalled: func(ctx context.Context, data []byte, opts ...pubsub.PubOpt) error {
							wasPublishCalled.SetValue(true)
							assert.Equal(t, providedMarshalledData, data)
							return nil
						},
					}
				},
			}
			args.Marshaller = &mock.MarshallerStub{
				MarshalCalled: func(obj interface{}) ([]byte, error) {
					return providedMarshalledData, nil
				},
			}
			mh, _ := libp2p.NewMessagesHandler(args)
			assert.False(t, check.IfNil(mh))
			time.Sleep(time.Millisecond * 5)
			assert.True(t, wasPublishCalled.IsSet())
			assert.Nil(t, mh.Close())
		})
	})
}

func TestMessagesHandler_broadcasts(t *testing.T) {
	t.Parallel()

	t.Run("data too big should error", testBroadcastOnChannelBlockingDataTooBig(nil))
	t.Run("empty data should error", testBroadcastOnChannelBlockingEmptyData(nil))
	t.Run("throttler can not process", testBroadcastOnChannelBlockingThrottlerCanNotProcess(nil, false))
	t.Run("should work", testBroadcastOnChannelBlockingShouldWork(nil, false))
	t.Run("Broadcast should work", testBroadcastOnChannelBlockingShouldWork(nil, true))
	t.Run("BroadcastOnChannel fails", testBroadcastOnChannelBlockingThrottlerCanNotProcess(nil, true))
}

func TestMessagesHandler_broadcastsUsingPrivateKey(t *testing.T) {
	t.Parallel()

	keyGen := signing.NewKeyGenerator(secp256k1.NewSecp256k1())
	privateKey, _ := keyGen.GeneratePair()
	skBytes, _ := privateKey.ToByteArray()
	t.Run("unmarshal of sk returns error", func(t *testing.T) {
		t.Parallel()

		mh := libp2p.NewMessagesHandlerWithNoRoutine(createMockArgMessagesHandler())
		assert.False(t, check.IfNil(mh))
		err := mh.BroadcastOnChannelBlockingUsingPrivateKey(providedChannel, providedTopic, providedData, providedPid, []byte("invalid sk"))
		assert.NotNil(t, err)
	})
	t.Run("data too big should error", testBroadcastOnChannelBlockingDataTooBig(skBytes))
	t.Run("empty data should error", testBroadcastOnChannelBlockingEmptyData(skBytes))
	t.Run("throttler can not process", testBroadcastOnChannelBlockingThrottlerCanNotProcess(skBytes, false))
	t.Run("should work", testBroadcastOnChannelBlockingShouldWork(skBytes, false))
	t.Run("BroadcastUsingPrivateKey should work", testBroadcastOnChannelBlockingShouldWork(skBytes, true))
	t.Run("BroadcastOnChannelUsingPrivateKey fails", testBroadcastOnChannelBlockingThrottlerCanNotProcess(skBytes, true))
}

func testBroadcastOnChannelBlockingEmptyData(skBytes []byte) func(t *testing.T) {
	isMultikey := len(skBytes) > 0
	return func(t *testing.T) {
		t.Parallel()

		mh := libp2p.NewMessagesHandlerWithNoRoutine(createMockArgMessagesHandler())
		assert.False(t, check.IfNil(mh))

		var err error
		if isMultikey {
			err = mh.BroadcastOnChannelBlockingUsingPrivateKey(providedChannel, providedTopic, []byte(""), providedPid, skBytes)
		} else {
			err = mh.BroadcastOnChannelBlocking(providedChannel, providedTopic, []byte(""))
		}
		assert.True(t, errors.Is(err, p2p.ErrEmptyBufferToSend))
	}
}

func testBroadcastOnChannelBlockingDataTooBig(skBytes []byte) func(t *testing.T) {
	isMultikey := len(skBytes) > 0
	return func(t *testing.T) {
		t.Parallel()

		mh := libp2p.NewMessagesHandlerWithNoRoutine(createMockArgMessagesHandler())
		assert.False(t, check.IfNil(mh))

		providedBuff := bytes.Repeat([]byte("a"), 1<<21)
		var err error
		if isMultikey {
			err = mh.BroadcastOnChannelBlockingUsingPrivateKey(providedChannel, providedTopic, providedBuff, providedPid, skBytes)
		} else {
			err = mh.BroadcastOnChannelBlocking(providedChannel, providedTopic, providedBuff)
		}
		assert.True(t, errors.Is(err, p2p.ErrMessageTooLarge))
	}
}

func testBroadcastOnChannelBlockingThrottlerCanNotProcess(skBytes []byte, testingExportedMethod bool) func(t *testing.T) {
	isMultikey := len(skBytes) > 0
	return func(t *testing.T) {
		t.Parallel()

		defer checkForPanic(t)
		args := createMockArgMessagesHandler()
		wasCalled := atomicCore.Flag{}
		defer assert.False(t, wasCalled.IsSet())
		args.Throttler = &mock.ThrottlerStub{
			CanProcessCalled: func() bool {
				return false
			},
			StartProcessingCalled: func() {
				wasCalled.SetValue(true)
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		var err error
		if isMultikey {
			if testingExportedMethod {
				mh.BroadcastOnChannelUsingPrivateKey(providedChannel, providedTopic, providedData, providedPid, skBytes)
				time.Sleep(time.Second) // wait for the go routine to finish
				return
			}

			err = mh.BroadcastOnChannelBlockingUsingPrivateKey(providedChannel, providedTopic, providedData, providedPid, skBytes)
			assert.Equal(t, p2p.ErrTooManyGoroutines, err)
			return
		}

		if testingExportedMethod {
			mh.BroadcastOnChannel(providedChannel, providedTopic, providedData)
			time.Sleep(time.Second) // wait for the go routine to finish
			return
		}

		err = mh.BroadcastOnChannelBlocking(providedChannel, providedTopic, providedData)
		assert.Equal(t, p2p.ErrTooManyGoroutines, err)
	}
}

func testBroadcastOnChannelBlockingShouldWork(skBytes []byte, testingExportedMethod bool) func(t *testing.T) {
	isMultikey := len(skBytes) > 0
	return func(t *testing.T) {
		t.Parallel()

		defer checkForPanic(t)
		ch := make(chan *libp2p.SendableData)
		args := createMockArgMessagesHandler()
		wasCalled := atomicCore.Flag{}
		args.Throttler = &mock.ThrottlerStub{
			CanProcessCalled: func() bool {
				return true
			},
			EndProcessingCalled: func() {
				wasCalled.SetValue(true)
			},
		}
		args.OutgoingCLB = &mock.ChannelLoadBalancerStub{
			GetChannelOrDefaultCalled: func(pipe string) chan *libp2p.SendableData {
				return ch
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		go func() {
			var err error
			defer assert.Nil(t, err)

			if isMultikey {
				if testingExportedMethod {
					mh.BroadcastUsingPrivateKey(providedTopic, providedData, providedPid, skBytes)
					return
				}

				err = mh.BroadcastOnChannelBlockingUsingPrivateKey(providedChannel, providedTopic, providedData, providedPid, skBytes)
				return
			}

			if testingExportedMethod {
				mh.Broadcast(providedTopic, providedData)
				return
			}

			err = mh.BroadcastOnChannelBlocking(providedChannel, providedTopic, providedData)
		}()
		waitForChannelBlockingWithFinalCheck(t, ch, func() {
			assert.True(t, wasCalled.IsSet())
		})
	}
}

func TestMessagesHandler_RegisterMessageProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil msg processor should error", func(t *testing.T) {
		t.Parallel()

		mh := libp2p.NewMessagesHandlerWithNoRoutine(createMockArgMessagesHandler())
		assert.False(t, check.IfNil(mh))

		err := mh.RegisterMessageProcessor(providedTopic, providedIdentifier, nil)
		assert.True(t, errors.Is(err, p2p.ErrNilValidator))
	})
	t.Run("new topic - should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.TopicsHandler = &mock.TopicsHandlerStub{
			AddNewTopicProcessorsCalled: func(topic string) libp2p.TopicProcessor {
				assert.Equal(t, providedTopic, topic)
				return &mock.TopicProcessorStub{}
			},
		}
		wasCalled := false
		args.PubSub = &mock.PubSubStub{
			RegisterTopicValidatorCalled: func(topic string, val interface{}, opts ...pubsub.ValidatorOpt) error {
				wasCalled = true
				assert.Equal(t, providedTopic, topic)
				return nil
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.RegisterMessageProcessor(providedTopic, providedIdentifier, &mock.MessageProcessorStub{})
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
	t.Run("new topic - register fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.TopicsHandler = &mock.TopicsHandlerStub{
			AddNewTopicProcessorsCalled: func(topic string) libp2p.TopicProcessor {
				return &mock.TopicProcessorStub{}
			},
		}
		args.PubSub = &mock.PubSubStub{
			RegisterTopicValidatorCalled: func(topic string, val interface{}, opts ...pubsub.ValidatorOpt) error {
				return expectedError
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.RegisterMessageProcessor(providedTopic, providedIdentifier, &mock.MessageProcessorStub{})
		assert.Equal(t, expectedError, err)
	})
	t.Run("known topic - should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.TopicsHandler = &mock.TopicsHandlerStub{
			GetTopicProcessorsCalled: func(topic string) libp2p.TopicProcessor {
				assert.Equal(t, providedTopic, topic)
				return &mock.TopicProcessorStub{}
			},
		}
		wasCalled := false
		args.PubSub = &mock.PubSubStub{
			RegisterTopicValidatorCalled: func(topic string, val interface{}, opts ...pubsub.ValidatorOpt) error {
				wasCalled = true
				return nil
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.RegisterMessageProcessor(providedTopic, providedIdentifier, &mock.MessageProcessorStub{})
		assert.Nil(t, err)
		assert.False(t, wasCalled)
	})
	t.Run("known topic - add topic processors fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.TopicsHandler = &mock.TopicsHandlerStub{
			GetTopicProcessorsCalled: func(topic string) libp2p.TopicProcessor {
				return &mock.TopicProcessorStub{
					AddTopicProcessorCalled: func(identifier string, processor p2p.MessageProcessor) error {
						return expectedError
					},
				}
			},
		}
		args.PubSub = &mock.PubSubStub{
			RegisterTopicValidatorCalled: func(topic string, val interface{}, opts ...pubsub.ValidatorOpt) error {
				return nil
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.RegisterMessageProcessor(providedTopic, providedIdentifier, &mock.MessageProcessorStub{})
		assert.True(t, errors.Is(err, expectedError))
	})
}

func TestMessagesHandler_pubsubCallback(t *testing.T) {
	t.Parallel()

	realPID, _ := core.NewPeerID("QmY33RXFSbFFpxD2ZfamQvXGULFUsxAYSR2VkTXVewuMNh")
	peerID := peer.ID(realPID)
	t.Run("transform and check message fails(nil msg) should return false", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		tp := &mock.TopicProcessorStub{}
		cb := mh.PubsubCallback(tp, providedTopic)
		assert.False(t, cb(context.Background(), peerID, nil))
	})
	t.Run("process message fails should return false", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.PeersRatingHandler = &mock.PeersRatingHandlerStub{
			IncreaseRatingCalled: func(pid core.PeerID) {
				assert.Fail(t, "should not have been called")
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		tp := &mock.TopicProcessorStub{
			GetListCalled: func() ([]string, []p2p.MessageProcessor) {
				return []string{providedIdentifier}, []p2p.MessageProcessor{&mock.MessageProcessorStub{
					ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
						return expectedError
					},
				}}
			},
		}
		cb := mh.PubsubCallback(tp, providedTopic)
		assert.False(t, cb(context.Background(), peerID, createPubSubMsgWithTimestamp(time.Now().Unix(), realPID, args.Marshaller)))
	})
	t.Run("should work and return true", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		wasCalled := false
		args.PeersRatingHandler = &mock.PeersRatingHandlerStub{
			IncreaseRatingCalled: func(pid core.PeerID) {
				wasCalled = true
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		tp := &mock.TopicProcessorStub{}
		cb := mh.PubsubCallback(tp, providedTopic)
		assert.True(t, cb(context.Background(), peerID, createPubSubMsgWithTimestamp(time.Now().Unix(), realPID, args.Marshaller)))
		assert.True(t, wasCalled)
	})
}

func TestMessagesHandler_UnregisterMessageProcessor(t *testing.T) {
	t.Parallel()

	t.Run("missing topic should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.TopicsHandler = &mock.TopicsHandlerStub{
			GetTopicProcessorsCalled: func(topic string) libp2p.TopicProcessor {
				return nil
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.UnregisterMessageProcessor(providedTopic, providedIdentifier)
		assert.Nil(t, err)
	})
	t.Run("remove topic processor returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.TopicsHandler = &mock.TopicsHandlerStub{
			GetTopicProcessorsCalled: func(topic string) libp2p.TopicProcessor {
				return &mock.TopicProcessorStub{
					RemoveTopicProcessorCalled: func(identifier string) error {
						return expectedError
					},
				}
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.UnregisterMessageProcessor(providedTopic, providedIdentifier)
		assert.Equal(t, expectedError, err)
	})
	t.Run("empty identifiers should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.TopicsHandler = &mock.TopicsHandlerStub{
			GetTopicProcessorsCalled: func(topic string) libp2p.TopicProcessor {
				return &mock.TopicProcessorStub{
					GetListCalled: func() ([]string, []p2p.MessageProcessor) {
						return []string{}, nil
					},
				}
			},
		}
		wasCalled := false
		args.PubSub = &mock.PubSubStub{
			UnregisterTopicValidatorCalled: func(topic string) error {
				wasCalled = true
				return nil
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.UnregisterMessageProcessor(providedTopic, providedIdentifier)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
	t.Run("identifiers left should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.TopicsHandler = &mock.TopicsHandlerStub{
			GetTopicProcessorsCalled: func(topic string) libp2p.TopicProcessor {
				return &mock.TopicProcessorStub{
					GetListCalled: func() ([]string, []p2p.MessageProcessor) {
						return []string{"id1", "id2"}, nil
					},
				}
			},
		}
		wasCalled := false
		args.PubSub = &mock.PubSubStub{
			UnregisterTopicValidatorCalled: func(topic string) error {
				wasCalled = true
				return nil
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.UnregisterMessageProcessor(providedTopic, providedIdentifier)
		assert.Nil(t, err)
		assert.False(t, wasCalled)
	})
}

func TestMessagesHandler_UnregisterAllMessageProcessors(t *testing.T) {
	t.Parallel()

	t.Run("pubSub returns error", func(t *testing.T) {
		t.Parallel()

		providedMap := map[string]libp2p.TopicProcessor{
			"topic1": &mock.TopicProcessorStub{
				RemoveTopicProcessorCalled: func(identifier string) error {
					return nil
				},
			},
			"topic2": &mock.TopicProcessorStub{
				RemoveTopicProcessorCalled: func(identifier string) error {
					return nil
				},
			},
		}
		args := createMockArgMessagesHandler()
		args.TopicsHandler = &mock.TopicsHandlerStub{
			GetAllTopicsProcessorsCalled: func() map[string]libp2p.TopicProcessor {
				return providedMap
			},
		}
		args.PubSub = &mock.PubSubStub{
			UnregisterTopicValidatorCalled: func(topic string) error {
				return expectedError
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.UnregisterAllMessageProcessors()
		assert.Equal(t, expectedError, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedMap := map[string]libp2p.TopicProcessor{
			"topic1": &mock.TopicProcessorStub{
				RemoveTopicProcessorCalled: func(identifier string) error {
					return nil
				},
			},
			"topic2": &mock.TopicProcessorStub{
				RemoveTopicProcessorCalled: func(identifier string) error {
					return nil
				},
			},
		}
		args := createMockArgMessagesHandler()
		counterRemove := uint32(0)
		args.TopicsHandler = &mock.TopicsHandlerStub{
			GetAllTopicsProcessorsCalled: func() map[string]libp2p.TopicProcessor {
				return providedMap
			},
			RemoveTopicProcessorsCalled: func(topic string) {
				atomic.AddUint32(&counterRemove, 1)
			},
		}
		counterUnregister := uint32(0)
		args.PubSub = &mock.PubSubStub{
			UnregisterTopicValidatorCalled: func(topic string) error {
				atomic.AddUint32(&counterUnregister, 1)
				return nil
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.UnregisterAllMessageProcessors()
		assert.Nil(t, err)
		expectedCounters := uint32(len(providedMap))
		assert.Equal(t, expectedCounters, atomic.LoadUint32(&counterRemove))
		assert.Equal(t, expectedCounters, atomic.LoadUint32(&counterUnregister))
	})
}

func TestMessagesHandler_SendToConnectedPeer(t *testing.T) {
	t.Parallel()

	t.Run("data not sendable should error", func(t *testing.T) {
		t.Parallel()

		mh := libp2p.NewMessagesHandlerWithNoRoutine(createMockArgMessagesHandler())
		assert.False(t, check.IfNil(mh))
		err := mh.SendToConnectedPeer(providedTopic, []byte(""), providedPid)
		assert.True(t, errors.Is(err, p2p.ErrEmptyBufferToSend))
	})
	t.Run("marshal returns error should return nil", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.Marshaller = &mock.MarshallerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedError
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))
		err := mh.SendToConnectedPeer(providedTopic, providedData, providedPid)
		assert.Nil(t, err)
	})
	t.Run("should work to other peers", func(t *testing.T) {
		t.Parallel()

		providedPeer := core.PeerID("provided pid")
		providedSendableData := []byte("provided data")
		args := createMockArgMessagesHandler()
		args.Marshaller = &mock.MarshallerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return providedSendableData, nil
			},
		}
		wasCalled := false
		args.DirectSender = &mock.DirectSenderStub{
			SendCalled: func(topic string, buff []byte, peer core.PeerID) error {
				wasCalled = true
				assert.Equal(t, providedTopic, topic)
				assert.Equal(t, providedSendableData, buff)
				assert.Equal(t, providedPeer, peer)
				return nil
			},
		}

		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))
		err := mh.SendToConnectedPeer(providedTopic, providedData, providedPeer)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
	t.Run("send to self, transform message fails", func(t *testing.T) {
		t.Parallel()

		mh := libp2p.NewMessagesHandlerWithNoRoutine(createMockArgMessagesHandler())
		assert.False(t, check.IfNil(mh))

		err := mh.SendToConnectedPeer(providedTopic, providedData, providedPid)
		assert.NotNil(t, err)
	})
	t.Run("send to self, transform message fails", func(t *testing.T) {
		t.Parallel()

		mh := libp2p.NewMessagesHandlerWithNoRoutine(createMockArgMessagesHandler())
		assert.False(t, check.IfNil(mh))

		err := mh.SendToConnectedPeer(providedTopic, providedData, providedPid)
		assert.NotNil(t, err)
	})
	realPID, _ := core.NewPeerID("QmY33RXFSbFFpxD2ZfamQvXGULFUsxAYSR2VkTXVewuMNh")
	t.Run("send to self, nil topic procs should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.IDProvider = &mock.IDProviderStub{
			IDCalled: func() peer.ID {
				return peer.ID(realPID)
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.SendToConnectedPeer(providedTopic, providedData, realPID)
		assert.True(t, errors.Is(err, p2p.ErrNilValidator))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.IDProvider = &mock.IDProviderStub{
			IDCalled: func() peer.ID {
				return peer.ID(realPID)
			},
		}
		args.TopicsHandler = &mock.TopicsHandlerStub{
			GetTopicProcessorsCalled: func(topic string) libp2p.TopicProcessor {
				return &mock.TopicProcessorStub{
					GetListCalled: func() ([]string, []p2p.MessageProcessor) {
						return []string{providedTopic}, []p2p.MessageProcessor{&mock.MessageProcessorStub{}}
					},
				}
			},
		}
		ch := make(chan *libp2p.SendableData)
		args.PeersRatingHandler = &mock.PeersRatingHandlerStub{
			IncreaseRatingCalled: func(pid core.PeerID) {
				assert.Equal(t, realPID, pid)
				ch <- &libp2p.SendableData{}
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.SendToConnectedPeer(providedTopic, providedData, realPID)
		assert.Nil(t, err)
		waitForChannelBlockingWithFinalCheck(t, ch, func() {})
	})
	t.Run("should work, but one message fails to process", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.IDProvider = &mock.IDProviderStub{
			IDCalled: func() peer.ID {
				return peer.ID(realPID)
			},
		}
		counter := uint32(0)
		providedProcessors := []p2p.MessageProcessor{
			&mock.MessageProcessorStub{
				ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
					atomic.AddUint32(&counter, 1)
					return expectedError
				},
			}, &mock.MessageProcessorStub{
				ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
					atomic.AddUint32(&counter, 1)
					return nil
				},
			},
		}
		args.TopicsHandler = &mock.TopicsHandlerStub{
			GetTopicProcessorsCalled: func(topic string) libp2p.TopicProcessor {
				return &mock.TopicProcessorStub{
					GetListCalled: func() ([]string, []p2p.MessageProcessor) {
						return []string{providedTopic, providedTopic}, providedProcessors
					},
				}
			},
		}
		args.PeersRatingHandler = &mock.PeersRatingHandlerStub{
			IncreaseRatingCalled: func(pid core.PeerID) {
				assert.Fail(t, "should have not been called")
			},
		}
		ch := make(chan *libp2p.SendableData)
		args.Debugger = &mock.DebuggerStub{
			AddIncomingMessageCalled: func(topic string, size uint64, isRejected bool) {
				assert.Equal(t, providedTopic, topic)
				ch <- &libp2p.SendableData{}
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.SendToConnectedPeer(providedTopic, providedData, realPID)
		assert.Nil(t, err)
		waitForChannelBlockingWithFinalCheck(t, ch, func() {
			assert.Equal(t, uint32(2), atomic.LoadUint32(&counter))
		})
	})
}

func waitForChannelBlockingWithFinalCheck(t *testing.T, ch chan *libp2p.SendableData, finalCheck func()) {
	for {
		select {
		case <-ch:
			time.Sleep(time.Millisecond * 50) // allow processing to end
			finalCheck()
			return
		case <-time.After(time.Second * 3):
			assert.Fail(t, "failed due to timeout")
			return
		}
	}
}

func checkForPanic(t *testing.T) {
	r := recover()
	if r != nil {
		assert.Fail(t, "should not have panicked")
	}
}

func TestMessagesHandler_blacklistPid(t *testing.T) {
	t.Parallel()

	t.Run("pid already denied should return", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		wasCalled := false
		args.ConnMonitorWrapper = &mock.ConnectionMonitorWrapperStub{
			PeerDenialEvaluatorCalled: func() p2p.PeerDenialEvaluator {
				return &mock.PeerDenialEvaluatorStub{
					IsDeniedCalled: func(pid core.PeerID) bool {
						wasCalled = true
						assert.Equal(t, providedPid, pid)
						return true
					},
				}
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		mh.BlacklistPid(providedPid, time.Second)
		assert.True(t, wasCalled)
	})
	t.Run("empty pid should return", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		wasCalled := false
		args.ConnMonitorWrapper = &mock.ConnectionMonitorWrapperStub{
			PeerDenialEvaluatorCalled: func() p2p.PeerDenialEvaluator {
				return &mock.PeerDenialEvaluatorStub{
					UpsertPeerIDCalled: func(pid core.PeerID, duration time.Duration) error {
						wasCalled = true
						return nil
					},
				}
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		mh.BlacklistPid("", time.Second)
		assert.False(t, wasCalled)
	})
	t.Run("upsert returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		wasCalled := false
		args.ConnMonitorWrapper = &mock.ConnectionMonitorWrapperStub{
			PeerDenialEvaluatorCalled: func() p2p.PeerDenialEvaluator {
				return &mock.PeerDenialEvaluatorStub{
					UpsertPeerIDCalled: func(pid core.PeerID, duration time.Duration) error {
						wasCalled = true
						return expectedError
					},
				}
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		mh.BlacklistPid(providedPid, time.Second)
		assert.True(t, wasCalled)
	})
}

func TestMessagesHandler_transformAndCheckMessage(t *testing.T) {
	t.Parallel()

	realPID, _ := core.NewPeerID("QmY33RXFSbFFpxD2ZfamQvXGULFUsxAYSR2VkTXVewuMNh")
	t.Run("new message fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		wasCalled := false
		args.ConnMonitorWrapper = &mock.ConnectionMonitorWrapperStub{
			PeerDenialEvaluatorCalled: func() p2p.PeerDenialEvaluator {
				return &mock.PeerDenialEvaluatorStub{
					UpsertPeerIDCalled: func(pid core.PeerID, duration time.Duration) error {
						wasCalled = true
						return nil
					},
				}
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		pubSubMsg := createPubSubMsgWithTimestamp(time.Now().Unix(), realPID, args.Marshaller)
		pubSubMsg.Topic = nil // fail NewMessage
		msg, err := mh.TransformAndCheckMessage(pubSubMsg, "pid", providedTopic)
		assert.Nil(t, msg)
		assert.NotNil(t, err)
		assert.True(t, wasCalled)
	})
	t.Run("validate timestamp fails, message in the future", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		timeStamp := time.Now().Unix() + 1
		timeStamp += int64(libp2p.AcceptMessagesInAdvanceDuration.Seconds())
		pubSubMsg := createPubSubMsgWithTimestamp(timeStamp, realPID, args.Marshaller)
		msg, err := mh.TransformAndCheckMessage(pubSubMsg, realPID, providedTopic)
		assert.Nil(t, msg)
		assert.True(t, errors.Is(err, p2p.ErrMessageTooNew))
	})
	t.Run("validate timestamp fails, message too old", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		timeStamp := time.Now().Unix() - 1
		timeStamp -= int64(libp2p.AcceptMessagesInAdvanceDuration.Seconds())
		timeStamp -= int64(libp2p.PubsubTimeCacheDuration.Seconds())
		pubSubMsg := createPubSubMsgWithTimestamp(timeStamp, realPID, args.Marshaller)
		msg, err := mh.TransformAndCheckMessage(pubSubMsg, realPID, providedTopic)
		assert.Nil(t, msg)
		assert.True(t, errors.Is(err, p2p.ErrMessageTooOld))
	})
	t.Run("validate timestamp fails from self, message too old", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.IDProvider = &mock.IDProviderStub{
			IDCalled: func() peer.ID {
				return peer.ID(realPID)
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		timeStamp := time.Now().Unix() - 1
		timeStamp -= int64(libp2p.AcceptMessagesInAdvanceDuration.Seconds())
		timeStamp -= int64(libp2p.PubsubTimeCacheDuration.Seconds())
		pubSubMsg := createPubSubMsgWithTimestamp(timeStamp, realPID, args.Marshaller)
		msg, err := mh.TransformAndCheckMessage(pubSubMsg, realPID, providedTopic)
		assert.Nil(t, msg)
		assert.True(t, errors.Is(err, p2p.ErrMessageTooOld))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		pubSubMsg := createPubSubMsgWithTimestamp(time.Now().Unix(), realPID, args.Marshaller)
		msg, err := mh.TransformAndCheckMessage(pubSubMsg, realPID, providedTopic)
		assert.NotNil(t, msg)
		assert.Nil(t, err)
	})
}

func createPubSubMsgWithTimestamp(timestamp int64, pid core.PeerID, marshaller marshal.Marshalizer) *pubsub.Message {
	innerMessage := &data.TopicMessage{
		Payload:   providedData,
		Timestamp: timestamp,
		Version:   1,
	}

	buff, _ := marshaller.Marshal(innerMessage)
	return &pubsub.Message{
		Message: &pubsubPb.Message{
			From:      pid.Bytes(),
			Data:      buff,
			Topic:     &providedTopic,
			Signature: pid.Bytes(),
		},
	}
}

func TestMessagesHandler_CreateTopic(t *testing.T) {
	t.Parallel()

	t.Run("existing topic should return nil", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.TopicsHandler = &mock.TopicsHandlerStub{
			HasTopicCalled: func(topic string) bool {
				return true
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.CreateTopic(providedTopic, false)
		assert.Nil(t, err)
	})
	t.Run("pubSub Join returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgMessagesHandler()
		args.PubSub = &mock.PubSubStub{
			JoinCalled: func(topic string, opts ...pubsub.TopicOpt) (*pubsub.Topic, error) {
				return nil, expectedError
			},
		}
		mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
		assert.False(t, check.IfNil(mh))

		err := mh.CreateTopic(providedTopic, false)
		assert.True(t, errors.Is(err, expectedError))
	})
}

func TestMessagesHandler_HasTopic(t *testing.T) {
	t.Parallel()

	mh := libp2p.NewMessagesHandlerWithNoRoutine(createMockArgMessagesHandler())
	assert.False(t, check.IfNil(mh))

	assert.False(t, mh.HasTopic(providedTopic))
}

func TestMessagesHandler_UnJoinAllTopics(t *testing.T) {
	t.Parallel()

	args := createMockArgMessagesHandler()
	counterGetAllTopics := 0
	counterCancel := 0
	counterRemoveTopic := 0
	args.TopicsHandler = &mock.TopicsHandlerStub{
		GetAllTopicsCalled: func() map[string]libp2p.PubSubTopic {
			return map[string]libp2p.PubSubTopic{
				"topic1": &mock.PubSubTopicStub{
					CloseCalled: func() error {
						counterGetAllTopics++
						return nil
					},
				},
				"topic2": &mock.PubSubTopicStub{
					CloseCalled: func() error {
						counterGetAllTopics++
						return expectedError
					},
				},
			}
		},
		GetSubscriptionCalled: func(topic string) libp2p.PubSubSubscription {
			return &mock.PubSubSubscriptionStub{
				CancelCalled: func() {
					counterCancel++
				},
			}
		},
		RemoveTopicCalled: func(topic string) {
			counterRemoveTopic++
		},
	}

	mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
	assert.False(t, check.IfNil(mh))

	err := mh.UnJoinAllTopics()
	assert.Equal(t, expectedError, err)
	assert.Equal(t, 2, counterGetAllTopics)
	assert.Equal(t, 2, counterCancel)
	assert.Equal(t, 2, counterRemoveTopic)
}

func TestMessagesHandler_Close(t *testing.T) {
	t.Parallel()

	defer checkForPanic(t)
	args := createMockArgMessagesHandler()
	errCloseCLB := fmt.Errorf("%w for CLB", expectedError)
	errCloseDebugger := fmt.Errorf("%w for debugger", expectedError)
	args.OutgoingCLB = &mock.ChannelLoadBalancerStub{
		CloseCalled: func() error {
			return errCloseCLB
		},
	}
	args.Debugger = &mock.DebuggerStub{
		CloseCalled: func() error {
			return errCloseDebugger
		},
	}

	mh := libp2p.NewMessagesHandlerWithNoRoutine(args)
	assert.False(t, check.IfNil(mh))
	err := mh.Close()
	assert.Equal(t, errCloseDebugger, err)
}
