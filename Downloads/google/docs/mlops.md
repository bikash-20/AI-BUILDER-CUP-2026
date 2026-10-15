# FinGuard AI — MLOps Design

> Distilled MLOps design. Source-of-truth lives in `ml/pipelines/`, `ml/training/`, and the Vertex AI Pipelines code.

## 1. MLOps Stack

| Stage | Tool | Purpose |
|---|---|---|
| Feature engineering | BigQuery + Dataflow | Batch + streaming features |
| Feature store | Vertex AI Feature Store | Point-in-time correctness, online serving |
| Training | Vertex AI Custom Training / BigQuery ML | Model training at scale |
| Experiment tracking | Vertex AI Experiments | Params, metrics, artifacts |
| Hyperparameter tuning | Vertex AI Vizier | Automated optimization |
| Model registry | Vertex AI Model Registry | Central model lifecycle |
| Pipelines | Vertex AI Pipelines (Kubeflow) | End-to-end automation |
| Serving | Vertex AI Endpoints | Real-time + batch prediction |
| Monitoring | Vertex AI Model Monitoring | Drift, skew, latency |

## 2. MLOps Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                    DATA & FEATURE LAYER                          │
│  ┌──────────────┐    ┌──────────────┐    ┌─────────────────┐    │
│  │ Pub/Sub      │───►│ Dataflow     │───►│ Vertex AI       │    │
│  │ (streaming)  │    │ (enrichment) │    │ Feature Store   │    │
│  └──────────────┘    └──────────────┘    └─────────────────┘    │
│  ┌──────────────┐    ┌──────────────┐    ┌─────────────────┐    │
│  │ BigQuery     │───►│ BigQuery ML  │───►│ Training Data   │    │
│  │ (historical) │    │ (features)   │    │ (BigQuery)      │    │
│  └──────────────┘    └──────────────┘    └─────────────────┘    │
├────────────────────────────────────────────────────────────────┤
│                    TRAINING & PIPELINE LAYER                     │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ Vertex AI Pipelines (Kubeflow)                          │    │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │    │
│  │  │ Data     │►│ Feature  │►│ Train    │►│ Evaluate │   │    │
│  │  │ Validate │ │ Eng.     │ │ XGBoost  │ │ Register │   │    │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘   │    │
│  │         │            │            │            │        │    │
│  │         ▼            ▼            ▼            ▼        │    │
│  │  TFDV validation + BigQuery metrics + Drift              │    │
│  └─────────────────────────────────────────────────────────┘    │
│  ┌──────────────────────┐  ┌──────────────────────────────┐     │
│  │ Vertex AI Experiments│  │ Vertex AI Model Registry     │     │
│  └──────────────────────┘  └──────────────────────────────┘     │
├────────────────────────────────────────────────────────────────┤
│                    SERVING & MONITORING LAYER                    │
│  ┌──────────────┐    ┌──────────────┐    ┌─────────────────┐    │
│  │ Vertex AI    │◄───│ Cloud Run    │◄───│ API Gateway     │    │
│  │ Endpoint     │    │ (Go fraud-   │    │ (auth + routing)│    │
│  │ (real-time)  │    │  service)    │    │                 │    │
│  └──────────────┘    └──────────────┘    └─────────────────┘    │
│  ┌──────────────┐    ┌──────────────┐    ┌─────────────────┐    │
│  │ Model        │───►│ Cloud        │───►│ Alerting +      │    │
│  │ Monitoring   │    │ Logging      │    │ Auto-retrain    │    │
│  │ (drift/skew) │    │              │    │ trigger         │    │
│  └──────────────┘    └──────────────┘    └─────────────────┘    │
└────────────────────────────────────────────────────────────────┘
```

## 3. Continuous Training Pipeline

**Trigger sources:**
- Scheduled (weekly)
- Data drift detected (TFDV / Model Monitoring)
- Performance degradation (accuracy < threshold)
- New labeled fraud cases available

**Pipeline steps:**
1. Data Validation (TFDV) → fail on schema violation
2. Feature Engineering → BigQuery SQL + Dataflow
3. Training → XGBoost / Autoencoder on Vertex AI
4. Evaluation → ROC-AUC, Precision@K, Recall
5. Conditional: if metrics > threshold → Register
6. Model Registry → version + alias (e.g., `default`)
7. Deployment → Canary → Full rollout

Feature engineering and training are decoupled — feature pipelines run independently and training pulls from Feature Store with point-in-time correctness.

## 4. Model Monitoring & Drift Detection

| Metric | Tool | Action |
|---|---|---|
| Feature Skew | Vertex AI Model Monitoring | Alert + investigate |
| Training-Serving Skew | TFDV | Auto-trigger retrain |
| Prediction Drift | Model Monitoring | Alert if PSI > 0.2 |
| Latency P95 | Cloud Monitoring | Alert if > 500 ms |
| Error Rate | Cloud Monitoring | Alert if > 1 % |

## 5. Key Design Principles

- **Decouple** feature engineering from training
- **Point-in-time correctness** via Feature Store
- **Validate at every boundary** (TFDV, JSON Schema, checksums)
- **Canary before full** deployment (A/B against baseline)
- **Monitor for drift**, not just errors (PSI, KL divergence)
- **Immutable audit trail** in BigQuery (append-only)

## 6. Implementation Pointers

- Kubeflow pipeline: [`ml/pipelines/fraud_training_pipeline.py`](../ml/pipelines/fraud_training_pipeline.py)
- CT pipeline: [`ml/pipelines/fraud_continuous_training.py`](../ml/pipelines/fraud_continuous_training.py)
- Components: [`ml/pipelines/components/`](../ml/pipelines/components)
- Training: [`ml/training/fraud_xgb/`](../ml/training/fraud_xgb), [`ml/training/fraud_autoencoder/`](../ml/training/fraud_autoencoder)
- Feature engineering: [`dataflow/jobs/txn_enrichment/`](../dataflow/jobs/txn_enrichment)
- Infra (Featurestore, Endpoint): [`terraform/modules/vertex_ai/`](../terraform/modules/vertex_ai)
