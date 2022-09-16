package mock

import (
	"github.com/ElrondNetwork/elrond-go-p2p/common"
)

// ChannelLoadBalancerStub -
type ChannelLoadBalancerStub struct {
	AddChannelCalled                    func(pipe string) error
	RemoveChannelCalled                 func(pipe string) error
	GetChannelOrDefaultCalled           func(pipe string) chan *common.SendableData
	CollectOneElementFromChannelsCalled func() *common.SendableData
	CloseCalled                         func() error
}

// AddChannel -
func (clbs *ChannelLoadBalancerStub) AddChannel(pipe string) error {
	return clbs.AddChannelCalled(pipe)
}

// RemoveChannel -
func (clbs *ChannelLoadBalancerStub) RemoveChannel(pipe string) error {
	return clbs.RemoveChannelCalled(pipe)
}

// GetChannelOrDefault -
func (clbs *ChannelLoadBalancerStub) GetChannelOrDefault(pipe string) chan *common.SendableData {
	return clbs.GetChannelOrDefaultCalled(pipe)
}

// CollectOneElementFromChannels -
func (clbs *ChannelLoadBalancerStub) CollectOneElementFromChannels() *common.SendableData {
	return clbs.CollectOneElementFromChannelsCalled()
}

// Close -
func (clbs *ChannelLoadBalancerStub) Close() error {
	if clbs.CloseCalled != nil {
		return clbs.CloseCalled()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (clbs *ChannelLoadBalancerStub) IsInterfaceNil() bool {
	return clbs == nil
}
