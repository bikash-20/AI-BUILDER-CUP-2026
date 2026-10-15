# vertex_ai

Provisions the FinGuard AI ML platform on Vertex AI:

- **Feature Store** (`featurestore`) with an `entity_type` per major
  feature group (customer, transaction, device).
- **Online + batch serving** via dedicated feature store serving configs.
- **Model Registry** (implicit) — Vertex AI automatically populates the
  registry as models are uploaded; this module doesn't create registry
  resources directly.
- **Endpoint** with a placeholder deployed model; production deployments
  are managed by the Vertex AI pipeline (`ml/pipelines/fraud_training_pipeline.py`).
- **Model Monitoring** job that watches the endpoint for
  feature/label/score drift and publishes alerts to Cloud Monitoring.

All resources are encrypted with the Vertex AI CMEK from the `security`
module and run inside the VPC created by the `network` module.
