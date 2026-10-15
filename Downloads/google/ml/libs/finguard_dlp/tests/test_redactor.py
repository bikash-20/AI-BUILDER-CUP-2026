"""Tests for finguard_dlp.redactor."""

from __future__ import annotations

from finguard_dlp.redactor import RegexRedactor


def test_redact_email() -> None:
    r = RegexRedactor().redact("ping alice@example.com for help")
    assert "[REDACTED_EMAIL]" in r.text
    assert "EMAIL" in r.categories


def test_redact_multiple() -> None:
    r = RegexRedactor().redact(
        "call +91 9876543210 and email bob@x.com, my card 4111111111111111"
    )
    assert any(c in r.categories for c in ("PHONE", "EMAIL", "CARD"))
    assert r.redacted_count >= 3


def test_redact_pan() -> None:
    r = RegexRedactor().redact("my PAN is ABCDE1234F")
    assert "PAN" in r.categories
