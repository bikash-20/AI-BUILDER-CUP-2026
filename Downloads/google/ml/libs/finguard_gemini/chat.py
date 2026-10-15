"""Gemini-based conversational service.

Uses Vertex AI's Gemini for reasoning with optional RAG over Vertex AI Search.
"""

from __future__ import annotations

import time
import uuid
from dataclasses import dataclass, field
from typing import Any

from finguard_common.config import get_settings, is_mock_mode
from finguard_common.logging import get_logger
from finguard_gemini.guardrails import apply_guardrails
from finguard_gemini.rag import build_rag_context

log = get_logger("finguard.gemini.chat")


SYSTEM_PROMPT = """You are FinGuard Assistant, an AI helper for fraud, risk, and
compliance analysts of a regulated financial institution.

- Always ground answers in retrieved policy documents when available.
- Cite your sources using the URI prefix provided (e.g., policy://...).
- If you cannot answer with high confidence, escalate to a human analyst.
- Never expose raw PII; rely on the placeholder tokens already in the input.
"""


@dataclass
class ChatTurn:
    session_id: str
    user_message: str
    reply: str
    citations: list[str] = field(default_factory=list)
    actions: list[str] = field(default_factory=list)
    created_at: float = field(default_factory=time.time)


class GeminiChatService:
    """Orchestrates Gemini + RAG; mockable."""

    def __init__(self, model: str | None = None) -> None:
        s = get_settings()
        self.model_name = model or s.gemini_model
        self._model: Any = None
        if not is_mock_mode():
            try:
                import vertexai  # type: ignore
                from vertexai.generative_models import GenerativeModel  # type: ignore

                vertexai.init(project=s.project_id, location=s.region)
                self._model = GenerativeModel(self.model_name, system_instruction=SYSTEM_PROMPT)
                log.info("gemini_initialized", model=self.model_name)
            except Exception as e:  # pragma: no cover
                log.warning("gemini_init_failed", error=str(e))
                self._model = None

    def chat(
        self,
        session_id: str,
        user_message: str,
        history: list[dict[str, str]] | None = None,
    ) -> ChatTurn:
        """Produce a single reply."""

        guard = apply_guardrails(user_message)
        if guard.violation:
            return ChatTurn(
                session_id=session_id,
                user_message=user_message,
                reply="I cannot process this request. Reason: " + guard.reason,
                citations=[],
                actions=["escalate"],
            )

        if self._model is None:
            return self._mock_reply(session_id, user_message)

        ctx, citations = build_rag_context(user_message)
        prompt = f"""Use the following context to answer the user question.
Context:
{ctx}

User: {user_message}
"""
        try:
            resp = self._model.generate_content(prompt)
            text = getattr(resp, "text", "") or ""
        except Exception as e:
            log.warning("gemini_generate_failed", error=str(e))
            return self._mock_reply(session_id, user_message)

        return ChatTurn(
            session_id=session_id,
            user_message=user_message,
            reply=text.strip() or "(no response)",
            citations=citations,
            actions=[],
        )

    def _mock_reply(self, session_id: str, message: str) -> ChatTurn:
        citations = ["policy://risk/geo-velocity-v2"] if "decline" in message.lower() else []
        actions = ["raise_dispute"] if "dispute" in message.lower() else []
        return ChatTurn(
            session_id=session_id or uuid.uuid4().hex,
            user_message=message,
            reply="(mock) Your request has been acknowledged. An analyst will follow up.",
            citations=citations,
            actions=actions,
        )
