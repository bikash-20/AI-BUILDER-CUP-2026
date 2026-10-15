# FinGuard AI — REST API Reference

> All endpoints are versioned under `/v1`. The OpenAPI source-of-truth is in [`terraform/modules/apigateway/openapi.yaml`](../terraform/modules/apigateway/openapi.yaml).

## Conventions

- **Base URL:** `https://api.finguard.example/v1`
- **Auth:** `Authorization: Bearer <jwt>` (validated at IAP + service).
- **Content-Type:** `application/json`
- **Idempotency:** clients may send `Idempotency-Key` header for non-GET requests.
- **Correlation:** every request gets `X-Request-Id` echoed in the response.
- **Errors:** standard envelope `{ "error": { "code": "...", "message": "...", "request_id": "..." } }`.

## Endpoints

### 1. Score Transaction

`POST /v1/score`

Request:
```json
{
  "txn_id": "T123",
  "customer_id": "C456",
  "amount": 1500.00,
  "currency": "INR",
  "merchant_id": "M789",
  "geo": { "lat": 12.97, "lon": 77.59 },
  "device_id": "D001",
  "timestamp": "2026-10-15T10:00:00Z"
}
```

Response (200):
```json
{
  "decision": "REVIEW",
  "fraud_probability": 0.87,
  "risk_score": 812,
  "reasons": ["unusual_geo", "high_velocity"],
  "model_version": "v3.2"
}
```

Errors: `400 invalid_request`, `401 unauthorized`, `429 rate_limited`, `500 upstream_timeout`.

### 2. Upload Document

`POST /v1/documents/upload`

Request:
```json
{ "doc_type": "KYC", "file_url": "gs://..." }
```

Response (202):
```json
{ "doc_id": "DOC123", "status": "PROCESSING" }
```

### 3. Get Document

`GET /v1/documents/{doc_id}` → `{ doc_id, status, extracted_json, confidence }`

### 4. Chat

`POST /v1/chat`

Request:
```json
{ "session_id": "S1", "message": "Why was my card declined?" }
```

Response (200):
```json
{
  "reply": "Your card was declined due to a high-risk geo-velocity pattern...",
  "citations": ["policy://risk/geo-velocity-v2"],
  "actions": ["raise_dispute"]
}
```

### 5. Get Risk Score

`GET /v1/risk/{customer_id}` → `{ score, band, breakdown }`

### 6. Recompute Risk Score

`POST /v1/risk/recompute` → `{ customer_id }` → 202 `{ "status": "queued" }`

### 7. Draft SAR

`POST /v1/compliance/sar/draft` → `{ case_id }` → 202 `{ "draft_id": "..." }`

### 8. List Alerts

`GET /v1/alerts?status=open&severity=high` → `{ alerts: [...] }`

### 9. Resolve Alert

`POST /v1/alerts/{alert_id}/resolve` → `{ resolution: "CONFIRMED_FRAUD" | "FALSE_POSITIVE", notes: "..." }`

## Rate Limits

- Per-user: 100 req / min
- Per-IP: 1000 req / min
- Score endpoint: 50,000 TPS aggregate

## SLOs

| Endpoint | SLO |
|---|---|
| `/v1/score` | P95 < 500 ms, 99.95 % available |
| `/v1/chat` | P95 < 2 s, 99 % available |
| `/v1/documents/*` | P95 < 30 s, 99 % available |
| `/v1/risk/*` | P95 < 200 ms, 99.9 % available |
