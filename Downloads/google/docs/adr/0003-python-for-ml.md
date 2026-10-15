# ADR-0003 — Python for ML / AI Layer

- **Status:** Accepted
- **Date:** 2026-10-15

## Context

FinGuard's ML/AI workloads (Vertex AI training, Gemini, Document AI, RAG, SHAP, TFDV) are best served by Python where the ecosystem is mature.

## Decision

Use **Python 3.11+** for:
- All training jobs (XGBoost, autoencoder, DocAI fine-tune)
- Vertex AI Pipelines (Kubeflow DSL)
- Document AI and Gemini Cloud Run services
- Apache Beam / Dataflow jobs
- SHAP, fairness, and offline evaluation

Libraries: `google-cloud-aiplatform`, `google-cloud-documentai`, `google-cloud-dlp`, `kfp`, `apache-beam`, `xgboost`, `tensorflow`, `tensorflow-data-validation`, `shap`, `scikit-learn`, `pydantic`, `fastapi`.

## Consequences

- Container sizes are larger (200–500 MB); mitigated by slim base images and `--no-cache-dir`.
- Two language runtimes to monitor; mitigated by shared OTel setup.
- Training jobs can run in Vertex AI Custom Training containers without changing code.
