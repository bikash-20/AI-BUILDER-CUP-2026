package riskstore

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/finguard/platform/pkg/domain"
)

// MockClient returns deterministic, simulated transactions without
// hitting a database. The shape mirrors the real BigQuery path so the
// frontend, BFF, and tests can rely on the same envelope.
//
// The data is clearly identified as simulated via the envelope fields
// (Simulated=true, Source="riskstore:MockClient") and the per-row
// Simulated flag. When riskstore.Select(cfg) wires the BigQuery client
// instead, those fields flip to false / "riskstore:BigQueryClient".
type MockClient struct{}

// seedMerchantPool is the catalog used to build deterministic customer
// transactions. The pool is small and shared so the resulting transactions
// look like a believable mix of merchant categories.
//
// Moved verbatim from services/cmd/risk-service/main.go so the existing
// in-process behavior is preserved when the live store is not configured.
var seedMerchantPool = []domain.Transaction{
	{
		TxnID: "M-blue-bottle", CustomerID: "C-1001", Amount: 8.5, Currency: "USD",
		MerchantID: "M-blue-bottle", MerchantCat: "FOOD", Channel: domain.ChannelCard,
		Timestamp: time.Date(2026, 10, 15, 8, 12, 0, 0, time.UTC),
	},
	{
		TxnID: "M-whole-foods", CustomerID: "C-1001", Amount: 64.21, Currency: "USD",
		MerchantID: "M-whole-foods", MerchantCat: "GROCERY", Channel: domain.ChannelCard,
		Timestamp: time.Date(2026, 10, 14, 18, 30, 0, 0, time.UTC),
	},
	{
		TxnID: "M-uber", CustomerID: "C-1001", Amount: 22.0, Currency: "USD",
		MerchantID: "M-uber", MerchantCat: "TRANSPORT", Channel: domain.ChannelMobile,
		Timestamp: time.Date(2026, 10, 14, 9, 5, 0, 0, time.UTC),
	},
	{
		TxnID: "M-amazon", CustomerID: "C-1001", Amount: 129.99, Currency: "USD",
		MerchantID: "M-amazon", MerchantCat: "RETAIL", Channel: domain.ChannelWeb,
		Timestamp: time.Date(2026, 10, 13, 21, 18, 0, 0, time.UTC),
	},
	{
		TxnID: "M-shell", CustomerID: "C-1001", Amount: 45.5, Currency: "USD",
		MerchantID: "M-shell", MerchantCat: "FUEL", Channel: domain.ChannelCard,
		Timestamp: time.Date(2026, 10, 12, 7, 42, 0, 0, time.UTC),
	},
	{
		TxnID: "M-netflix", CustomerID: "C-1001", Amount: 15.99, Currency: "USD",
		MerchantID: "M-netflix", MerchantCat: "SUBSCRIPTION", Channel: domain.ChannelWeb,
		Timestamp: time.Date(2026, 10, 10, 0, 5, 0, 0, time.UTC),
	},
}

// ListCustomerTransactions returns the seed pool, rewritten per customer
// and ordered most-recent-first, with limit applied (zero means all).
func (MockClient) ListCustomerTransactions(_ context.Context, customerID string, limit int) (domain.CustomerTransactions, error) {
	out := make([]domain.Transaction, 0, len(seedMerchantPool))
	for i, m := range seedMerchantPool {
		tx := m
		tx.TxnID = fmt.Sprintf("T-%s-%04d-%08x", customerID, i+1, stableHash(customerID+":"+m.MerchantID)&0xffffffff)
		tx.CustomerID = customerID
		tx.Simulated = true
		out = append(out, tx)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Timestamp.After(out[j].Timestamp)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return domain.CustomerTransactions{
		CustomerID:   customerID,
		Count:        len(out),
		Simulated:    true,
		Source:       "riskstore:MockClient",
		Transactions: out,
	}, nil
}

// stableHash is a tiny deterministic hash for IDs. Kept here so the
// mock produces stable TxnIDs regardless of which package called it.
func stableHash(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}
