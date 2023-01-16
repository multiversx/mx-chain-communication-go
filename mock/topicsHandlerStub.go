package mock

import "github.com/ElrondNetwork/elrond-go-p2p/libp2p"

// TopicsHandlerStub -
type TopicsHandlerStub struct {
	GetTopicCalled               func(topic string) libp2p.PubSubTopic
	HasTopicCalled               func(topic string) bool
	AddTopicCalled               func(topic string, pubSubTopic libp2p.PubSubTopic)
	RemoveTopicCalled            func(topic string)
	GetAllTopicsCalled           func() map[string]libp2p.PubSubTopic
	GetTopicProcessorsCalled     func(topic string) libp2p.TopicProcessor
	AddNewTopicProcessorsCalled  func(topic string) libp2p.TopicProcessor
	RemoveTopicProcessorsCalled  func(topic string)
	GetAllTopicsProcessorsCalled func() map[string]libp2p.TopicProcessor
	GetSubscriptionCalled        func(topic string) libp2p.PubSubSubscription
	AddSubscriptionCalled        func(topic string, sub libp2p.PubSubSubscription)
}

// GetTopic -
func (stub *TopicsHandlerStub) GetTopic(topic string) libp2p.PubSubTopic {
	if stub.GetTopicCalled != nil {
		return stub.GetTopicCalled(topic)
	}
	return nil
}

// HasTopic -
func (stub *TopicsHandlerStub) HasTopic(topic string) bool {
	if stub.HasTopicCalled != nil {
		return stub.HasTopicCalled(topic)
	}
	return false
}

// AddTopic -
func (stub *TopicsHandlerStub) AddTopic(topic string, pubSubTopic libp2p.PubSubTopic) {
	if stub.AddTopicCalled != nil {
		stub.AddTopicCalled(topic, pubSubTopic)
	}
}

// RemoveTopic -
func (stub *TopicsHandlerStub) RemoveTopic(topic string) {
	if stub.RemoveTopicCalled != nil {
		stub.RemoveTopicCalled(topic)
	}
}

// GetAllTopics -
func (stub *TopicsHandlerStub) GetAllTopics() map[string]libp2p.PubSubTopic {
	if stub.GetAllTopicsCalled != nil {
		return stub.GetAllTopicsCalled()
	}
	return nil
}

// GetTopicProcessors -
func (stub *TopicsHandlerStub) GetTopicProcessors(topic string) libp2p.TopicProcessor {
	if stub.GetTopicProcessorsCalled != nil {
		return stub.GetTopicProcessorsCalled(topic)
	}
	return nil
}

// AddNewTopicProcessors -
func (stub *TopicsHandlerStub) AddNewTopicProcessors(topic string) libp2p.TopicProcessor {
	if stub.AddNewTopicProcessorsCalled != nil {
		return stub.AddNewTopicProcessorsCalled(topic)
	}
	return nil
}

// RemoveTopicProcessors -
func (stub *TopicsHandlerStub) RemoveTopicProcessors(topic string) {
	if stub.RemoveTopicProcessorsCalled != nil {
		stub.RemoveTopicProcessorsCalled(topic)
	}
}

// GetAllTopicsProcessors -
func (stub *TopicsHandlerStub) GetAllTopicsProcessors() map[string]libp2p.TopicProcessor {
	if stub.GetAllTopicsProcessorsCalled != nil {
		return stub.GetAllTopicsProcessorsCalled()
	}
	return nil
}

// GetSubscription -
func (stub *TopicsHandlerStub) GetSubscription(topic string) libp2p.PubSubSubscription {
	if stub.GetSubscriptionCalled != nil {
		return stub.GetSubscriptionCalled(topic)
	}
	return nil
}

// AddSubscription -
func (stub *TopicsHandlerStub) AddSubscription(topic string, sub libp2p.PubSubSubscription) {
	if stub.AddSubscriptionCalled != nil {
		stub.AddSubscriptionCalled(topic, sub)
	}
}

// IsInterfaceNil -
func (stub *TopicsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
