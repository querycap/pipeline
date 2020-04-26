package mem

import (
	"context"
	"sync"

	"github.com/querycap/pipeline/pkg/eventbus"
)

func NewMemEventBus() eventbus.EventBus {
	return &MemEventBus{
		handlers: map[string]map[int]eventbus.Handler{},
	}
}

type MemEventBus struct {
	rw       sync.RWMutex
	handlers map[string]map[int]eventbus.Handler
	incr     int
}

func (m *MemEventBus) Publish(ctx context.Context, topic string, data []byte) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	handlers, ok := m.handlers[topic]
	if !ok {
		return eventbus.ErrNoSubscriptionsForTopic
	}

	if len(handlers) == 0 {
		return eventbus.ErrNoSubscriptionsForTopic
	}

	for i := range handlers {
		handler := handlers[i]
		go func() {
			handler(ctx, data)
		}()
	}

	return nil
}

func (m *MemEventBus) Subscribe(topic string, handler eventbus.Handler) eventbus.Subscription {
	m.rw.Lock()
	defer m.rw.Unlock()

	if m.handlers[topic] == nil {
		m.handlers[topic] = map[int]eventbus.Handler{}
	}

	i := m.incr

	m.handlers[topic][i] = handler
	m.incr++

	return eventbus.NewSubscription(func() {
		m.rw.Lock()
		defer m.rw.Unlock()

		handlers, ok := m.handlers[topic]
		if ok {
			delete(handlers, i)
		}
	})
}
