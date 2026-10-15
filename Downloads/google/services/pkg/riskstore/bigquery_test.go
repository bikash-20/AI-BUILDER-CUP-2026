package riskstore

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

// Integration test for the live BigQuery path. Skipped when the env var
// is not set, so the standard `go test ./...` run stays offline.
//
//   TEST_BIGQUERY_EMULATOR_HOST=http://127.0.0.1:9050 \
//   TEST_BIGQUERY_PROJECT=finguard-dev \
//   TEST_BIGQUERY_DATASET=finguard_dev \
//     go test ./pkg/riskstore/...
//
// Before running, the emulator must be seeded. The
// scripts/seed_local.sh script does this end-to-end.

func openTestClient(t *testing.T) *BigQueryClient {
	t.Helper()
	host := os.Getenv("TEST_BIGQUERY_EMULATOR_HOST")
	if host == "" {
		t.Skip("TEST_BIGQUERY_EMULATOR_HOST not set; skipping live riskstore test")
	}
	project := os.Getenv("TEST_BIGQUERY_PROJECT")
	if project == "" {
		project = "finguard-dev"
	}
	dataset := os.Getenv("TEST_BIGQUERY_DATASET")
	if dataset == "" {
		dataset = "finguard_dev"
	}
	return &BigQueryClient{
		Host:       host,
		Project:    project,
		Dataset:    dataset,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func TestBigQueryClientListCustomerTransactions(t *testing.T) {
	c := openTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Use a customer id that the seed script produces. The seed script
	// generates ids in the range C-000000..C-000999, so C-000000 is a
	// safe pick.
	out, err := c.ListCustomerTransactions(ctx, "C-000000", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if out.Simulated {
		t.Error("expected Simulated=false for live client")
	}
	if out.Source != "riskstore:BigQueryClient" {
		t.Errorf("unexpected source: %q", out.Source)
	}
	if out.CustomerID != "C-000000" {
		t.Errorf("customer id mismatch: %q", out.CustomerID)
	}
}

func TestBigQueryClientUnknownCustomer(t *testing.T) {
	c := openTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := c.ListCustomerTransactions(ctx, "C-does-not-exist", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if out.Count != 0 {
		t.Errorf("expected empty result, got %d", out.Count)
	}
}

func TestSelectBigQuery(t *testing.T) {
	host := os.Getenv("TEST_BIGQUERY_EMULATOR_HOST")
	if host == "" {
		t.Skip("TEST_BIGQUERY_EMULATOR_HOST not set; skipping Select test")
	}
	t.Setenv("BIGQUERY_EMULATOR_HOST", host)
	// Build a minimal config struct via selectConfig helper is not
	// exported; just verify IsMock behavior.
	c := &BigQueryClient{Host: host, Project: "p", Dataset: "d"}
	if IsMock(c) {
		t.Error("expected non-mock client")
	}
}

func TestSelectMockFallback(t *testing.T) {
	t.Setenv("BIGQUERY_EMULATOR_HOST", "")
	c := Select(nil)
	if !IsMock(c) {
		t.Errorf("expected mock client, got %T", c)
	}
}

func TestIsSafeBQString(t *testing.T) {
	cases := []struct {
		s    string
		want bool
	}{
		{"C-1001", true},
		{"a", true},
		{"", false},
		{"C 1001", false},   // space
		{"C-1001'; DROP", false}, // quote
		{string([]byte{'a', 0x00, 'b'}), false}, // NUL
		{"a.b_c-d", true},
	}
	for _, tc := range cases {
		if got := isSafeBQString(tc.s); got != tc.want {
			t.Errorf("isSafeBQString(%q) = %v, want %v", tc.s, got, tc.want)
		}
	}
}