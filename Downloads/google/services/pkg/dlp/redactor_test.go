package dlp

import "testing"

func TestRedactEmail(t *testing.T) {
	r := NewRedactor()
	out := r.Redact("Please contact alice@example.com for help")
	if !contains(out, "[REDACTED_EMAIL]") {
		t.Errorf("email not redacted: %s", out)
	}
}

func TestRedactPhone(t *testing.T) {
	r := NewRedactor()
	out := r.Redact("Call me at +91 9876543210 now")
	if !contains(out, "[REDACTED_PHONE]") {
		t.Errorf("phone not redacted: %s", out)
	}
}

func TestRedactEmpty(t *testing.T) {
	r := NewRedactor()
	if got := r.Redact(""); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestContainsPII(t *testing.T) {
	r := NewRedactor()
	out := r.Redact("user: bob@x.com")
	if !r.ContainsPII(out) {
		t.Error("expected ContainsPII=true")
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
