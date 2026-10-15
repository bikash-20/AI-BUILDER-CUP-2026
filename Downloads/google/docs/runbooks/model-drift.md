# Runbook — Model Drift

> Triggered when Vertex AI Model Monitoring reports PSI > 0.2 on a feature, or training-serving skew detected by TFDV.

## 1. Detect

- Cloud Monitoring: `prediction_drift_psi > 0.2`
- Cloud Monitoring: `feature_skew_psi > 0.2`
- Vertex AI Experiments: AUC drop > 5 % in shadow evaluation

## 2. Triage

1. Identify drifting features via Model Monitoring dashboard
2. Compare serving traffic distribution to training baseline (BigQuery)
3. Check for upstream changes (data contract, schema, new geo)
4. Decide: data issue vs. genuine concept drift

## 3. Mitigate

| Cause | Action |
|---|---|
| Upstream schema change | Roll back upstream; rerun TFDV validation |
| Genuine drift (population shift) | Trigger CT pipeline (`ml/pipelines/fraud_continuous_training.py`) |
| Drift in low-importance feature | No action; document and monitor |
| Drift after data quality incident | Halt inference; engage data team |

## 4. CT Pipeline

```bash
# Submit CT pipeline run
python ml/pipelines/fraud_continuous_training.py \
  --project=$PROJECT --region=asia-south1 \
  --trigger=drift --feature=txn_velocity_24h
```

Pipeline:
1. Validate → 2. Features → 3. Train → 4. Evaluate → 5. Conditional register → 6. Canary deploy

## 5. Rollback

If new model underperforms in canary:
```bash
gcloud ai endpoints undeploy-model ENDPOINT_ID --deployed-model-id=CURRENT \
  --region=asia-south1
# redeploy previous alias
gcloud ai endpoints deploy-model ENDPOINT_ID --model=$PREVIOUS_MODEL \
  --display-name=fraud-model --traffic-split=0=100
```

## 6. Post-incident

- Update monitoring thresholds if needed
- Add drift check to pre-deployment validation
- Add labeled examples to training set
