# alloydb

Provisions the FinGuard AI AlloyDB cluster for OLTP-style workloads
(alert metadata, case state, analyst assignments). AlloyDB is
**optional** in dev; we provision it only in staging and prod where the
workloads justify the cost.

Key properties:

- **Private IP only** (no public endpoint).
- **CMEK encryption** with rotation enabled.
- **High availability** primary + read pool in prod.
- **Continuous backup + point-in-time recovery** enabled.
