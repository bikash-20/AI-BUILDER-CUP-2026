"""FastAPI app for FinGuard document_service.

Endpoints:
    GET  /healthz
    POST /extract     - call Document AI on a gcs_uri
    POST /validate    - validate an extracted document against KYC rules
"""

from __future__ import annotations

from typing import Any

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from finguard_common.config import is_mock_mode
from finguard_common.logging import get_logger
from finguard_docai.extractor import DocumentAIExtractor, MockDocumentExtractor
from finguard_docai.validators import KYCValidator

log = get_logger("finguard.document_service")

app = FastAPI(title="FinGuard Document Service", version="0.1.0")


class ExtractRequest(BaseModel):
    gcs_uri: str = Field(..., pattern=r"^gs://.+$")
    doc_type: str = Field(default="KYC")


class ExtractResponse(BaseModel):
    doc_type: str
    entities: list[dict[str, Any]]
    pages: int
    confidence: float


class ValidateRequest(BaseModel):
    doc_type: str = "KYC"
    entities: list[dict[str, Any]] = Field(default_factory=list)
    confidence: float = 0.0


class ValidateResponse(BaseModel):
    valid: bool
    needs_review: bool
    issues: list[str]


def _extractor() -> Any:
    if is_mock_mode():
        return MockDocumentExtractor()
    return DocumentAIExtractor()


@app.get("/healthz")
def healthz() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/extract", response_model=ExtractResponse)
def extract(req: ExtractRequest) -> ExtractResponse:
    log.info("extract_request", gcs_uri=req.gcs_uri, doc_type=req.doc_type)
    result = _extractor().extract(req.gcs_uri, req.doc_type)
    return ExtractResponse(
        doc_type=result.doc_type,
        entities=result.entities,
        pages=result.pages,
        confidence=result.confidence,
    )


@app.post("/validate", response_model=ValidateResponse)
def validate(req: ValidateRequest) -> ValidateResponse:
    if req.doc_type.upper() not in {"KYC", "IDENTITY"}:
        raise HTTPException(status_code=400, detail="only KYC/IDENTITY supported")
    v = KYCValidator().validate({"entities": req.entities})
    return ValidateResponse(
        valid=v.valid,
        needs_review=v.needs_review,
        issues=v.issues,
    )
