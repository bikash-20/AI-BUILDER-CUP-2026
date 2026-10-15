// Package domain contains types shared across FinGuard services.
//
// The types in this package are the canonical representations of business
// entities (Transaction, Alert, RiskScore, Document) that flow through the
// event bus and REST APIs.
package domain

import "time"

// Geo is a (lat, lon) pair in WGS-84.
type Geo struct {
	Lat float64 `json:"lat" validate:"gte=-90,lte=90"`
	Lon float64 `json:"lon" validate:"gte=-180,lte=180"`
}

// Decision is the outcome of a fraud score.
type Decision string

const (
	DecisionApprove Decision = "APPROVE"
	DecisionReview  Decision = "REVIEW"
	DecisionBlock   Decision = "BLOCK"
)

// Channel is the transaction channel (e.g., card, mobile, web).
type Channel string

const (
	ChannelCard   Channel = "CARD"
	ChannelMobile Channel = "MOBILE"
	ChannelWeb    Channel = "WEB"
	ChannelATM    Channel = "ATM"
	ChannelUPI    Channel = "UPI"
)

// Transaction is the canonical transaction event.
//
// Mirrors the BigQuery `fin.transactions` schema.
type Transaction struct {
	TxnID       string    `json:"txn_id"        validate:"required"`
	CustomerID  string    `json:"customer_id"   validate:"required"`
	Amount      float64   `json:"amount"        validate:"required,gt=0"`
	Currency    string    `json:"currency"      validate:"required,len=3"`
	MerchantID  string    `json:"merchant_id"   validate:"required"`
	MerchantCat string    `json:"merchant_category"`
	Geo         Geo       `json:"geo"`
	DeviceID    string    `json:"device_id"`
	IPAddress   string    `json:"ip_address"`
	Channel     Channel   `json:"channel"`
	Timestamp   time.Time `json:"timestamp"     validate:"required"`
	// Simulated is true when this transaction was produced by deterministic
	// seed data rather than a live BigQuery query. Downstream consumers
	// should surface a "simulated data" badge whenever any row carries this
	// flag. The field is omitted from JSON for real records.
	Simulated bool `json:"simulated,omitempty"`
}

// ScoreRequest is the inbound payload for POST /v1/score.
//
// It is a subset of Transaction that is meaningful without core-banking context.
type ScoreRequest struct {
	TxnID      string    `json:"txn_id"      validate:"required"`
	CustomerID string    `json:"customer_id" validate:"required"`
	Amount     float64   `json:"amount"      validate:"required,gt=0"`
	Currency   string    `json:"currency"    validate:"required,len=3"`
	MerchantID string    `json:"merchant_id" validate:"required"`
	Geo        Geo       `json:"geo"`
	DeviceID   string    `json:"device_id"`
	Channel    Channel   `json:"channel"`
	Timestamp  time.Time `json:"timestamp"   validate:"required"`
}

// ToTransaction converts a ScoreRequest to a canonical Transaction.
func (r ScoreRequest) ToTransaction() Transaction {
	return Transaction{
		TxnID:      r.TxnID,
		CustomerID: r.CustomerID,
		Amount:     r.Amount,
		Currency:   r.Currency,
		MerchantID: r.MerchantID,
		Geo:        r.Geo,
		DeviceID:   r.DeviceID,
		Channel:    r.Channel,
		Timestamp:  r.Timestamp,
	}
}

// CustomerTransactions is the envelope returned by risk-service for the
// GET /v1/customers/{id}/transactions endpoint. The Simulated and Source
// fields let the BFF and frontend clearly identify mock data. We omit
// `omitempty` on Simulated so the JSON always carries the flag — the
// frontend decides whether to show the "Simulated data" badge based on
// this field, and a missing key is indistinguishable from a bug.
type CustomerTransactions struct {
	CustomerID   string        `json:"customer_id"`
	Count        int           `json:"count"`
	Simulated    bool          `json:"simulated"`
	Source       string        `json:"source,omitempty"`
	Transactions []Transaction `json:"transactions"`
}
