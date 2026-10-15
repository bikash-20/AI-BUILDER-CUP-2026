"""Guardrails for Gemini-powered chat.

Provides pre-call content moderation and post-call output checks.
"""

from __future__ import annotations

import re
from dataclasses import dataclass, field


@dataclass
class GuardrailViolation(Exception):
    code: str
    reason: str


@dataclass
class Guardrail:
    """Result of guardrail evaluation."""

    violation: bool = False
    code: str | None = None
    reason: str = ""
    reasons: list[str] = field(default_factory=list)


# Substrings considered unsafe (illustrative; expand with real classifier).
TOXIC_KEYWORDS = [
    "rm -rf",
    "drop table",
    "ignore previous",
    "ignore all instructions",
]


def apply_guardrails(text: str) -> Guardrail:
    """Run input-side guardrails. Returns Guardrail indicating violation if any."""
    if not text:
        return Guardrail(violation=True, code="empty", reason="empty input")

    lower = text.lower()
    for k in TOXIC_KEYWORDS:
        if k in lower:
            return Guardrail(
                violation=True, code="toxic", reason=f"matched toxic keyword: {k}"
            )

    if len(text) > 8000:
        return Guardrail(violation=True, code="too_long", reason="input too long")

    return Guardrail()


# Output-side checks

LEAK_PATTERNS = [
    re.compile(r"\b(?:\d[ -]?){13,19}\b"),  # 13-19 digit card number, optional spaces/dashes
    re.compile(r"\b\d{3}-\d{2}-\d{4}\b"),  # SSN
]


def output_guardrails(text: str) -> Guardrail:
    """Run output-side guardrails. Returns Guardrail indicating violation if any."""
    for p in LEAK_PATTERNS:
        if p.search(text):
            return Guardrail(
                violation=True, code="pii_leak", reason="output contains raw PII pattern"
            )
    return Guardrail()
