package mock

import p2p "github.com/multiversx/mx-chain-p2p-go"

// TopicProcessorStub -
type TopicProcessorStub struct {
	AddTopicProcessorCalled    func(identifier string, processor p2p.MessageProcessor) error
	RemoveTopicProcessorCalled func(identifier string) error
	GetListCalled              func() ([]string, []p2p.MessageProcessor)
}

// AddTopicProcessor -
func (stub *TopicProcessorStub) AddTopicProcessor(identifier string, processor p2p.MessageProcessor) error {
	if stub.AddTopicProcessorCalled != nil {
		return stub.AddTopicProcessorCalled(identifier, processor)
	}
	return nil
}

// RemoveTopicProcessor -
func (stub *TopicProcessorStub) RemoveTopicProcessor(identifier string) error {
	if stub.RemoveTopicProcessorCalled != nil {
		return stub.RemoveTopicProcessorCalled(identifier)
	}
	return nil
}

// GetList -
func (stub *TopicProcessorStub) GetList() ([]string, []p2p.MessageProcessor) {
	if stub.GetListCalled != nil {
		return stub.GetListCalled()
	}
	return nil, nil
}

// IsInterfaceNil -
func (stub *TopicProcessorStub) IsInterfaceNil() bool {
	return stub == nil
}
