# cloud_run

Provisions the FinGuard AI Cloud Run services for each Go and Python
microservice. Services are intentionally **not** deployed from source by
Terraform — they're deployed by **Cloud Deploy** (see `deploy/clouddeploy/`).
This module just creates the **Service** resources, IAM bindings, and
Cloud Deploy targets.

Each service:

- Runs with a dedicated service account from the `security` module
  (least privilege).
- Uses **direct VPC egress** so it can reach BigQuery, Pub/Sub, Vertex
  AI, and AlloyDB over private IP.
- Sets an explicit `max_instance_request_concurrency` to bound latency
  under load (matches Go service concurrency settings).
- Reads configuration from env vars injected from Secret Manager.
