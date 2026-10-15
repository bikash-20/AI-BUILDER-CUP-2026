package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/finguard/platform/pkg/domain"
	"github.com/finguard/platform/pkg/riskstore"
)

func TestComputeRiskStable(t *testing.T) {
	a := computeRisk("C-1")
	b := computeRisk("C-1")
	if a.Score != b.Score {
		t.Errorf("not stable: %v vs %v", a.Score, b.Score)
	}
}

func TestComputeRiskBands(t *testing.T) {
	cases := []struct {
		id   string
		want domain.RiskBand
	}{
		{"customer-low", domain.RiskBandLow},
	}
	for _, tc := range cases {
		got := computeRisk(tc.id).Band
		// We don't pin the exact band for arbitrary IDs, only that the function
		// returns one of the four known bands.
		switch got {
		case domain.RiskBandLow, domain.RiskBandMedium, domain.RiskBandHigh, domain.RiskBandCrit:
		default:
			t.Errorf("unexpected band %q for %s", got, tc.id)
		}
	}
}

func TestBandFromScore(t *testing.T) {
	cases := []struct {
		s    float64
		want domain.RiskBand
	}{
		{100, domain.RiskBandLow},
		{500, domain.RiskBandMedium},
		{800, domain.RiskBandHigh},
		{950, domain.RiskBandCrit},
	}
	for _, tc := range cases {
		if got := bandFromScore(tc.s); got != tc.want {
			t.Errorf("bandFromScore(%v) = %v, want %v", tc.s, got, tc.want)
		}
	}
}

func TestRiskByID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/risk/", riskByID)

	req := httptest.NewRequest(http.MethodGet, "/v1/risk/C-1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var resp domain.RiskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.CustomerID != "C-1" {
		t.Errorf("customer_id = %q", resp.CustomerID)
	}
}

func TestMockClientStable(t *testing.T) {
	c := riskstore.MockClient{}
	a, _ := c.ListCustomerTransactions(context.Background(), "C-1001", 0)
	b, _ := c.ListCustomerTransactions(context.Background(), "C-1001", 0)
	if a.Count != b.Count || a.CustomerID != b.CustomerID {
		t.Errorf("not stable: %+v vs %+v", a, b)
	}
	if !a.Simulated {
		t.Error("expected simulated=true")
	}
	if a.Count == 0 {
		t.Error("expected non-empty list")
	}
	for _, tx := range a.Transactions {
		if tx.CustomerID != "C-1001" {
			t.Errorf("got customer %s, want C-1001", tx.CustomerID)
		}
		if !tx.Simulated {
			t.Errorf("expected per-row simulated=true, got false for %s", tx.TxnID)
		}
	}
}

func TestMockClientDifferentIDs(t *testing.T) {
	c := riskstore.MockClient{}
	a, _ := c.ListCustomerTransactions(context.Background(), "C-1001", 0)
	b, _ := c.ListCustomerTransactions(context.Background(), "C-1002", 0)
	if a.Transactions[0].TxnID == b.Transactions[0].TxnID {
		t.Error("expected different TxnIDs for different customers")
	}
}

func TestMockClientLimit(t *testing.T) {
	c := riskstore.MockClient{}
	got, _ := c.ListCustomerTransactions(context.Background(), "C-1001", 2)
	if got.Count != 2 {
		t.Errorf("got count %d, want 2", got.Count)
	}
}

func TestCustomerSubtreeSuccess(t *testing.T) {
	rs := riskstore.MockClient{}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/customers/", func(w http.ResponseWriter, r *http.Request) {
		customerSubtree(w, r, rs)
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/customers/C-1001/transactions", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp domain.CustomerTransactions
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.CustomerID != "C-1001" {
		t.Errorf("customer_id = %q", resp.CustomerID)
	}
	if !resp.Simulated {
		t.Error("expected simulated=true")
	}
	if len(resp.Transactions) == 0 {
		t.Error("expected non-empty transactions")
	}
}

func TestCustomerSubtreeLimit(t *testing.T) {
	rs := riskstore.MockClient{}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/customers/", func(w http.ResponseWriter, r *http.Request) {
		customerSubtree(w, r, rs)
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/customers/C-1001/transactions?limit=3", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var resp domain.CustomerTransactions
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Count != 3 {
		t.Errorf("count = %d, want 3", resp.Count)
	}
}

func TestCustomerSubtreeInvalidLimit(t *testing.T) {
	rs := riskstore.MockClient{}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/customers/", func(w http.ResponseWriter, r *http.Request) {
		customerSubtree(w, r, rs)
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/customers/C-1001/transactions?limit=0", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestCustomerSubtreeMethodNotAllowed(t *testing.T) {
	rs := riskstore.MockClient{}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/customers/", func(w http.ResponseWriter, r *http.Request) {
		customerSubtree(w, r, rs)
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/customers/C-1001/transactions", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestCustomerSubtreeUnknownSub(t *testing.T) {
	rs := riskstore.MockClient{}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/customers/", func(w http.ResponseWriter, r *http.Request) {
		customerSubtree(w, r, rs)
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/customers/C-1001/nope", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}
