package pubsub

import (
	"context"
	"errors"
)

// Handler processes a single message.
type Handler func(ctx context.Context, data []byte, attrs map[string]string) error

// Subscriber pulls messages from a subscription.
type Subscriber interface {
	Subscribe(ctx context.Context, subscription string, h Handler) error
}

// MockSubscriber runs h for n message iterations, used in tests.
type MockSubscriber struct {
	N int
}

// Subscribe calls h N times synchronously.
func (m *MockSubscriber) Subscribe(_ context.Context, subscription string, h Handler) error {
	if subscription == "" {
		return errors.New("pubsub: empty subscription")
	}
	if h == nil {
		return errors.New("pubsub: nil handler")
	}
	for i := 0; i < m.N; i++ {
		if err := h(context.Background(), []byte(`{"n":0}`), nil); err != nil {
			return err
		}
	}
	return nil
}

// RealSubscriber wraps google.golang.org/cloud/pubsub; not implemented in scaffold.
type RealSubscriber struct {
	Project string
}

// Subscribe unimplemented.
func (r *RealSubscriber) Subscribe(_ context.Context, _ string, _ Handler) error {
	return errors.New("pubsub: RealSubscriber.Subscribe not implemented")
}
