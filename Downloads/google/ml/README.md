# FinGuard AI — Python ML & AI Layer

Python 3.11+ for Vertex AI training/serving, Gemini, Document AI, MLOps, and Apache Beam pipelines.

## Layout

```
ml/
├── libs/                # shared libraries (finguard_common, finguard_vertex, ...)
├── services/            # FastAPI Cloud Run services (document_service, gemini_service, ...)
├── training/            # Vertex AI Custom Training entry points
├── pipelines/           # Vertex AI Pipelines (Kubeflow DSL)
└── notebooks/           # EDA / experimentation
```

## Local development

```bash
cd ml
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
pip install -e libs/finguard_common -e libs/finguard_vertex -e libs/finguard_docai -e libs/finguard_gemini -e libs/finguard_dlp -e libs/finguard_evaluation

# Run a service in mock mode
MOCK_MODE=true uvicorn services.document_service.app:app --reload
```

## Testing

```bash
pytest libs/ services/ training/ pipelines/
```

## GCP integration

Set `MOCK_MODE=false` and provide credentials via:

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/sa.json
```

or use Workload Identity in production.
