# FinGuard Dataflow Jobs

Apache Beam pipelines that run on Google Cloud Dataflow.

| Job | Purpose | Trigger |
|---|---|---|
| `txn_enrichment` | Read `txn-events` Pub/Sub, enrich with BigQuery features, write to `txn-scored` | streaming |
| `doc_intake` | Watch GCS bucket, fan out to Document AI, write to BigQuery | streaming |
| `risk_features_batch` | Nightly batch feature recompute | schedule |

## Run locally

```bash
cd dataflow/jobs/txn_enrichment
MOCK_MODE=true python pipeline.py --runner=DirectRunner
```

## Build a Flex Template

```bash
gcloud dataflow flex-template build \
  gs://finguard-dataflow-templates/txn_enrichment.json \
  --image-gcr-path=asia-south1-docker.pkg.dev/finguard-dev/dataflow/txn-enrichment:latest \
  --sdk-language=python \
  --metadata-file=metadata.json
```
