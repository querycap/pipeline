package eventbus

import (
	"context"
	"strings"
)

func EventBusWithPrefix(eventBus EventBus, prefix string) EventBus {
	return &eventBusWithPrefix{
		prefix:   prefix,
		eventBus: eventBus,
	}
}

type eventBusWithPrefix struct {
	prefix   string
	eventBus EventBus
}

func (j *eventBusWithPrefix) Publish(ctx context.Context, topic string, data []byte) error {
	return j.eventBus.Publish(ctx, concat(":", j.prefix, topic), data)
}

func (j *eventBusWithPrefix) Subscribe(topic string, callback func(ctx context.Context, data []byte)) Subscription {
	return j.eventBus.Subscribe(concat(":", j.prefix, topic), callback)
}

func NewSubscription(callback func()) Subscription {
	return &subscription{callback: callback}
}

type subscription struct {
	callback func()
}

func (s *subscription) Unsubscribe() {
	s.callback()
}

func concat(sep string, parts ...string) string {
	return strings.Join(parts, sep)
}
