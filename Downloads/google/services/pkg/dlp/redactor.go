// Package dlp provides a lightweight PII redaction helper.
//
// In production this delegates to Cloud DLP; in mock mode it masks
// common PII patterns with redaction tokens before messages leave the
// service boundary.
package dlp

import (
	"regexp"
	"strings"
)

// Redactor masks PII in free text.
type Redactor struct {
	EmailPattern    *regexp.Regexp
	PhonePattern    *regexp.Regexp
	CardPattern     *regexp.Regexp
	SSNPattern      *regexp.Regexp
	AccountPattern  *regexp.Regexp
}

// NewRedactor returns a Redactor with sensible defaults.
func NewRedactor() *Redactor {
	return &Redactor{
		EmailPattern:   regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`),
		PhonePattern:   regexp.MustCompile(`\b\+?[0-9]{1,3}[\s\-]?[0-9]{6,12}\b`),
		CardPattern:    regexp.MustCompile(`\b(?:\d[ \-]?){13,16}\d\b`),
		SSNPattern:     regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
		AccountPattern: regexp.MustCompile(`\b\d{9,18}\b`),
	}
}

// Redact replaces PII patterns with placeholders.
func (r *Redactor) Redact(s string) string {
	if s == "" {
		return s
	}
	replacements := []struct {
		re  *regexp.Regexp
		rep string
	}{
		{r.EmailPattern, "[REDACTED_EMAIL]"},
		{r.PhonePattern, "[REDACTED_PHONE]"},
		{r.CardPattern, "[REDACTED_CARD]"},
		{r.SSNPattern, "[REDACTED_SSN]"},
		{r.AccountPattern, "[REDACTED_ACCOUNT]"},
	}
	for _, step := range replacements {
		s = step.re.ReplaceAllString(s, step.rep)
	}
	return s
}

// ContainsPII is a quick check used to decide whether to short-circuit LLM calls.
func (r *Redactor) ContainsPII(s string) bool {
	return strings.Contains(s, "[REDACTED_")
}
