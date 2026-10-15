package main

import (
	"context"
	"testing"
)

func TestNotifierSend(t *testing.T) {
	n := NewNotifier()
	if err := n.Send(context.Background(), alertMsg{AlertID: "A-1", Severity: "HIGH"}); err != nil {
		t.Errorf("Send error: %v", err)
	}
}
