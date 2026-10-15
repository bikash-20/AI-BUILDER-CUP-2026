package domain

import "time"

// Severity is the alert severity.
type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// AlertStatus is the lifecycle status of an alert.
type AlertStatus string

const (
	AlertStatusOpen     AlertStatus = "OPEN"
	AlertStatusAssigned AlertStatus = "ASSIGNED"
	AlertStatusResolved AlertStatus = "RESOLVED"
	AlertStatusClosed   AlertStatus = "CLOSED"
)

// Resolution captures the analyst's verdict on a resolved alert.
type Resolution string

const (
	ResolutionConfirmedFraud Resolution = "CONFIRMED_FRAUD"
	ResolutionFalsePositive  Resolution = "FALSE_POSITIVE"
	ResolutionInconclusive   Resolution = "INCONCLUSIVE"
)

// Alert represents a fraud alert produced by fraud-service.
type Alert struct {
	AlertID      string      `json:"alert_id"`
	TxnID        string      `json:"txn_id"`
	CustomerID   string      `json:"customer_id"`
	Severity     Severity    `json:"severity"`
	Status       AlertStatus `json:"status"`
	Reason       string      `json:"reason"`
	ModelVersion string      `json:"model_version"`
	CreatedAt    time.Time   `json:"created_at"`
	ResolvedAt   *time.Time  `json:"resolved_at,omitempty"`
	Resolution   *Resolution `json:"resolution,omitempty"`
	AssignedTo   string      `json:"assigned_to,omitempty"`
	// Simulated is true when this alert was produced by deterministic seed
	// data rather than a live AlloyDB query. Downstream consumers should
	// surface a "simulated data" badge whenever any record carries this flag.
	Simulated bool `json:"simulated,omitempty"`
}
