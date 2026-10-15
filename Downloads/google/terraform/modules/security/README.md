# security

Provides platform-wide security primitives:

- **KMS keyrings + CMEK keys** for BigQuery, GCS, Pub/Sub, AlloyDB, and
  Vertex AI. Resources throughout the platform reference these keys via
  `crypto_key_id` so all data is encrypted with customer-managed keys.
- **VPC Service Controls perimeter** to mitigate data exfiltration. The
  perimeter covers BigQuery, Cloud Storage, Pub/Sub, Vertex AI, and
  Cloud Run.
- **Cloud Armor** security policy with a managed WAF rule set and
  rate-limiting thresholds suitable for the public API Gateway.

> Note: VPC-SC perimeters must be created against an **access policy**.
> The access policy is a one-time manual creation; this module assumes
> it already exists and is referenced via `var.access_policy_id`.
