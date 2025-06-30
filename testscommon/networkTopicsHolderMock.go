package testscommon

import "github.com/multiversx/mx-chain-communication-go/p2p"

// NetworkTopicsHolderMock -
type NetworkTopicsHolderMock struct {
	AddTopicOnNetworkIfNeededCalled func(networkType p2p.NetworkType, topic string)
	GetNetworkTypeForTopicCalled    func(topic string) p2p.NetworkType
	RemoveTopicCalled               func(topic string)
}

// AddTopicOnNetworkIfNeeded -
func (mock *NetworkTopicsHolderMock) AddTopicOnNetworkIfNeeded(networkType p2p.NetworkType, topic string) {
	if mock.AddTopicOnNetworkIfNeededCalled != nil {
		mock.AddTopicOnNetworkIfNeededCalled(networkType, topic)
	}
}

// GetNetworkTypeForTopic -
func (mock *NetworkTopicsHolderMock) GetNetworkTypeForTopic(topic string) p2p.NetworkType {
	if mock.GetNetworkTypeForTopicCalled != nil {
		return mock.GetNetworkTypeForTopicCalled(topic)
	}

	return "main"
}

// RemoveTopic -
func (mock *NetworkTopicsHolderMock) RemoveTopic(topic string) {
	if mock.RemoveTopicCalled != nil {
		mock.RemoveTopicCalled(topic)
	}
}

// IsInterfaceNil -
func (mock *NetworkTopicsHolderMock) IsInterfaceNil() bool {
	return mock == nil
}
