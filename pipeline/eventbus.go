package pipeline

import (
	"context"
	"errors"
)

var ErrNoSubscriptionsForTopic = errors.New("no subscriptions for topic")

type Subscription interface {
	Unsubscribe()
}

type Handler = func(ctx context.Context, data []byte)

type EventBus interface {
	Publish(ctx context.Context, topic string, data []byte) error
	Subscribe(topic string, callback Handler) Subscription
}
