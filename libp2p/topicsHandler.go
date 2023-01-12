package libp2p

import (
	"sync"

	"github.com/libp2p/go-libp2p-pubsub"
)

// TODO[Sorin]: add unit tests
type topicsHandler struct {
	mutTopics     sync.RWMutex
	processors    map[string]TopicProcessor
	topics        map[string]*pubsub.Topic
	subscriptions map[string]*pubsub.Subscription
}

// NewTopicsHandler creates a new topicsHandler instance
func NewTopicsHandler() *topicsHandler {
	return &topicsHandler{
		processors:    make(map[string]TopicProcessor),
		topics:        make(map[string]*pubsub.Topic),
		subscriptions: make(map[string]*pubsub.Subscription),
	}
}

// GetTopic returns the pubsub topic for the given topic
func (handler *topicsHandler) GetTopic(topic string) *pubsub.Topic {
	handler.mutTopics.RLock()
	defer handler.mutTopics.RUnlock()

	return handler.topics[topic]
}

// HasTopic returns true if the topic is already known
func (handler *topicsHandler) HasTopic(topic string) bool {
	handler.mutTopics.RLock()
	defer handler.mutTopics.RUnlock()

	_, ok := handler.topics[topic]
	return ok
}

// AddTopic adds a new topic to the handler
func (handler *topicsHandler) AddTopic(topic string, pubSubTopic *pubsub.Topic) {
	handler.mutTopics.Lock()
	defer handler.mutTopics.Unlock()

	handler.topics[topic] = pubSubTopic
}

// RemoveTopic removes the topic from the handler
func (handler *topicsHandler) RemoveTopic(topic string) {
	handler.mutTopics.Lock()
	defer handler.mutTopics.Unlock()

	delete(handler.topics, topic)
}

// GetAllTopics returns all topics
func (handler *topicsHandler) GetAllTopics() map[string]*pubsub.Topic {
	handler.mutTopics.Lock()
	defer handler.mutTopics.Unlock()

	return handler.topics
}

// GetTopicProcessors returns the topic processors for the given topic
func (handler *topicsHandler) GetTopicProcessors(topic string) TopicProcessor {
	handler.mutTopics.RLock()
	defer handler.mutTopics.RUnlock()

	return handler.processors[topic]
}

// AddNewTopicProcessors adds a new topic processors for the given topic, returning it
func (handler *topicsHandler) AddNewTopicProcessors(topic string) TopicProcessor {
	handler.mutTopics.Lock()
	defer handler.mutTopics.Unlock()

	topicProcs := newTopicProcessors()
	handler.processors[topic] = topicProcs
	return topicProcs
}

// RemoveTopicProcessors removes the topic processors for the given topic
func (handler *topicsHandler) RemoveTopicProcessors(topic string) {
	handler.mutTopics.Lock()
	defer handler.mutTopics.Unlock()

	delete(handler.processors, topic)
}

// GetAllTopicsProcessors returns all topic processors for all topics
func (handler *topicsHandler) GetAllTopicsProcessors() map[string]TopicProcessor {
	handler.mutTopics.Lock()
	defer handler.mutTopics.Unlock()

	return handler.processors
}

// GetSubscription returns the topic subscription for the given topic
func (handler *topicsHandler) GetSubscription(topic string) *pubsub.Subscription {
	handler.mutTopics.Lock()
	defer handler.mutTopics.Unlock()

	return handler.subscriptions[topic]
}

// AddSubscription adds a new topic subscription for the given topic
func (handler *topicsHandler) AddSubscription(topic string, sub *pubsub.Subscription) {
	handler.mutTopics.Lock()
	defer handler.mutTopics.Unlock()

	handler.subscriptions[topic] = sub
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *topicsHandler) IsInterfaceNil() bool {
	return handler == nil
}
