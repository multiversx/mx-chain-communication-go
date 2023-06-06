package facade

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/mock"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

const (
	missingMessengerType = "missing messenger type"
	topic                = "topic"
	channel              = "channel"
	identifier           = "identifier"
	pid                  = core.PeerID("pid")
	address              = "address"
)

var (
	buff       = []byte("buff")
	privateKey = []byte("private key")
	payload    = []byte("payload")
	signature  = []byte("signature")
)

func createMessengersStubs() (*mock.MessengerStub, *mock.MessengerStub) {
	m1 := &mock.MessengerStub{
		TypeCalled: func() p2p.NetworkMessengerType {
			return p2p.RegularMessenger
		},
	}
	m2 := &mock.MessengerStub{
		TypeCalled: func() p2p.NetworkMessengerType {
			return p2p.FullArchiveMessenger
		},
	}

	return m1, m2
}

func TestNewNetworkMessengersFacade(t *testing.T) {
	t.Parallel()

	t.Run("no messenger should error", func(t *testing.T) {
		t.Parallel()

		facadeInstance, err := NewNetworkMessengersFacade()
		require.Equal(t, p2p.ErrEmptyMessengersList, err)
		require.Nil(t, facadeInstance)
	})
	t.Run("nil messenger should error", func(t *testing.T) {
		t.Parallel()

		facadeInstance, err := NewNetworkMessengersFacade(&mock.MessengerStub{}, nil)
		require.Equal(t, p2p.ErrNilMessenger, err)
		require.Nil(t, facadeInstance)
	})
	t.Run("should work with one messenger", func(t *testing.T) {
		t.Parallel()

		facadeInstance, err := NewNetworkMessengersFacade(&mock.MessengerStub{})
		require.NoError(t, err)
		require.NotNil(t, facadeInstance)
	})
	t.Run("should work multiple messengers", func(t *testing.T) {
		t.Parallel()

		facadeInstance, err := NewNetworkMessengersFacade(&mock.MessengerStub{}, &mock.MessengerStub{}, &mock.MessengerStub{})
		require.NoError(t, err)
		require.NotNil(t, facadeInstance)
	})
}

func TestNetworkMessengersFacade_CreateTopic(t *testing.T) {
	t.Parallel()

	t.Run("CreateTopic: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.CreateTopicCalled = func(name string, createChannelForTopic bool) error {
			wasCalled = true
			return nil
		}
		fullArchiveMes.CreateTopicCalled = func(name string, createChannelForTopic bool) error {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.CreateTopic(topic, false)
		require.NoError(t, err)
		require.True(t, wasCalled)
	})
	t.Run("CreateTopicForMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.CreateTopicCalled = func(name string, createChannelForTopic bool) error {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.CreateTopicCalled = func(name string, createChannelForTopic bool) error {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.CreateTopicForMessenger(p2p.FullArchiveMessenger, topic, false)
		require.NoError(t, err)
		require.True(t, wasCalled)

		err = facadeInstance.CreateTopicForMessenger(missingMessengerType, topic, false)
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
	})
}

func TestNetworkMessengersFacade_HasTopic(t *testing.T) {
	t.Parallel()

	t.Run("HasTopic: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.HasTopicCalled = func(name string) bool {
			wasCalled = true
			return true
		}
		fullArchiveMes.HasTopicCalled = func(name string) bool {
			require.Fail(t, "should have not been called")
			return true
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		hasTopic := facadeInstance.HasTopic(topic)
		require.True(t, hasTopic)
		require.True(t, wasCalled)
	})
	t.Run("HasMessengerTopic", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.HasTopicCalled = func(name string) bool {
			require.Fail(t, "should have not been called")
			return true
		}
		fullArchiveMes.HasTopicCalled = func(name string) bool {
			wasCalled = true
			return true
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		hasTopic := facadeInstance.HasMessengerTopic(p2p.FullArchiveMessenger, topic)
		require.True(t, hasTopic)
		require.True(t, wasCalled)

		hasTopic = facadeInstance.HasMessengerTopic(missingMessengerType, topic)
		require.False(t, hasTopic)
	})
}

func TestNetworkMessengersFacade_RegisterMessageProcessor(t *testing.T) {
	t.Parallel()

	t.Run("RegisterMessageProcessor: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.RegisterMessageProcessorCalled = func(topic string, identifier string, handler p2p.MessageProcessor) error {
			wasCalled = true
			return nil
		}
		fullArchiveMes.RegisterMessageProcessorCalled = func(topic string, identifier string, handler p2p.MessageProcessor) error {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.RegisterMessageProcessor(topic, identifier, &mock.MessageProcessorStub{})
		require.NoError(t, err)
		require.True(t, wasCalled)
	})
	t.Run("RegisterMessageProcessorForMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.RegisterMessageProcessorCalled = func(topic string, identifier string, handler p2p.MessageProcessor) error {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.RegisterMessageProcessorCalled = func(topic string, identifier string, handler p2p.MessageProcessor) error {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.RegisterMessageProcessorForMessenger(p2p.FullArchiveMessenger, topic, identifier, &mock.MessageProcessorStub{})
		require.NoError(t, err)
		require.True(t, wasCalled)

		err = facadeInstance.RegisterMessageProcessorForMessenger(missingMessengerType, topic, identifier, &mock.MessageProcessorStub{})
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
	})
}

func TestNetworkMessengersFacade_UnRegisterAllMessageProcessor(t *testing.T) {
	t.Parallel()

	t.Run("UnregisterAllMessageProcessors: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.UnregisterAllMessageProcessorsCalled = func() error {
			wasCalled = true
			return nil
		}
		fullArchiveMes.UnregisterAllMessageProcessorsCalled = func() error {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.UnregisterAllMessageProcessors()
		require.NoError(t, err)
		require.True(t, wasCalled)
	})
	t.Run("UnregisterAllMessageProcessorsForMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.UnregisterAllMessageProcessorsCalled = func() error {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.UnregisterAllMessageProcessorsCalled = func() error {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.UnregisterAllMessageProcessorsForMessenger(p2p.FullArchiveMessenger)
		require.NoError(t, err)
		require.True(t, wasCalled)

		err = facadeInstance.UnregisterAllMessageProcessorsForMessenger(missingMessengerType)
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
	})
}

func TestNetworkMessengersFacade_UnregisterMessageProcessor(t *testing.T) {
	t.Parallel()

	t.Run("UnregisterAllMessageProcessors: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.UnregisterMessageProcessorCalled = func(topic string, identifier string) error {
			wasCalled = true
			return nil
		}
		fullArchiveMes.UnregisterMessageProcessorCalled = func(topic string, identifier string) error {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.UnregisterMessageProcessor(topic, identifier)
		require.NoError(t, err)
		require.True(t, wasCalled)
	})
	t.Run("UnregisterMessageProcessorForMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.UnregisterMessageProcessorCalled = func(topic string, identifier string) error {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.UnregisterMessageProcessorCalled = func(topic string, identifier string) error {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.UnregisterMessageProcessorForMessenger(p2p.FullArchiveMessenger, topic, identifier)
		require.NoError(t, err)
		require.True(t, wasCalled)

		err = facadeInstance.UnregisterMessageProcessorForMessenger(missingMessengerType, topic, identifier)
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
	})
}

func TestNetworkMessengersFacade_Broadcast(t *testing.T) {
	t.Parallel()

	t.Run("Broadcast: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.BroadcastCalled = func(topic string, buff []byte) {
			wasCalled = true
		}
		fullArchiveMes.BroadcastCalled = func(topic string, buff []byte) {
			require.Fail(t, "should have not been called")
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		facadeInstance.Broadcast(topic, buff)
		require.True(t, wasCalled)
	})
	t.Run("broadcastOnMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.BroadcastCalled = func(topic string, buff []byte) {
			require.Fail(t, "should have not been called")
		}
		fullArchiveMes.BroadcastCalled = func(topic string, buff []byte) {
			wasCalled = true
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		facadeInstance.broadcastOnMessenger(p2p.FullArchiveMessenger, topic, buff)
		require.True(t, wasCalled)

		facadeInstance.broadcastOnMessenger(missingMessengerType, topic, buff)
	})
}

func TestNetworkMessengersFacade_BroadcastOnChannel(t *testing.T) {
	t.Parallel()

	t.Run("BroadcastOnChannel: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.BroadcastOnChannelCalled = func(channel string, topic string, buff []byte) {
			wasCalled = true
		}
		fullArchiveMes.BroadcastOnChannelCalled = func(channel string, topic string, buff []byte) {
			require.Fail(t, "should have not been called")
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		facadeInstance.BroadcastOnChannel(channel, topic, buff)
		require.True(t, wasCalled)
	})
	t.Run("broadcastOnChannelOnMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.BroadcastOnChannelCalled = func(channel string, topic string, buff []byte) {
			require.Fail(t, "should have not been called")
		}
		fullArchiveMes.BroadcastOnChannelCalled = func(channel string, topic string, buff []byte) {
			wasCalled = true
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		facadeInstance.broadcastOnChannelOnMessenger(p2p.FullArchiveMessenger, channel, topic, buff)
		require.True(t, wasCalled)

		facadeInstance.broadcastOnChannelOnMessenger(missingMessengerType, channel, topic, buff)
	})
}

func TestNetworkMessengersFacade_BroadcastUsingPrivateKey(t *testing.T) {
	t.Parallel()

	t.Run("BroadcastUsingPrivateKey: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.BroadcastUsingPrivateKeyCalled = func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			wasCalled = true
		}
		fullArchiveMes.BroadcastUsingPrivateKeyCalled = func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			require.Fail(t, "should have not been called")
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		facadeInstance.BroadcastUsingPrivateKey(topic, buff, pid, privateKey)
		require.True(t, wasCalled)
	})
	t.Run("broadcastUsingPrivateKeyOnMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.BroadcastUsingPrivateKeyCalled = func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			require.Fail(t, "should have not been called")
		}
		fullArchiveMes.BroadcastUsingPrivateKeyCalled = func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			wasCalled = true
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		facadeInstance.broadcastUsingPrivateKeyOnMessenger(p2p.FullArchiveMessenger, topic, buff, pid, privateKey)
		require.True(t, wasCalled)

		facadeInstance.broadcastUsingPrivateKeyOnMessenger(missingMessengerType, topic, buff, pid, privateKey)
	})
}

func TestNetworkMessengersFacade_BroadcastOnChannelUsingPrivateKey(t *testing.T) {
	t.Parallel()

	t.Run("BroadcastOnChannelUsingPrivateKey: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.BroadcastOnChannelUsingPrivateKeyCalled = func(channel string, topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			wasCalled = true
		}
		fullArchiveMes.BroadcastOnChannelUsingPrivateKeyCalled = func(channel string, topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			require.Fail(t, "should have not been called")
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		facadeInstance.BroadcastOnChannelUsingPrivateKey(channel, topic, buff, pid, privateKey)
		require.True(t, wasCalled)
	})
	t.Run("broadcastOnChannelUsingPrivateKeyOnMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.BroadcastOnChannelUsingPrivateKeyCalled = func(channel string, topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			require.Fail(t, "should have not been called")
		}
		fullArchiveMes.BroadcastOnChannelUsingPrivateKeyCalled = func(channel string, topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			wasCalled = true
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		facadeInstance.broadcastOnChannelUsingPrivateKeyOnMessenger(p2p.FullArchiveMessenger, channel, topic, buff, pid, privateKey)
		require.True(t, wasCalled)

		facadeInstance.broadcastOnChannelUsingPrivateKeyOnMessenger(missingMessengerType, channel, topic, buff, pid, privateKey)
	})
}

func TestNetworkMessengersFacade_SendToConnectedPeer(t *testing.T) {
	t.Parallel()

	t.Run("SendToConnectedPeer: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.SendToConnectedPeerCalled = func(topic string, buff []byte, peerID core.PeerID) error {
			wasCalled = true
			return nil
		}
		fullArchiveMes.SendToConnectedPeerCalled = func(topic string, buff []byte, peerID core.PeerID) error {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.SendToConnectedPeer(topic, buff, pid)
		require.True(t, wasCalled)
		require.NoError(t, err)
	})
	t.Run("SendToConnectedPeerOnMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.SendToConnectedPeerCalled = func(topic string, buff []byte, peerID core.PeerID) error {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.SendToConnectedPeerCalled = func(topic string, buff []byte, peerID core.PeerID) error {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.SendToConnectedPeerOnMessenger(p2p.FullArchiveMessenger, topic, buff, pid)
		require.True(t, wasCalled)
		require.NoError(t, err)

		err = facadeInstance.SendToConnectedPeerOnMessenger(missingMessengerType, topic, buff, pid)
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
	})
}

func TestNetworkMessengersFacade_UnJoinAllTopics(t *testing.T) {
	t.Parallel()

	t.Run("UnJoinAllTopics: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.UnJoinAllTopicsCalled = func() error {
			wasCalled = true
			return nil
		}
		fullArchiveMes.UnJoinAllTopicsCalled = func() error {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.UnJoinAllTopics()
		require.True(t, wasCalled)
		require.NoError(t, err)
	})
	t.Run("unJoinAllTopicsOnMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.UnJoinAllTopicsCalled = func() error {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.UnJoinAllTopicsCalled = func() error {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.unJoinAllTopicsOnMessenger(p2p.FullArchiveMessenger)
		require.True(t, wasCalled)
		require.NoError(t, err)

		err = facadeInstance.unJoinAllTopicsOnMessenger(missingMessengerType)
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
	})
}

func TestNetworkMessengersFacade_Bootstrap(t *testing.T) {
	t.Parallel()

	t.Run("Bootstrap: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.BootstrapCalled = func() error {
			wasCalled = true
			return nil
		}
		fullArchiveMes.BootstrapCalled = func() error {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.Bootstrap()
		require.True(t, wasCalled)
		require.NoError(t, err)
	})
	t.Run("BootstrapMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.BootstrapCalled = func() error {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.BootstrapCalled = func() error {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.BootstrapMessenger(p2p.FullArchiveMessenger)
		require.True(t, wasCalled)
		require.NoError(t, err)

		err = facadeInstance.BootstrapMessenger(missingMessengerType)
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
	})
}

func TestNetworkMessengersFacade_Peers(t *testing.T) {
	t.Parallel()

	t.Run("Peers: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.PeersCalled = func() []core.PeerID {
			wasCalled = true
			return nil
		}
		fullArchiveMes.PeersCalled = func() []core.PeerID {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.Peers()
		require.True(t, wasCalled)
	})
	t.Run("peersOfMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.PeersCalled = func() []core.PeerID {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.PeersCalled = func() []core.PeerID {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.peersOfMessenger(p2p.FullArchiveMessenger)
		require.True(t, wasCalled)

		_ = facadeInstance.peersOfMessenger(missingMessengerType)
	})
}

func TestNetworkMessengersFacade_Addresses(t *testing.T) {
	t.Parallel()

	t.Run("Addresses: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.AddressesCalled = func() []string {
			wasCalled = true
			return nil
		}
		fullArchiveMes.AddressesCalled = func() []string {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.Addresses()
		require.True(t, wasCalled)
	})
	t.Run("addressesOfMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.AddressesCalled = func() []string {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.AddressesCalled = func() []string {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.addressesOfMessenger(p2p.FullArchiveMessenger)
		require.True(t, wasCalled)

		_ = facadeInstance.addressesOfMessenger(missingMessengerType)
	})
}

func TestNetworkMessengersFacade_ConnectToPeer(t *testing.T) {
	t.Parallel()

	t.Run("ConnectToPeer: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.ConnectToPeerCalled = func(address string) error {
			wasCalled = true
			return nil
		}
		fullArchiveMes.ConnectToPeerCalled = func(address string) error {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.ConnectToPeer(address)
		require.True(t, wasCalled)
		require.NoError(t, err)
	})
	t.Run("connectMessengerToPeer", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.ConnectToPeerCalled = func(address string) error {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.ConnectToPeerCalled = func(address string) error {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.connectMessengerToPeer(p2p.FullArchiveMessenger, address)
		require.True(t, wasCalled)
		require.NoError(t, err)

		err = facadeInstance.connectMessengerToPeer(missingMessengerType, address)
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
	})
}

func TestNetworkMessengersFacade_IsConnected(t *testing.T) {
	t.Parallel()

	t.Run("IsConnected: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.IsConnectedCalled = func(peerID core.PeerID) bool {
			wasCalled = true
			return true
		}
		fullArchiveMes.IsConnectedCalled = func(peerID core.PeerID) bool {
			require.Fail(t, "should have not been called")
			return true
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.IsConnected(pid)
		require.True(t, wasCalled)
	})
	t.Run("isConnectedOnMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.IsConnectedCalled = func(peerID core.PeerID) bool {
			require.Fail(t, "should have not been called")
			return true
		}
		fullArchiveMes.IsConnectedCalled = func(peerID core.PeerID) bool {
			wasCalled = true
			return true
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.isConnectedOnMessenger(p2p.FullArchiveMessenger, pid)
		require.True(t, wasCalled)

		_ = facadeInstance.isConnectedOnMessenger(missingMessengerType, pid)
	})
}

func TestNetworkMessengersFacade_ConnectedPeers(t *testing.T) {
	t.Parallel()

	t.Run("ConnectedPeers: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.ConnectedPeersCalled = func() []core.PeerID {
			wasCalled = true
			return nil
		}
		fullArchiveMes.ConnectedPeersCalled = func() []core.PeerID {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.ConnectedPeers()
		require.True(t, wasCalled)
	})
	t.Run("connectedPeersOfMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.ConnectedPeersCalled = func() []core.PeerID {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.ConnectedPeersCalled = func() []core.PeerID {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.connectedPeersOfMessenger(p2p.FullArchiveMessenger)
		require.True(t, wasCalled)

		_ = facadeInstance.connectedPeersOfMessenger(missingMessengerType)
	})
}

func TestNetworkMessengersFacade_ConnectedAddresses(t *testing.T) {
	t.Parallel()

	t.Run("ConnectedAddresses: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.ConnectedAddressesCalled = func() []string {
			wasCalled = true
			return nil
		}
		fullArchiveMes.ConnectedAddressesCalled = func() []string {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.ConnectedAddresses()
		require.True(t, wasCalled)
	})
	t.Run("connectedAddressesOfMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.ConnectedAddressesCalled = func() []string {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.ConnectedAddressesCalled = func() []string {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.connectedAddressesOfMessenger(p2p.FullArchiveMessenger)
		require.True(t, wasCalled)

		_ = facadeInstance.connectedAddressesOfMessenger(missingMessengerType)
	})
}

func TestNetworkMessengersFacade_PeerAddresses(t *testing.T) {
	t.Parallel()

	t.Run("PeerAddresses: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.PeerAddressesCalled = func(pid core.PeerID) []string {
			wasCalled = true
			return nil
		}
		fullArchiveMes.PeerAddressesCalled = func(pid core.PeerID) []string {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.PeerAddresses(pid)
		require.True(t, wasCalled)
	})
	t.Run("peerAddressesOfMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.PeerAddressesCalled = func(pid core.PeerID) []string {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.PeerAddressesCalled = func(pid core.PeerID) []string {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.peerAddressesOfMessenger(p2p.FullArchiveMessenger, pid)
		require.True(t, wasCalled)

		_ = facadeInstance.peerAddressesOfMessenger(missingMessengerType, pid)
	})
}

func TestNetworkMessengersFacade_ConnectedPeersOnTopic(t *testing.T) {
	t.Parallel()

	t.Run("ConnectedPeersOnTopic: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.ConnectedPeersOnTopicCalled = func(topic string) []core.PeerID {
			wasCalled = true
			return nil
		}
		fullArchiveMes.ConnectedPeersOnTopicCalled = func(topic string) []core.PeerID {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.ConnectedPeersOnTopic(topic)
		require.True(t, wasCalled)
	})
	t.Run("ConnectedPeersOnTopicForMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.ConnectedPeersOnTopicCalled = func(topic string) []core.PeerID {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.ConnectedPeersOnTopicCalled = func(topic string) []core.PeerID {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.ConnectedPeersOnTopicForMessenger(p2p.FullArchiveMessenger, topic)
		require.True(t, wasCalled)

		_ = facadeInstance.ConnectedPeersOnTopicForMessenger(missingMessengerType, topic)
	})
}

func TestNetworkMessengersFacade_SetPeerShardResolver(t *testing.T) {
	t.Parallel()

	t.Run("SetPeerShardResolver: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.SetPeerShardResolverCalled = func(peerShardResolver p2p.PeerShardResolver) error {
			wasCalled = true
			return nil
		}
		fullArchiveMes.SetPeerShardResolverCalled = func(peerShardResolver p2p.PeerShardResolver) error {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.SetPeerShardResolver(&mock.PeerShardResolverStub{})
		require.True(t, wasCalled)
		require.NoError(t, err)
	})
	t.Run("SetPeerShardResolverForMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.SetPeerShardResolverCalled = func(peerShardResolver p2p.PeerShardResolver) error {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.SetPeerShardResolverCalled = func(peerShardResolver p2p.PeerShardResolver) error {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.SetPeerShardResolverForMessenger(p2p.FullArchiveMessenger, &mock.PeerShardResolverStub{})
		require.True(t, wasCalled)
		require.NoError(t, err)

		err = facadeInstance.SetPeerShardResolverForMessenger(missingMessengerType, &mock.PeerShardResolverStub{})
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
	})
}

func TestNetworkMessengersFacade_GetConnectedPeersInfo(t *testing.T) {
	t.Parallel()

	t.Run("GetConnectedPeersInfo: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.GetConnectedPeersInfoCalled = func() *p2p.ConnectedPeersInfo {
			wasCalled = true
			return nil
		}
		fullArchiveMes.GetConnectedPeersInfoCalled = func() *p2p.ConnectedPeersInfo {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.GetConnectedPeersInfo()
		require.True(t, wasCalled)
	})
	t.Run("GetConnectedPeersInfoForMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.GetConnectedPeersInfoCalled = func() *p2p.ConnectedPeersInfo {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.GetConnectedPeersInfoCalled = func() *p2p.ConnectedPeersInfo {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.GetConnectedPeersInfoForMessenger(p2p.FullArchiveMessenger)
		require.True(t, wasCalled)

		_ = facadeInstance.GetConnectedPeersInfoForMessenger(missingMessengerType)
	})
}

func TestNetworkMessengersFacade_WaitForConnections(t *testing.T) {
	t.Parallel()

	t.Run("WaitForConnections: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.WaitForConnectionsCalled = func(maxWaitingTime time.Duration, minNumOfPeers uint32) {
			wasCalled = true
		}
		fullArchiveMes.WaitForConnectionsCalled = func(maxWaitingTime time.Duration, minNumOfPeers uint32) {
			require.Fail(t, "should have not been called")
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		facadeInstance.WaitForConnections(time.Second, 1)
		require.True(t, wasCalled)
	})
	t.Run("WaitForConnectionsOnMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.WaitForConnectionsCalled = func(maxWaitingTime time.Duration, minNumOfPeers uint32) {
			require.Fail(t, "should have not been called")
		}
		fullArchiveMes.WaitForConnectionsCalled = func(maxWaitingTime time.Duration, minNumOfPeers uint32) {
			wasCalled = true
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		facadeInstance.WaitForConnectionsOnMessenger(p2p.FullArchiveMessenger, time.Second, 1)
		require.True(t, wasCalled)

		facadeInstance.WaitForConnectionsOnMessenger(missingMessengerType, time.Second, 1)
	})
}

func TestNetworkMessengersFacade_IsConnectedToTheNetwork(t *testing.T) {
	t.Parallel()

	t.Run("IsConnectedToTheNetwork: should work", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.IsConnectedToTheNetworkCalled = func() bool {
			wasCalled = true
			return true
		}
		fullArchiveMes.IsConnectedToTheNetworkCalled = func() bool {
			require.Fail(t, "should have not been called")
			return true
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.IsConnectedToTheNetwork()
		require.True(t, wasCalled)
	})
	t.Run("IsConnectedToTheNetwork: missing regular messenger should not call", func(t *testing.T) {
		t.Parallel()

		facadeInstance, err := NewNetworkMessengersFacade(&mock.MessengerStub{
			IsConnectedToTheNetworkCalled: func() bool {
				require.Fail(t, "should have not been called")
				return true
			},
			TypeCalled: func() p2p.NetworkMessengerType {
				return "not regular"
			},
		})
		require.NoError(t, err)

		_ = facadeInstance.IsConnectedToTheNetwork()
	})
}

func TestNetworkMessengersFacade_ThresholdMinConnectedPeers(t *testing.T) {
	t.Parallel()

	t.Run("ThresholdMinConnectedPeers: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.ThresholdMinConnectedPeersCalled = func() int {
			wasCalled = true
			return 1
		}
		fullArchiveMes.ThresholdMinConnectedPeersCalled = func() int {
			require.Fail(t, "should have not been called")
			return 1
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		threshold := facadeInstance.ThresholdMinConnectedPeers()
		require.True(t, wasCalled)
		require.Equal(t, 1, threshold)
	})
	t.Run("thresholdMinConnectedPeersOnMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.ThresholdMinConnectedPeersCalled = func() int {
			require.Fail(t, "should have not been called")
			return 1
		}
		fullArchiveMes.ThresholdMinConnectedPeersCalled = func() int {
			wasCalled = true
			return 1
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		threshold := facadeInstance.thresholdMinConnectedPeersOnMessenger(p2p.FullArchiveMessenger)
		require.True(t, wasCalled)
		require.Equal(t, 1, threshold)

		threshold = facadeInstance.thresholdMinConnectedPeersOnMessenger(missingMessengerType)
		require.Equal(t, 0, threshold)
	})
}

func TestNetworkMessengersFacade_SetThresholdMinConnectedPeers(t *testing.T) {
	t.Parallel()

	t.Run("SetThresholdMinConnectedPeers: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.SetThresholdMinConnectedPeersCalled = func(minConnectedPeers int) error {
			wasCalled = true
			return nil
		}
		fullArchiveMes.SetThresholdMinConnectedPeersCalled = func(minConnectedPeers int) error {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.SetThresholdMinConnectedPeers(10)
		require.True(t, wasCalled)
		require.NoError(t, err)
	})
	t.Run("setThresholdMinConnectedPeersOnMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.SetThresholdMinConnectedPeersCalled = func(minConnectedPeers int) error {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.SetThresholdMinConnectedPeersCalled = func(minConnectedPeers int) error {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.setThresholdMinConnectedPeersOnMessenger(p2p.FullArchiveMessenger, 10)
		require.True(t, wasCalled)
		require.NoError(t, err)

		err = facadeInstance.setThresholdMinConnectedPeersOnMessenger(missingMessengerType, 10)
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
	})
}

func TestNetworkMessengersFacade_SetPeerDenialEvaluator(t *testing.T) {
	t.Parallel()

	t.Run("SetPeerDenialEvaluator: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.SetPeerDenialEvaluatorCalled = func(handler p2p.PeerDenialEvaluator) error {
			wasCalled = true
			return nil
		}
		fullArchiveMes.SetPeerDenialEvaluatorCalled = func(handler p2p.PeerDenialEvaluator) error {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.SetPeerDenialEvaluator(&mock.PeerDenialEvaluatorStub{})
		require.True(t, wasCalled)
		require.NoError(t, err)
	})
	t.Run("setPeerDenialEvaluatorOnMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.SetPeerDenialEvaluatorCalled = func(handler p2p.PeerDenialEvaluator) error {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.SetPeerDenialEvaluatorCalled = func(handler p2p.PeerDenialEvaluator) error {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.setPeerDenialEvaluatorOnMessenger(p2p.FullArchiveMessenger, &mock.PeerDenialEvaluatorStub{})
		require.True(t, wasCalled)
		require.NoError(t, err)

		err = facadeInstance.setPeerDenialEvaluatorOnMessenger(missingMessengerType, &mock.PeerDenialEvaluatorStub{})
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
	})
}

func TestNetworkMessengersFacade_ID(t *testing.T) {
	t.Parallel()

	t.Run("ID: should work", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.IDCalled = func() core.PeerID {
			wasCalled = true
			return ""
		}
		fullArchiveMes.IDCalled = func() core.PeerID {
			require.Fail(t, "should have not been called")
			return ""
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		_ = facadeInstance.ID()
		require.True(t, wasCalled)
	})
	t.Run("ID: missing regular messenger should not call", func(t *testing.T) {
		t.Parallel()

		facadeInstance, err := NewNetworkMessengersFacade(&mock.MessengerStub{
			IDCalled: func() core.PeerID {
				require.Fail(t, "should have not been called")
				return ""
			},
			TypeCalled: func() p2p.NetworkMessengerType {
				return "not regular"
			},
		})
		require.NoError(t, err)

		_ = facadeInstance.ID()
	})
}

func TestNetworkMessengersFacade_Port(t *testing.T) {
	t.Parallel()

	t.Run("Port: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.PortCalled = func() int {
			wasCalled = true
			return 1234
		}
		fullArchiveMes.PortCalled = func() int {
			require.Fail(t, "should have not been called")
			return 1234
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		port := facadeInstance.Port()
		require.True(t, wasCalled)
		require.Equal(t, 1234, port)
	})
	t.Run("portOfMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.PortCalled = func() int {
			require.Fail(t, "should have not been called")
			return 1234
		}
		fullArchiveMes.PortCalled = func() int {
			wasCalled = true
			return 1234
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		port := facadeInstance.portOfMessenger(p2p.FullArchiveMessenger)
		require.True(t, wasCalled)
		require.Equal(t, 1234, port)

		port = facadeInstance.portOfMessenger(missingMessengerType)
		require.Equal(t, 0, port)
	})
}

func TestNetworkMessengersFacade_Sign(t *testing.T) {
	t.Parallel()

	t.Run("Sign: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.SignCalled = func(payload []byte) ([]byte, error) {
			wasCalled = true
			return signature, nil
		}
		fullArchiveMes.SignCalled = func(payload []byte) ([]byte, error) {
			require.Fail(t, "should have not been called")
			return signature, nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		sig, err := facadeInstance.Sign(payload)
		require.True(t, wasCalled)
		require.NoError(t, err)
		require.Equal(t, signature, sig)
	})
	t.Run("Sign: missing regular messenger should return error", func(t *testing.T) {
		t.Parallel()

		facadeInstance, err := NewNetworkMessengersFacade(&mock.MessengerStub{
			SignCalled: func(payload []byte) ([]byte, error) {
				require.Fail(t, "should have not been called")
				return signature, nil
			},
			TypeCalled: func() p2p.NetworkMessengerType {
				return "not regular"
			},
		})
		require.NoError(t, err)

		sig, err := facadeInstance.Sign(payload)
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
		require.Nil(t, sig)
	})
}

func TestNetworkMessengersFacade_Verify(t *testing.T) {
	t.Parallel()

	t.Run("Verify: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.VerifyCalled = func(payload []byte, pid core.PeerID, signature []byte) error {
			wasCalled = true
			return nil
		}
		fullArchiveMes.VerifyCalled = func(payload []byte, pid core.PeerID, signature []byte) error {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.Verify(payload, pid, signature)
		require.True(t, wasCalled)
		require.NoError(t, err)
	})
	t.Run("Verify: missing regular messenger should return error", func(t *testing.T) {
		t.Parallel()

		facadeInstance, err := NewNetworkMessengersFacade(&mock.MessengerStub{
			VerifyCalled: func(payload []byte, pid core.PeerID, signature []byte) error {
				require.Fail(t, "should have not been called")
				return nil
			},
			TypeCalled: func() p2p.NetworkMessengerType {
				return "not regular"
			},
		})
		require.NoError(t, err)

		err = facadeInstance.Verify(payload, pid, signature)
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
	})
}

func TestNetworkMessengersFacade_SignUsingPrivateKey(t *testing.T) {
	t.Parallel()

	t.Run("SignUsingPrivateKey: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.SignUsingPrivateKeyCalled = func(skBytes []byte, payload []byte) ([]byte, error) {
			wasCalled = true
			return signature, nil
		}
		fullArchiveMes.SignUsingPrivateKeyCalled = func(skBytes []byte, payload []byte) ([]byte, error) {
			require.Fail(t, "should have not been called")
			return signature, nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		sig, err := facadeInstance.SignUsingPrivateKey(privateKey, payload)
		require.True(t, wasCalled)
		require.NoError(t, err)
		require.Equal(t, signature, sig)
	})
	t.Run("SignUsingPrivateKey: missing regular messenger should return error", func(t *testing.T) {
		t.Parallel()

		facadeInstance, err := NewNetworkMessengersFacade(&mock.MessengerStub{
			SignUsingPrivateKeyCalled: func(skBytes []byte, payload []byte) ([]byte, error) {
				require.Fail(t, "should have not been called")
				return signature, nil
			},
			TypeCalled: func() p2p.NetworkMessengerType {
				return "not regular"
			},
		})
		require.NoError(t, err)

		sig, err := facadeInstance.SignUsingPrivateKey(privateKey, payload)
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
		require.Nil(t, sig)
	})
}

func TestNetworkMessengersFacade_AddPeerTopicNotifier(t *testing.T) {
	t.Parallel()

	t.Run("AddPeerTopicNotifier: should call the regular messenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.AddPeerTopicNotifierCalled = func(notifier p2p.PeerTopicNotifier) error {
			wasCalled = true
			return nil
		}
		fullArchiveMes.AddPeerTopicNotifierCalled = func(notifier p2p.PeerTopicNotifier) error {
			require.Fail(t, "should have not been called")
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.AddPeerTopicNotifier(&mock.PeerTopicNotifierStub{})
		require.True(t, wasCalled)
		require.NoError(t, err)
	})
	t.Run("addPeerTopicNotifierOnMessenger", func(t *testing.T) {
		t.Parallel()

		regularMes, fullArchiveMes := createMessengersStubs()
		wasCalled := false
		regularMes.AddPeerTopicNotifierCalled = func(notifier p2p.PeerTopicNotifier) error {
			require.Fail(t, "should have not been called")
			return nil
		}
		fullArchiveMes.AddPeerTopicNotifierCalled = func(notifier p2p.PeerTopicNotifier) error {
			wasCalled = true
			return nil
		}

		facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
		require.NoError(t, err)

		err = facadeInstance.addPeerTopicNotifierOnMessenger(p2p.FullArchiveMessenger, &mock.PeerTopicNotifierStub{})
		require.True(t, wasCalled)
		require.NoError(t, err)

		err = facadeInstance.addPeerTopicNotifierOnMessenger(missingMessengerType, &mock.PeerTopicNotifierStub{})
		require.True(t, errors.Is(err, p2p.ErrUnknownMessenger))
	})
}

func TestNetworkMessengersFacade_Type(t *testing.T) {
	t.Parallel()

	facadeInstance, err := NewNetworkMessengersFacade(&mock.MessengerStub{})
	require.NoError(t, err)

	facadeType := facadeInstance.Type()
	require.Equal(t, "Facade", string(facadeType))
}

func TestNetworkMessengersFacade_Close(t *testing.T) {
	t.Parallel()

	regularMes, fullArchiveMes := createMessengersStubs()
	expectedErr := errors.New("expected error")
	cnt := uint32(0)
	regularMes.CloseCalled = func() error {
		cnt++
		return expectedErr
	}
	fullArchiveMes.CloseCalled = func() error {
		cnt++
		return nil
	}
	facadeInstance, err := NewNetworkMessengersFacade(regularMes, fullArchiveMes)
	require.NoError(t, err)

	err = facadeInstance.Close()
	require.True(t, errors.Is(err, expectedErr))
	require.Equal(t, uint32(2), cnt)
}

func TestNetworkMessengersFacade_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var facadeInstance *networkMessengersFacade
	require.True(t, facadeInstance.IsInterfaceNil())

	facadeInstance, _ = NewNetworkMessengersFacade(&mock.MessengerStub{})
	require.False(t, facadeInstance.IsInterfaceNil())
}
