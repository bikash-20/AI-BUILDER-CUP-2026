"""KYC / AML rule validators over extracted documents."""

from __future__ import annotations

import re
from dataclasses import dataclass, field
from datetime import datetime
from typing import Any

from finguard_common.logging import get_logger

log = get_logger("finguard.docai.validators")


@dataclass
class ValidationResult:
    valid: bool
    confidence_threshold: float = 0.9
    issues: list[str] = field(default_factory=list)
    needs_review: bool = False

    def to_dict(self) -> dict[str, Any]:
        return {
            "valid": self.valid,
            "needs_review": self.needs_review,
            "issues": self.issues,
        }


class KYCValidator:
    """Validate identity documents against basic KYC rules."""

    def __init__(self, confidence_threshold: float = 0.9) -> None:
        self.confidence_threshold = confidence_threshold
        self.id_regex = re.compile(r"^[A-Z]{5}[0-9]{4}[A-Z]$")  # PAN-like

    def validate(self, doc: dict[str, Any]) -> ValidationResult:
        entities = doc.get("entities", []) if isinstance(doc, dict) else []
        result = ValidationResult(valid=True, confidence_threshold=self.confidence_threshold)

        if not entities:
            result.valid = False
            result.needs_review = True
            result.issues.append("no_entities_extracted")
            return result

        # Confidence check
        below = [e for e in entities if e.get("confidence", 1.0) < self.confidence_threshold]
        if below:
            result.needs_review = True
            result.issues.append("low_confidence")

        # Check that we have a name and an ID
        types_present = {e.get("type") for e in entities}
        if "person_name" not in types_present and "name" not in types_present:
            result.valid = False
            result.issues.append("missing_name")

        id_value = next(
            (e["value"] for e in entities if e.get("type") in {"id_number", "pan", "passport"}),
            None,
        )
        if id_value and not self.id_regex.match(id_value):
            result.valid = False
            result.issues.append("invalid_id_format")

        # Check DOB is in the past
        dob = next((e["value"] for e in entities if e.get("type") == "dob"), None)
        if dob:
            try:
                date_val = datetime.fromisoformat(dob)
                if date_val > datetime.now():
                    result.valid = False
                    result.issues.append("dob_in_future")
            except Exception:
                result.needs_review = True
                result.issues.append("unparseable_dob")

        return result


class AMLValidator:
    """Validate transaction-related documents against AML rules."""

    def validate_invoice(self, doc: dict[str, Any]) -> ValidationResult:
        result = ValidationResult(valid=True)
        entities = doc.get("entities", []) if isinstance(doc, dict) else []
        types_present = {e.get("type") for e in entities}

        for required in ("invoice_number", "total_amount", "due_date"):
            if required not in types_present:
                result.valid = False
                result.needs_review = True
                result.issues.append(f"missing_{required}")

        # Amount sanity
        amount = next((e["value"] for e in entities if e.get("type") == "total_amount"), None)
        if amount:
            try:
                val = float(amount)
                if val <= 0:
                    result.valid = False
                    result.issues.append("non_positive_amount")
            except ValueError:
                result.valid = False
                result.issues.append("unparseable_amount")
        return result
