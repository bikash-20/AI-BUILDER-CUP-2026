-- documents.sql
-- BigQuery DDL for ingested documents (KYC, statements, claims).
-- Document AI extraction results live alongside the GCS URI of the
-- source file; per-field confidences and validators are stored as JSON
-- for flexible downstream validation.

CREATE TABLE IF NOT EXISTS `finguard.documents`
(
  doc_id            STRING    NOT NULL,
  customer_id       STRING    NOT NULL,
  doc_type          STRING    NOT NULL, -- KYC_ID / PROOF_OF_ADDRESS / BANK_STATEMENT / CLAIM_FORM
  gcs_uri           STRING    NOT NULL,
  uploaded_at       TIMESTAMP NOT NULL,
  processed_at      TIMESTAMP,
  status            STRING,              -- QUEUED / PROCESSED / FAILED / NEEDS_REVIEW
  processor_id      STRING,              -- Document AI processor used
  extracted_fields  JSON,                -- { field_name: { value, confidence, bbox } }
  validation_result JSON,                -- { passed, failed_rules: [{rule, message}] }
  error_message     STRING
)
PARTITION BY DATE(uploaded_at)
CLUSTER BY customer_id, doc_type, status
OPTIONS (
  description = "FinGuard AI document intake + Document AI results."
);
