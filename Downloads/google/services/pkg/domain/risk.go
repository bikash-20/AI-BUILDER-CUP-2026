package domain

import "time"

// RiskBand classifies a Customer risk score into a band.
type RiskBand string

const (
	RiskBandLow    RiskBand = "LOW"
	RiskBandMedium RiskBand = "MEDIUM"
	RiskBandHigh   RiskBand = "HIGH"
	RiskBandCrit   RiskBand = "CRITICAL"
)

// RiskScore is the composite customer risk score.
//
// Score is in [0, 1000]; 0 = safest, 1000 = highest risk.
//
// It is the public-facing representation of the fin.risk_scores table.
type RiskScore struct {
	CustomerID      string    `json:"customer_id"`
	Score           float64   `json:"score"             validate:"gte=0,lte=1000"`
	Band            RiskBand  `json:"band"`
	FraudProbability float64  `json:"fraud_probability"`
	CreditScore     float64   `json:"credit_score"`
	BehavioralScore float64   `json:"behavioral_score"`
	ComplianceFlag  bool      `json:"compliance_flag"`
	ComputedAt      time.Time `json:"computed_at"`
	ModelVersion    string    `json:"model_version"`
}

// RiskBreakdown details how the composite score was derived.
type RiskBreakdown struct {
	Fraud      float64 `json:"fraud"`
	Credit     float64 `json:"credit"`
	Behavioral float64 `json:"behavioral"`
	Compliance float64 `json:"compliance"`
}

// RiskResponse is the API response for GET /v1/risk/{customer_id}.
type RiskResponse struct {
	CustomerID string        `json:"customer_id"`
	Score      float64       `json:"score"`
	Band       RiskBand      `json:"band"`
	Breakdown  RiskBreakdown `json:"breakdown"`
	UpdatedAt  time.Time     `json:"updated_at"`
}
