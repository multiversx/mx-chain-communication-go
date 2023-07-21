package libp2p_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/stretchr/testify/assert"
)

var errInvalidType = errors.New("invalid type")
var errLenDifferent = errors.New("len different for names and chans")
var errLenDifferentNamesChans = errors.New("len different for names and chans")
var errMissingChannel = errors.New("missing channel")
var errChannelsMismatch = errors.New("channels mismatch")
var durationWait = time.Second * 2

func checkIntegrity(oclbInstance libp2p.ChannelLoadBalancer, name string) error {
	type x interface {
		Chans() []chan *libp2p.SendableData
		Names() []string
		NamesChans() map[string]chan *libp2p.SendableData
	}

	oclb, ok := oclbInstance.(x)
	if !ok {
		return errInvalidType
	}

	if len(oclb.Names()) != len(oclb.Chans()) {
		return errLenDifferent
	}

	if len(oclb.Names()) != len(oclb.NamesChans()) {
		return errLenDifferentNamesChans
	}

	idxFound := -1
	for i, n := range oclb.Names() {
		if n == name {
			idxFound = i
			break
		}
	}

	if idxFound == -1 && oclb.NamesChans()[name] == nil {
		return errMissingChannel
	}

	if oclb.NamesChans()[name] != oclb.Chans()[idxFound] {
		return errChannelsMismatch
	}

	return nil
}

func TestNewOutgoingChannelLoadBalancer(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		oclb, err := libp2p.NewOutgoingChannelLoadBalancer(nil)
		assert.Equal(t, p2p.ErrNilLogger, err)
		assert.Nil(t, oclb)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		oclb, err := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})
		assert.Nil(t, err)
		assert.NotNil(t, oclb)
	})
	t.Run("should work and add default channel", func(t *testing.T) {
		t.Parallel()

		oclb, err := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})
		assert.Nil(t, err)
		assert.NotNil(t, oclb)

		assert.Equal(t, 1, len(oclb.Names()))
		assert.Nil(t, checkIntegrity(oclb, libp2p.DefaultSendChannel()))
	})
}

//------- AddChannel

func TestOutgoingChannelLoadBalancer_AddChannelNewChannelShouldNotErrAndAddNewChannel(t *testing.T) {
	t.Parallel()

	oclb, _ := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})

	err := oclb.AddChannel("test")

	assert.Nil(t, err)
	assert.Equal(t, 2, len(oclb.Names()))
	assert.Nil(t, checkIntegrity(oclb, libp2p.DefaultSendChannel()))
	assert.Nil(t, checkIntegrity(oclb, "test"))
}

func TestOutgoingChannelLoadBalancer_AddChannelDefaultChannelShouldErr(t *testing.T) {
	t.Parallel()

	oclb, _ := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})

	err := oclb.AddChannel(libp2p.DefaultSendChannel())

	assert.Equal(t, p2p.ErrChannelCanNotBeReAdded, err)
}

func TestOutgoingChannelLoadBalancer_AddChannelReAddChannelShouldDoNothing(t *testing.T) {
	t.Parallel()

	oclb, _ := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})

	_ = oclb.AddChannel("test")
	err := oclb.AddChannel("test")

	assert.Nil(t, err)
	assert.Equal(t, 2, len(oclb.Chans()))
}

//------- RemoveChannel

func TestOutgoingChannelLoadBalancer_RemoveChannelRemoveDefaultShouldErr(t *testing.T) {
	t.Parallel()

	oclb, _ := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})

	err := oclb.RemoveChannel(libp2p.DefaultSendChannel())

	assert.Equal(t, p2p.ErrChannelCanNotBeDeleted, err)
}

func TestOutgoingChannelLoadBalancer_RemoveChannelRemoveNotFoundChannelShouldErr(t *testing.T) {
	t.Parallel()

	oclb, _ := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})

	err := oclb.RemoveChannel("test")

	assert.Equal(t, p2p.ErrChannelDoesNotExist, err)
}

func TestOutgoingChannelLoadBalancer_RemoveChannelRemoveLastChannelAddedShouldWork(t *testing.T) {
	t.Parallel()

	oclb, _ := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})

	_ = oclb.AddChannel("test1")
	_ = oclb.AddChannel("test2")
	_ = oclb.AddChannel("test3")

	err := oclb.RemoveChannel("test3")

	assert.Nil(t, err)

	assert.Equal(t, 3, len(oclb.Names()))
	assert.Nil(t, checkIntegrity(oclb, libp2p.DefaultSendChannel()))
	assert.Nil(t, checkIntegrity(oclb, "test1"))
	assert.Nil(t, checkIntegrity(oclb, "test2"))
	assert.Equal(t, errMissingChannel, checkIntegrity(oclb, "test3"))
}

func TestOutgoingChannelLoadBalancer_RemoveChannelRemoveFirstChannelAddedShouldWork(t *testing.T) {
	t.Parallel()

	oclb, _ := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})

	_ = oclb.AddChannel("test1")
	_ = oclb.AddChannel("test2")
	_ = oclb.AddChannel("test3")

	err := oclb.RemoveChannel("test1")

	assert.Nil(t, err)

	assert.Equal(t, 3, len(oclb.Names()))
	assert.Nil(t, checkIntegrity(oclb, libp2p.DefaultSendChannel()))
	assert.Equal(t, errMissingChannel, checkIntegrity(oclb, "test1"))
	assert.Nil(t, checkIntegrity(oclb, "test2"))
	assert.Nil(t, checkIntegrity(oclb, "test3"))
}

func TestOutgoingChannelLoadBalancer_RemoveChannelRemoveMiddleChannelAddedShouldWork(t *testing.T) {
	t.Parallel()

	oclb, _ := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})

	_ = oclb.AddChannel("test1")
	_ = oclb.AddChannel("test2")
	_ = oclb.AddChannel("test3")

	err := oclb.RemoveChannel("test2")

	assert.Nil(t, err)

	assert.Equal(t, 3, len(oclb.Names()))
	assert.Nil(t, checkIntegrity(oclb, libp2p.DefaultSendChannel()))
	assert.Nil(t, checkIntegrity(oclb, "test1"))
	assert.Equal(t, errMissingChannel, checkIntegrity(oclb, "test2"))
	assert.Nil(t, checkIntegrity(oclb, "test3"))
}

//------- GetChannelOrDefault

func TestOutgoingChannelLoadBalancer_GetChannelOrDefaultNotFoundShouldReturnDefault(t *testing.T) {
	t.Parallel()

	oclb, _ := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})

	_ = oclb.AddChannel("test1")

	channel := oclb.GetChannelOrDefault("missing channel")

	assert.True(t, oclb.NamesChans()[libp2p.DefaultSendChannel()] == channel)
}

func TestOutgoingChannelLoadBalancer_GetChannelOrDefaultFoundShouldReturnChannel(t *testing.T) {
	t.Parallel()

	oclb, _ := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})

	_ = oclb.AddChannel("test1")

	channel := oclb.GetChannelOrDefault("test1")

	assert.True(t, oclb.NamesChans()["test1"] == channel)
}

//------- CollectOneElementFromChannels

func TestOutgoingChannelLoadBalancer_CollectFromChannelsNoObjectsShouldWaitBlocking(t *testing.T) {
	t.Parallel()

	oclb, _ := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})

	chanDone := make(chan struct{})

	go func() {
		_ = oclb.CollectOneElementFromChannels()

		chanDone <- struct{}{}
	}()

	select {
	case <-chanDone:
		assert.Fail(t, "should have not received object")
	case <-time.After(durationWait):
	}
}

func TestOutgoingChannelLoadBalancer_CollectOneElementFromChannelsShouldWork(t *testing.T) {
	t.Parallel()

	oclb, _ := libp2p.NewOutgoingChannelLoadBalancer(&testscommon.LoggerStub{})

	_ = oclb.AddChannel("test")

	obj1 := &libp2p.SendableData{Topic: "test"}
	obj2 := &libp2p.SendableData{Topic: "default"}

	chanDone := make(chan bool)
	wg := sync.WaitGroup{}
	wg.Add(3)

	//send on channel test
	go func() {
		oclb.GetChannelOrDefault("test") <- obj1
		wg.Done()
	}()

	//send on default channel
	go func() {
		oclb.GetChannelOrDefault(libp2p.DefaultSendChannel()) <- obj2
		wg.Done()
	}()

	//func to wait finishing sending and receiving
	go func() {
		wg.Wait()
		chanDone <- true
	}()

	//func to periodically consume from channels
	go func() {
		foundObj1 := false
		foundObj2 := false

		for {
			obj := oclb.CollectOneElementFromChannels()

			if !foundObj1 {
				if obj == obj1 {
					foundObj1 = true
				}
			}

			if !foundObj2 {
				if obj == obj2 {
					foundObj2 = true
				}
			}

			if foundObj1 && foundObj2 {
				break
			}
		}

		wg.Done()
	}()

	select {
	case <-chanDone:
		return
	case <-time.After(durationWait):
		assert.Fail(t, "timeout")
		return
	}
}
