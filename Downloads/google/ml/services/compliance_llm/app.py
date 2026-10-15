"""FastAPI app that drafts SAR/CTR filings using Gemini + RAG.

Endpoints:
    GET  /healthz
    POST /draft_sar
    POST /draft_ctr
"""

from __future__ import annotations

import uuid
from typing import Any

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from finguard_common.logging import get_logger
from finguard_gemini.chat import GeminiChatService
from finguard_gemini.rag import build_rag_context

log = get_logger("finguard.compliance_llm")

app = FastAPI(title="FinGuard Compliance LLM", version="0.1.0")


class DraftRequest(BaseModel):
    case_id: str
    transactions: list[dict[str, Any]] = Field(default_factory=list)
    notes: str = ""


class DraftResponse(BaseModel):
    draft_id: str
    case_id: str
    filing_type: str
    summary: str
    citations: list[str]


_svc = GeminiChatService()


@app.get("/healthz")
def healthz() -> dict[str, str]:
    return {"status": "ok"}


def _build_summary(filing: str, req: DraftRequest, ctx: str) -> str:
    n = len(req.transactions)
    amount = sum(float(t.get("amount", 0.0)) for t in req.transactions)
    return (
        f"{filing} draft for case {req.case_id}. "
        f"Involves {n} transactions totaling {amount:.2f}. "
        f"Notes: {req.notes or 'n/a'}. "
        f"Refer to policy: {ctx[:200]}"
    )


@app.post("/draft_sar", response_model=DraftResponse)
def draft_sar(req: DraftRequest) -> DraftResponse:
    return _draft("SAR", req)


@app.post("/draft_ctr", response_model=DraftResponse)
def draft_ctr(req: DraftRequest) -> DraftResponse:
    return _draft("CTR", req)


def _draft(filing: str, req: DraftRequest) -> DraftResponse:
    if not req.case_id:
        raise HTTPException(status_code=400, detail="case_id required")
    ctx, citations = build_rag_context(f"{filing} policy for case {req.case_id}")
    summary = _build_summary(filing, req, ctx)
    return DraftResponse(
        draft_id=uuid.uuid4().hex,
        case_id=req.case_id,
        filing_type=filing,
        summary=summary,
        citations=citations,
    )
