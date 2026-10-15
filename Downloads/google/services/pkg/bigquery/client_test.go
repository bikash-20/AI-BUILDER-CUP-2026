package bigquery

import (
	"context"
	"testing"
)

func TestMockClientStable(t *testing.T) {
	c := MockClient{}
	a, _ := c.GetCustomerFeatures(context.Background(), "C-1")
	b, _ := c.GetCustomerFeatures(context.Background(), "C-1")
	if a.AmountZScore != b.AmountZScore {
		t.Errorf("not stable: %v vs %v", a.AmountZScore, b.AmountZScore)
	}
}

func TestMockClientEmptyID(t *testing.T) {
	c := MockClient{}
	if _, err := c.GetCustomerFeatures(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}
