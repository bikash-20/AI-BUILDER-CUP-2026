"""Pydantic schemas for FinGuard API contracts.

These mirror the Go domain types in services/pkg/domain/.
"""

from __future__ import annotations

from datetime import datetime
from enum import Enum
from typing import Any

from pydantic import BaseModel, Field, field_validator


class Geo(BaseModel):
    lat: float = Field(..., ge=-90, le=90)
    lon: float = Field(..., ge=-180, le=180)


class Decision(str, Enum):
    APPROVE = "APPROVE"
    REVIEW = "REVIEW"
    BLOCK = "BLOCK"


class Channel(str, Enum):
    CARD = "CARD"
    MOBILE = "MOBILE"
    WEB = "WEB"
    ATM = "ATM"
    UPI = "UPI"


class Transaction(BaseModel):
    txn_id: str
    customer_id: str
    amount: float = Field(..., gt=0)
    currency: str = Field(..., min_length=3, max_length=3)
    merchant_id: str
    merchant_category: str | None = None
    geo: Geo
    device_id: str | None = None
    ip_address: str | None = None
    channel: Channel
    timestamp: datetime


class ScoreRequest(BaseModel):
    txn_id: str
    customer_id: str
    amount: float = Field(..., gt=0)
    currency: str = Field(..., min_length=3, max_length=3)
    merchant_id: str
    geo: Geo
    device_id: str | None = None
    channel: Channel
    timestamp: datetime


class ScoreResponse(BaseModel):
    decision: Decision
    fraud_probability: float = Field(..., ge=0, le=1)
    risk_score: int = Field(..., ge=0, le=1000)
    reasons: list[str] = Field(default_factory=list)
    model_version: str

    @field_validator("decision", mode="before")
    @classmethod
    def _coerce_decision(cls, v: Any) -> Any:
        if isinstance(v, str):
            return v.upper()
        return v


class RiskBand(str, Enum):
    LOW = "LOW"
    MEDIUM = "MEDIUM"
    HIGH = "HIGH"
    CRITICAL = "CRITICAL"


class RiskScore(BaseModel):
    customer_id: str
    score: float = Field(..., ge=0, le=1000)
    band: RiskBand
    fraud_probability: float = 0.0
    credit_score: float = 0.0
    behavioral_score: float = 0.0
    compliance_flag: bool = False
    computed_at: datetime
    model_version: str


class ChatMessage(BaseModel):
    role: str = Field(..., pattern="^(user|assistant|system)$")
    content: str
    ts: datetime | None = None


class ChatRequest(BaseModel):
    session_id: str | None = None
    message: str
    history: list[ChatMessage] = Field(default_factory=list)


class ChatResponse(BaseModel):
    session_id: str
    reply: str
    citations: list[str] = Field(default_factory=list)
    actions: list[str] = Field(default_factory=list)


class Severity(str, Enum):
    LOW = "LOW"
    MEDIUM = "MEDIUM"
    HIGH = "HIGH"
    CRITICAL = "CRITICAL"


class Alert(BaseModel):
    alert_id: str
    txn_id: str
    customer_id: str
    severity: Severity
    reason: str
    model_version: str
    created_at: datetime


class DocumentType(str, Enum):
    KYC = "KYC"
    INVOICE = "INVOICE"
    CLAIM = "CLAIM"
    LOAN_APPLICATION = "LOAN_APPLICATION"
    IDENTITY = "IDENTITY"
    BANK_STATEMENT = "BANK_STATEMENT"


class DocumentEntity(BaseModel):
    type: str
    value: str
    confidence: float = Field(..., ge=0, le=1)
    page: int | None = None


class DocumentRecord(BaseModel):
    doc_id: str
    customer_id: str
    doc_type: DocumentType
    gcs_uri: str
    confidence: float = 0.0
    entities: list[DocumentEntity] = Field(default_factory=list)
    status: str
    uploaded_at: datetime
