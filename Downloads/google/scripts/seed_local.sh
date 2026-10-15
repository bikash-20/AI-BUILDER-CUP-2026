#!/usr/bin/env bash
# seed_local.sh — seed local BigQuery emulator with sample data.
#
# Requires the BigQuery emulator running on localhost:9050
# (https://github.com/goccy/bigquery-emulator).

set -euo pipefail

PROJECT_ID="${PROJECT_ID:-finguard-dev}"
DATASET="${DATASET:-finguard_dev}"
EMU_HOST="${EMU_HOST:-http://localhost:9050}"

export BIGQUERY_EMULATOR_HOST="$EMU_HOST"

# Create the dataset (idempotent).
bq --project_id="$PROJECT_ID" mk --dataset --location=us-central1 "$DATASET" 2>/dev/null || true

# Apply each schema file.
SCHEMA_DIR="$(dirname "$0")/../terraform/modules/data/schemas"
for f in "$SCHEMA_DIR"/*.sql; do
  echo "Applying $(basename "$f")..."
  sed "s/finguard\./${DATASET}./g; s/ENV_PLACEHOLDER/dev/g" "$f" | \
    bq --project_id="$PROJECT_ID" query --use_legacy_sql=false
done

# Generate fake transactions in memory and stream into BQ.
python3 - <<'PY'
import datetime as dt
import json
import os
import random
import subprocess
import sys

random.seed(42)
merchants = [
    ("M-001", "Blue Bottle Coffee", "5812"),
    ("M-002", "Whole Foods", "5411"),
    ("M-003", "Uber", "4121"),
    ("M-004", "Apple Store", "5732"),
    ("M-005", "Shell", "5541"),
]
countries = ["US", "GB", "DE", "IN", "BR"]

rows = []
now = dt.datetime.utcnow()
for i in range(1000):
    m = random.choice(merchants)
    rows.append({
        "txn_id": f"T-{i:06d}",
        "customer_id": f"C-{random.randint(0, 999):06d}",
        "amount": round(random.lognormvariate(4, 1), 2),
        "currency": "USD",
        "merchant_id": m[0],
        "merchant_name": m[1],
        "mcc": m[2],
        "device_id": f"D-{random.randint(0, 5000):06d}",
        "ip_address": f"10.{random.randint(0,255)}.{random.randint(0,255)}.{random.randint(0,255)}",
        "geo_country": random.choice(countries),
        "geo_city": "Anytown",
        "event_time": (now - dt.timedelta(minutes=random.randint(0, 60 * 24 * 30))).isoformat() + "Z",
        "source": random.choice(["WEB", "MOBILE", "POS"]),
        "status": "APPROVED",
        "ingested_at": now.isoformat() + "Z",
    })

# Emit each row as a single-row INSERT.
ns = f"{os.environ['DATASET']}.transactions"
sql = f"INSERT INTO `{ns}` (txn_id, customer_id, amount, currency, merchant_id, merchant_name, mcc, device_id, ip_address, geo_country, geo_city, event_time, source, status, ingested_at) VALUES "
values = []
for r in rows:
    values.append(
        f"('{r['txn_id']}','{r['customer_id']}',{r['amount']},'{r['currency']}','{r['merchant_id']}','{r['merchant_name'].replace(chr(39), chr(39)+chr(39))}','{r['mcc']}','{r['device_id']}','{r['ip_address']}','{r['geo_country']}','{r['geo_city']}','{r['event_time']}','{r['source']}','{r['status']}','{r['ingested_at']}')"
    )
# BQ has a 1MB query limit; chunk to 100 rows.
chunk_size = 100
for i in range(0, len(values), chunk_size):
    chunk = ",\n".join(values[i:i+chunk_size])
    full = sql + chunk
    res = subprocess.run(
        ["bq", "--project_id", os.environ["PROJECT_ID"], "query", "--use_legacy_sql=false"],
        input=full, text=True, capture_output=True,
    )
    if res.returncode != 0:
        print("bq failed:", res.stderr, file=sys.stderr)
        sys.exit(1)

print(f"Seeded {len(rows)} transactions.")
PY
