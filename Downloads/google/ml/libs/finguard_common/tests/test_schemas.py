"""Tests for finguard_common.schemas."""

from __future__ import annotations

from datetime import UTC, datetime

import pytest
from pydantic import ValidationError

from finguard_common.schemas import (
    ChatRequest,
    ChatResponse,
    Decision,
    Geo,
    ScoreRequest,
    ScoreResponse,
)


def _sample_score() -> ScoreRequest:
    return ScoreRequest(
        txn_id="T-1",
        customer_id="C-1",
        amount=100.0,
        currency="INR",
        merchant_id="M-1",
        geo=Geo(lat=12.97, lon=77.59),
        channel="CARD",
        timestamp=datetime(2026, 10, 15, tzinfo=UTC),
    )


def test_score_request_valid() -> None:
    req = _sample_score()
    assert req.txn_id == "T-1"
    assert req.amount == 100.0


def test_score_request_invalid_amount() -> None:
    with pytest.raises(ValidationError):
        ScoreRequest(
            txn_id="T-1",
            customer_id="C-1",
            amount=-1,
            currency="INR",
            merchant_id="M-1",
            geo=Geo(lat=0, lon=0),
            channel="CARD",
            timestamp=datetime.now(UTC),
        )


def test_geo_bounds() -> None:
    with pytest.raises(ValidationError):
        Geo(lat=120, lon=0)


def test_score_response_decision_coercion() -> None:
    resp = ScoreResponse(
        decision="approve",  # type: ignore[arg-type]
        fraud_probability=0.1,
        risk_score=100,
        reasons=[],
        model_version="v1",
    )
    assert resp.decision is Decision.APPROVE


def test_chat_roundtrip() -> None:
    req = ChatRequest(message="hi", session_id="S-1")
    assert req.session_id == "S-1"
    resp = ChatResponse(session_id="S-1", reply="hello")
    assert resp.session_id == "S-1"
