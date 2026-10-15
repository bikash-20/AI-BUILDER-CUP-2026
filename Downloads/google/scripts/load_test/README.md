# Load testing

The `k6_score.js` script stress-tests the `/v1/score` endpoint at 200 RPS
and asserts the SLO (p95 < 450ms, error rate < 1%).

## Running

```bash
# Local stack (port-forwarded API Gateway):
k6 run \
  -e BASE_URL=http://localhost:8080 \
  -e TOKEN=dev-token \
  scripts/load_test/k6_score.js

# Staging:
k6 run \
  -e BASE_URL=https://gateway.staging.finguard.example.com \
  -e TOKEN=$(gcloud auth print-identity-token) \
  scripts/load_test/k6_score.js
```

## Why k6?

- JavaScript-based (matches the frontend team).
- Native support for thresholds / SLO assertions.
- Streams results to a `cloud` output for centralized dashboards.
- Easy to extend with `xk6` plugins for specific protocols.

## Adding new tests

Create a new `*.js` file in this directory; copy the structure of
`k6_score.js`. Add a corresponding target to the top-level `Makefile`.
