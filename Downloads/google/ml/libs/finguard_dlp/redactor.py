"""PII redaction helpers backed by Cloud DLP.

`RegexRedactor` runs entirely offline (used in tests + local dev).
`DlpRedactor` calls the Cloud DLP API in production.
"""

from __future__ import annotations

import re
from dataclasses import dataclass, field
from typing import Any

from finguard_common.config import is_mock_mode
from finguard_common.logging import get_logger

log = get_logger("finguard.dlp")


@dataclass
class RedactionResult:
    text: str
    redacted_count: int = 0
    categories: list[str] = field(default_factory=list)


# Common PII patterns.
PATTERNS = {
    "EMAIL": re.compile(r"[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+"),
    "PHONE": re.compile(r"\+?\d[\d\s\-]{8,}\d"),
    "PAN": re.compile(r"\b[A-Z]{5}\d{4}[A-Z]\b"),
    "SSN": re.compile(r"\b\d{3}-\d{2}-\d{4}\b"),
    "CARD": re.compile(r"\b\d{12,19}\b"),
}


class RegexRedactor:
    """In-process redactor for tests and offline scenarios."""

    def redact(self, text: str) -> RedactionResult:
        categories: list[str] = []
        count = 0
        for name, pat in PATTERNS.items():
            new_text, n = pat.subn(f"[REDACTED_{name}]", text)
            if n > 0:
                categories.append(name)
                count += n
                text = new_text
        return RedactionResult(text=text, redacted_count=count, categories=categories)


class DlpRedactor:
    """Cloud DLP-backed redactor."""

    def __init__(self, project: str | None = None) -> None:
        from finguard_common.config import project_id as _project_id

        self.project = project or _project_id()
        self._client: Any = None
        if not is_mock_mode():
            try:
                from google.cloud import dlp_v2  # type: ignore

                self._client = dlp_v2.DlpServiceClient()
            except Exception as e:
                log.warning("dlp_init_failed", error=str(e))

    def redact(self, text: str) -> RedactionResult:
        if self._client is None:
            return RegexRedactor().redact(text)
        try:
            from google.cloud import dlp_v2  # type: ignore

            info_types = [
                {"name": "EMAIL_ADDRESS"},
                {"name": "PHONE_NUMBER"},
                {"name": "CREDIT_CARD_NUMBER"},
                {"name": "INDIA_PAN_INDIVIDUAL"},
            ]
            request = {
                "parent": f"projects/{self.project}",
                "item": {"value": text},
                "deidentify_config": {
                    "info_type_transformations": {
                        "transformations": [
                            {"primitive_transformation": {"replace_with_info_type_config": {}}}
                        ]
                    }
                },
                "inspect_config": {"info_types": info_types},
            }
            response = self._client.deidentify_content(request=request)
            return RedactionResult(text=response.item.value)
        except Exception as e:
            log.warning("dlp_redact_failed", error=str(e))
            return RegexRedactor().redact(text)
