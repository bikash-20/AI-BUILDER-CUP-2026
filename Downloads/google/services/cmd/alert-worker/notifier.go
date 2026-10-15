package main

import (
	"context"
	"fmt"
	"time"
)

// Notifier sends alert notifications over multiple channels.
type Notifier struct {
	Channels []string
}

// NewNotifier returns a notifier with default channels.
func NewNotifier() *Notifier {
	return &Notifier{Channels: []string{"websocket", "email"}}
}

// Send emits a notification for the alert.
func (n *Notifier) Send(_ context.Context, alert alertMsg) error {
	for _, ch := range n.Channels {
		fmt.Printf("notify[%s] %s sev=%s\n", ch, alert.AlertID, alert.Severity)
	}
	return nil
}

// keep time package used in scaffold
var _ = time.Second
