package pubsub

import (
	"context"
	"testing"
)

func TestMockPublisherPublishJSON(t *testing.T) {
	p := NewMockPublisher()
	if err := p.PublishJSON(context.Background(), "topic-a", nil, map[string]int{"x": 1}); err != nil {
		t.Fatal(err)
	}
	if p.Published["topic-a"] != 1 {
		t.Errorf("count = %d", p.Published["topic-a"])
	}
}

func TestMockPublisherRejectsEmpty(t *testing.T) {
	p := NewMockPublisher()
	if err := p.PublishJSON(context.Background(), "", nil, map[string]int{}); err == nil {
		t.Error("expected empty topic error")
	}
	if err := p.PublishJSON(context.Background(), "t", nil, nil); err == nil {
		t.Error("expected nil payload error")
	}
}

func TestMockSubscriberInvokes(t *testing.T) {
	s := &MockSubscriber{N: 3}
	count := 0
	err := s.Subscribe(context.Background(), "sub", func(_ context.Context, _ []byte, _ map[string]string) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("count = %d", count)
	}
}

func TestMockSubscriberRejects(t *testing.T) {
	if err := (&MockSubscriber{}).Subscribe(context.Background(), "", nil); err == nil {
		t.Error("expected empty subscription")
	}
}
