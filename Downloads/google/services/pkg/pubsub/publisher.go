// Package pubsub wraps the Cloud Pub/Sub publisher and subscriber clients.
package pubsub

import (
	"context"
	"encoding/json"
	"errors"
)

// Publisher publishes JSON messages.
type Publisher interface {
	PublishJSON(ctx context.Context, topic string, attrs map[string]string, payload any) error
}

// MockPublisher records calls in-memory for tests.
type MockPublisher struct {
	Published map[string]int
}

// NewMockPublisher creates an empty MockPublisher.
func NewMockPublisher() *MockPublisher {
	return &MockPublisher{Published: map[string]int{}}
}

// PublishJSON no-op except incrementing a counter.
func (m *MockPublisher) PublishJSON(_ context.Context, topic string, _ map[string]string, payload any) error {
	if topic == "" {
		return errors.New("pubsub: empty topic")
	}
	if payload == nil {
		return errors.New("pubsub: nil payload")
	}
	m.Published[topic]++
	return nil
}

// RealPublisher wraps google.golang.org/cloud/pubsub; not implemented in scaffold.
type RealPublisher struct {
	Project string
}

// PublishJSON unimplemented in scaffold.
func (r *RealPublisher) PublishJSON(_ context.Context, topic string, _ map[string]string, _ any) error {
	return errors.New("pubsub: RealPublisher.PublishJSON not implemented")
}

// Helpers ----

// MarshalOrError returns json.Marshal or error.
func MarshalOrError(v any) ([]byte, error) {
	if v == nil {
		return nil, errors.New("pubsub: nil payload")
	}
	return json.Marshal(v)
}
