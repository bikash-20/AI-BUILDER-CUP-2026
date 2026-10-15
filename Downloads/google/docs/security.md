# FinGuard AI — Security & Compliance

## 1. Identity & Auth

- **End users** — Identity Platform + OAuth2, MFA enforced.
- **Service-to-service** — Workload Identity; no service-account keys in env.
- **Edge auth** — IAP in front of Cloud Run; Apigee / Cloud Endpoints enforces JWT validation.
- **Analyst access** — role-based (analyst, risk officer, compliance officer) enforced both at IAP and in services.

## 2. Network Security

- **Hub-and-spoke VPC** with Network Connectivity Center (NCC).
- **No public IPs in prod** — Private Service Connect to Google APIs.
- **Cloud NAT** for egress; no direct internet for prod workloads.
- **Cloud Armor (WAF)** in front of public endpoints; managed rule sets + custom rules.
- **VPC Service Controls (VPC-SC)** perimeter around BigQuery, GCS, Vertex AI, Document AI, Cloud Run.

## 3. Data Protection

- **At rest** — CMEK via Cloud KMS for BigQuery, GCS, AlloyDB, Pub/Sub, Vertex AI.
- **In transit** — TLS 1.3 everywhere; mTLS for service-to-service.
- **PII handling** — Cloud DLP for redaction before LLM calls and at ingestion.
- **Data residency** — primary region `asia-south1`; DR `asia-southeast1`.
- **Retention** — 7 years for audit log; configurable per data class.

## 4. Secrets

- All secrets in **Secret Manager**; access via Workload Identity, never in env or code.
- Rotation policy: 90 days for API keys, immediate on suspected compromise.

## 5. Audit & Observability

- **Cloud Audit Logs** streamed to BigQuery (append-only).
- **Structured JSON logs** with correlation IDs across services.
- **OpenTelemetry traces** across all services.
- **SLOs** — availability, latency, error rate, model drift; alerts routed to PagerDuty.
- **Tamper-evident** audit trail: append-only BigQuery + Object Versioning on GCS.

## 6. Compliance Standards

| Standard | Coverage |
|---|---|
| PCI-DSS | Tokenized card data, network segmentation, audit logs |
| SOC 2 Type II | Access controls, monitoring, change management |
| ISO 27001 | ISMS, risk assessment, statement of applicability |
| GDPR | Data subject rights, DLP, residency |
| RBI / SEBI | Localization, audit, model explainability |
| HIPAA (insurer) | PHI handling via DLP + dedicated project |

## 7. Threat Model (top 5)

1. **Credential theft** → mitigated by Workload Identity + IAP + Secret Manager.
2. **Data exfiltration via LLM prompt injection** → mitigated by DLP pre-redaction + Vertex AI Search citations + guardrails.
3. **Model inversion / membership inference** → mitigated by differential privacy budget in training + access control.
4. **Insider threat** → mitigated by least-privilege IAM, audit logs, just-in-time access.
5. **Supply chain** → mitigated by Artifact Registry scanning, SLSA provenance, signed images.

## 8. Runbook Cross-refs

- [Model drift runbook](./runbooks/model-drift.md)
- [DocAI failures runbook](./runbooks/docai-failures.md)
- [Fraud spike runbook](./runbooks/fraud-spike.md)
