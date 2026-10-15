# FinGuard AI — Go Services

Cloud Run microservices written in **Go 1.22+**. Stateless, single-binary, distroless.

## Services

| Cmd | Purpose | Key Endpoints |
|---|---|---|
| `api-gateway` | Edge proxy, JWT verification, rate limit | `/*` |
| `fraud-service` | Real-time transaction scoring (FR-1/FR-2) | `POST /v1/score` |
| `risk-service` | Customer risk score (FR-5) | `GET /v1/risk/{customer_id}` |
| `document-service` | KYC/document handling BFF (FR-3) | `POST /v1/documents/upload`, `GET /v1/documents/{id}` |
| `chat-bff` | Chat BFF (FR-4) | `POST /v1/chat` |
| `compliance-service` | AML/SAR/CTR (FR-6) | `POST /v1/compliance/sar/draft` |
| `alert-worker` | Background alert notification | (Pub/Sub subscriber) |

## Layout

```
services/
├── go.mod
├── pkg/          # shared libraries (auth, config, logging, ...)
└── cmd/          # one binary per service
```

## Local Development

Each service can be run in **mock mode** without GCP credentials:

```bash
cd cmd/fraud-service
MOCK_MODE=true go run .
# -> binds to :8080, responds to /healthz, /v1/score (rule-based)
```

Configuration via env (12-factor):

| Variable | Default | Notes |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `MOCK_MODE` | `false` | Use in-memory clients |
| `GCP_PROJECT_ID` | `finguard-dev` | |
| `VERTEX_ENDPOINT_ID` | — | Online prediction endpoint |
| `BIGQUERY_PROJECT` | `finguard-dev` | |
| `PUBSUB_TOPIC_SCORED` | `txn-scored` | |
| `ALLOYDB_DSN` | — | libpq DSN |

## Tests

```bash
cd services
go test ./...
```

## Build & Container

Each service has a `Dockerfile` in `services/deploy/`:

```bash
docker build -f deploy/fraud-service.Dockerfile -t finguard/fraud-service:dev .
```
