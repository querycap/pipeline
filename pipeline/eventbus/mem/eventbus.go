package mem

import (
	"context"
	"sync"

	"github.com/querycap/pipeline/pipeline"
)

func NewMemEventBus() pipeline.EventBus {
	return &MemEventBus{
		handlers: map[string]map[int]pipeline.Handler{},
	}
}

type MemEventBus struct {
	rw       sync.RWMutex
	handlers map[string]map[int]pipeline.Handler
	incr     int
}

func (m *MemEventBus) Publish(ctx context.Context, topic string, data []byte) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	handlers, ok := m.handlers[topic]
	if !ok {
		return pipeline.ErrNoSubscriptionsForTopic
	}

	if len(handlers) == 0 {
		return pipeline.ErrNoSubscriptionsForTopic
	}

	for i := range handlers {
		handler := handlers[i]
		go func() {
			handler(ctx, data)
		}()
	}

	return nil
}

func (m *MemEventBus) Subscribe(topic string, handler pipeline.Handler) pipeline.Subscription {
	m.rw.Lock()
	defer m.rw.Unlock()

	if m.handlers[topic] == nil {
		m.handlers[topic] = map[int]pipeline.Handler{}
	}

	i := m.incr

	m.handlers[topic][i] = handler
	m.incr++

	return pipeline.NewSubscription(func() {
		m.rw.Lock()
		defer m.rw.Unlock()

		handlers, ok := m.handlers[topic]
		if ok {
			delete(handlers, i)
		}
	})
}
