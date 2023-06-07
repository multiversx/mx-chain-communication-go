package libp2p

import (
	"context"
	"sync"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

var _ ChannelLoadBalancer = (*outgoingChannelLoadBalancer)(nil)

const defaultSendChannel = "default send channel"

// outgoingChannelLoadBalancer is a component that evenly balances requests to be sent
type outgoingChannelLoadBalancer struct {
	mut      sync.RWMutex
	chans    []chan *SendableData
	mainChan chan *SendableData
	names    []string
	//namesChans is defined only for performance purposes as to fast search by name
	//iteration is done directly on slices as that is used very often and is about 50x
	//faster than an iteration over a map
	namesChans map[string]chan *SendableData
	cancelFunc context.CancelFunc
	ctx        context.Context //we need the context saved here in order to call appendChannel from exported func AddChannel
	log        p2p.Logger
}

// NewOutgoingChannelLoadBalancer creates a new instance of a ChannelLoadBalancer instance
func NewOutgoingChannelLoadBalancer(logger p2p.Logger) (*outgoingChannelLoadBalancer, error) {
	if check.IfNil(logger) {
		return nil, p2p.ErrNilLogger
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	oclb := &outgoingChannelLoadBalancer{
		chans:      make([]chan *SendableData, 0),
		names:      make([]string, 0),
		namesChans: make(map[string]chan *SendableData),
		mainChan:   make(chan *SendableData),
		cancelFunc: cancelFunc,
		ctx:        ctx,
		log:        logger,
	}

	oclb.appendChannel(defaultSendChannel)

	return oclb, nil
}

func (oplb *outgoingChannelLoadBalancer) appendChannel(channel string) {
	oplb.names = append(oplb.names, channel)
	ch := make(chan *SendableData)
	oplb.chans = append(oplb.chans, ch)
	oplb.namesChans[channel] = ch

	go func() {
		for {
			var obj *SendableData

			select {
			case obj = <-ch:
			case <-oplb.ctx.Done():
				oplb.log.Debug("closing OutgoingChannelLoadBalancer's append channel go routine")
				return
			}

			oplb.mainChan <- obj
		}
	}()
}

// AddChannel adds a new channel to the throttler, if it does not exists
func (oplb *outgoingChannelLoadBalancer) AddChannel(channel string) error {
	if channel == defaultSendChannel {
		return p2p.ErrChannelCanNotBeReAdded
	}

	oplb.mut.Lock()
	defer oplb.mut.Unlock()

	_, alreadyExists := oplb.namesChans[channel]
	if alreadyExists {
		return nil
	}

	oplb.appendChannel(channel)

	return nil
}

// RemoveChannel removes an existing channel from the throttler
func (oplb *outgoingChannelLoadBalancer) RemoveChannel(channel string) error {
	if channel == defaultSendChannel {
		return p2p.ErrChannelCanNotBeDeleted
	}

	oplb.mut.Lock()
	defer oplb.mut.Unlock()

	index := -1

	for idx, name := range oplb.names {
		if name == channel {
			index = idx
			break
		}
	}

	if index == -1 {
		return p2p.ErrChannelDoesNotExist
	}

	sendableChan := oplb.chans[index]

	//remove the index-th element in the chan slice
	copy(oplb.chans[index:], oplb.chans[index+1:])
	oplb.chans[len(oplb.chans)-1] = nil
	oplb.chans = oplb.chans[:len(oplb.chans)-1]

	//remove the index-th element in the names slice
	copy(oplb.names[index:], oplb.names[index+1:])
	oplb.names = oplb.names[:len(oplb.names)-1]

	close(sendableChan)

	delete(oplb.namesChans, channel)

	return nil
}

// GetChannelOrDefault fetches the required channel or the default if the channel is not present
func (oplb *outgoingChannelLoadBalancer) GetChannelOrDefault(channel string) chan *SendableData {
	oplb.mut.RLock()
	defer oplb.mut.RUnlock()

	ch := oplb.namesChans[channel]
	if ch != nil {
		return ch
	}

	return oplb.chans[0]
}

// CollectOneElementFromChannels gets the waiting object from mainChan. It is a blocking call.
func (oplb *outgoingChannelLoadBalancer) CollectOneElementFromChannels() *SendableData {
	select {
	case obj := <-oplb.mainChan:
		return obj
	case <-oplb.ctx.Done():
		return nil
	}
}

// Close finishes all started go routines in this instance
func (oplb *outgoingChannelLoadBalancer) Close() error {
	oplb.cancelFunc()
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (oplb *outgoingChannelLoadBalancer) IsInterfaceNil() bool {
	return oplb == nil
}
