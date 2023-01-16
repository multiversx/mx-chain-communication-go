package mock

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// PubSubSubscriptionStub -
type PubSubSubscriptionStub struct {
	TopicCalled  func() string
	NextCalled   func(ctx context.Context) (*pubsub.Message, error)
	CancelCalled func()
}

// Topic -
func (stub *PubSubSubscriptionStub) Topic() string {
	if stub.TopicCalled != nil {
		return stub.TopicCalled()
	}
	return ""
}

// Next -
func (stub *PubSubSubscriptionStub) Next(ctx context.Context) (*pubsub.Message, error) {
	if stub.NextCalled != nil {
		return stub.NextCalled(ctx)
	}
	return nil, nil
}

// Cancel -
func (stub *PubSubSubscriptionStub) Cancel() {
	if stub.CancelCalled != nil {
		stub.CancelCalled()
	}
}
