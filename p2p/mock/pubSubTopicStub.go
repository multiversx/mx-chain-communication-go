package mock

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// PubSubTopicStub -
type PubSubTopicStub struct {
	SubscribeCalled func(opts ...pubsub.SubOpt) (*pubsub.Subscription, error)
	PublishCalled   func(ctx context.Context, data []byte, opts ...pubsub.PubOpt) error
	CloseCalled     func() error
}

// Subscribe -
func (stub *PubSubTopicStub) Subscribe(opts ...pubsub.SubOpt) (*pubsub.Subscription, error) {
	if stub.SubscribeCalled != nil {
		return stub.SubscribeCalled(opts...)
	}
	return nil, nil
}

// Publish -
func (stub *PubSubTopicStub) Publish(ctx context.Context, data []byte, opts ...pubsub.PubOpt) error {
	if stub.PublishCalled != nil {
		return stub.PublishCalled(ctx, data, opts...)
	}
	return nil
}

// Close -
func (stub *PubSubTopicStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}
	return nil
}
