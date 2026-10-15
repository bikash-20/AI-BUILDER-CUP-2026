# prod environment

This is the **production** environment. It includes everything dev and
staging have, plus:

- **AlloyDB** cluster with a read pool and PITR.
- **Higher availability targets** (SLOs at 99.9%+).
- **Higher min instances** for Cloud Run to avoid cold starts.
- **Destructive protections** — `deletion_protection = true` on critical
  resources; CI gates require manual approval for `terraform apply`.

`terraform apply` here should always go through the Cloud Build pipeline
with two-person review. Do **not** apply from a laptop.
