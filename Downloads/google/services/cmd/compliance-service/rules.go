package main

// Rules is a placeholder for the AML rule engine. In production this would
// evaluate a list of rules and emit a structured RuleResult.
type Rules struct{}

// NewRules creates an empty rule set.
func NewRules() *Rules { return &Rules{} }

// Evaluate returns true if the case meets the criteria for filing.
func (r *Rules) Evaluate(amount float64, suspiciousCountries int) bool {
	if amount >= 10_000 {
		return true
	}
	if suspiciousCountries > 2 {
		return true
	}
	return false
}
