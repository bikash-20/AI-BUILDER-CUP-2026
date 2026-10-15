package alloydb

import (
	"context"
	"testing"
	"time"
)

func TestMockInsertAlert(t *testing.T) {
	c := MockClient{}
	if err := c.InsertAlert(context.Background(), AlertRecord{AlertID: "a", CreatedAt: time.Now()}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := c.InsertAlert(context.Background(), AlertRecord{}); err == nil {
		t.Error("expected error for empty alert id")
	}
}

func TestMockMarkResolved(t *testing.T) {
	c := MockClient{}
	if err := c.MarkResolved(context.Background(), "a", "OK"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := c.MarkResolved(context.Background(), "", "OK"); err == nil {
		t.Error("expected error for empty alert id")
	}
}

func TestMockListAlerts(t *testing.T) {
	c := MockClient{}

	// Unfiltered: returns all seed rows, most-recent-first.
	got, err := c.ListAlerts(context.Background(), ListAlertsFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(seedAlerts) {
		t.Errorf("got %d alerts, want %d", len(got), len(seedAlerts))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].CreatedAt.Before(got[i].CreatedAt) {
			t.Errorf("list not sorted newest-first at index %d", i)
		}
	}

	// Customer filter narrows results to a known multi-record customer.
	const wantCustomer = "C-1001"
	filtered, err := c.ListAlerts(context.Background(), ListAlertsFilter{CustomerID: wantCustomer})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) < 2 {
		t.Errorf("expected at least 2 alerts for %s, got %d", wantCustomer, len(filtered))
	}
	for _, a := range filtered {
		if a.CustomerID != wantCustomer {
			t.Errorf("got customer %s, want %s", a.CustomerID, wantCustomer)
		}
	}

	// Unknown customer returns an empty list, not an error.
	empty, err := c.ListAlerts(context.Background(), ListAlertsFilter{CustomerID: "C-DOES-NOT-EXIST"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("got %d alerts, want 0", len(empty))
	}

	// Limit caps the result size.
	limited, err := c.ListAlerts(context.Background(), ListAlertsFilter{Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("got %d alerts, want 2", len(limited))
	}
}
