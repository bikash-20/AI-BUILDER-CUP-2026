# staging environment

This is the **staging** environment. It mirrors prod topology but with
scaled-down resource sizing and a single region (multi-region is added in
prod). Key differences from dev:

- Cloud Run services have `min_instances >= 1` (no cold starts).
- All resources are encrypted with CMEK.
- State is isolated under the `Staging` folder.
- AlloyDB is **not** provisioned (staging uses a Postgres Cloud SQL
  instance for cost reasons — see `modules/alloydb/README.md`).
