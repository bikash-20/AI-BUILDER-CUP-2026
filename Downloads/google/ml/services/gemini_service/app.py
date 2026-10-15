"""FastAPI app for FinGuard gemini_service.

Endpoints:
    GET  /healthz
    POST /chat     - single-turn chat with RAG
    POST /summarize - summarize a case file
"""

from __future__ import annotations

from typing import Any

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from finguard_common.logging import get_logger
from finguard_gemini.chat import GeminiChatService
from finguard_gemini.guardrails import output_guardrails

log = get_logger("finguard.gemini_service")

app = FastAPI(title="FinGuard Gemini Service", version="0.1.0")


class ChatRequest(BaseModel):
    session_id: str | None = None
    message: str
    history: list[dict[str, str]] = Field(default_factory=list)


class ChatResponse(BaseModel):
    session_id: str
    reply: str
    citations: list[str]
    actions: list[str]


class SummarizeRequest(BaseModel):
    case_id: str
    body: str
    max_words: int = 200


class SummarizeResponse(BaseModel):
    case_id: str
    summary: str
    citations: list[str] = Field(default_factory=list)


_svc = GeminiChatService()


@app.get("/healthz")
def healthz() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/chat", response_model=ChatResponse)
def chat(req: ChatRequest) -> ChatResponse:
    if not req.message:
        raise HTTPException(status_code=400, detail="message is required")
    turn = _svc.chat(req.session_id or "anon", req.message, req.history)
    g = output_guardrails(turn.reply)
    if g.violation:
        turn.reply = "(output redacted by guardrails)"
    return ChatResponse(
        session_id=turn.session_id,
        reply=turn.reply,
        citations=turn.citations,
        actions=turn.actions,
    )


@app.post("/summarize", response_model=SummarizeResponse)
def summarize(req: SummarizeRequest) -> SummarizeResponse:
    text = req.body
    summary = (text[: req.max_words * 5] + "...") if len(text) > req.max_words * 5 else text
    return SummarizeResponse(case_id=req.case_id, summary=summary, citations=[])
