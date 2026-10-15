#!/usr/bin/env python3
"""seed_local_http.py — seed the goccy BigQuery emulator via HTTP.

This is a simpler alternative to scripts/seed_local.sh that does not
require the `bq` CLI to be installed; it talks to the emulator's REST
endpoint directly. Run with:

    BIGQUERY_EMULATOR_HOST=http://127.0.0.1:9050 python3 scripts/seed_local_http.py
"""

import datetime as dt
import json
import os
import random
import sys
import urllib.request

random.seed(42)
HOST = os.environ.get("BIGQUERY_EMULATOR_HOST", "http://127.0.0.1:9050")
PROJECT = os.environ.get("PROJECT_ID", "finguard-dev")
DATASET = os.environ.get("DATASET", "finguard_dev")

def query(sql: str, allow_dup: bool = False) -> dict:
    body = json.dumps({"query": sql}).encode()
    req = urllib.request.Request(
        f"{HOST}/bigquery/v2/projects/{PROJECT}/queries",
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            return json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        msg = e.read().decode()
        if allow_dup and ("duplicate" in msg or "already exists" in msg):
            return {}
        print(f"[seed] HTTP {e.code}: {msg}\n  SQL: {sql[:200]}", file=sys.stderr)
        raise

# 1. Create the (simplified) transactions table. The goccy emulator
# does not honor `IF NOT EXISTS`; tolerate the duplicate error so the
# script can be re-run idempotently.
print(f"[seed] CREATE TABLE {DATASET}.transactions on {HOST}")
query(f"""
CREATE TABLE {DATASET}.transactions (
  txn_id        STRING,
  customer_id   STRING,
  amount        NUMERIC,
  currency      STRING,
  merchant_id   STRING,
  merchant_name STRING,
  mcc           STRING,
  device_id     STRING,
  ip_address    STRING,
  geo_country   STRING,
  event_time    TIMESTAMP,
  source        STRING
)
""", allow_dup=True)

# 2. Build a batch of fake transactions.
merchants = [
    ("M-001", "Blue Bottle Coffee", "5812"),
    ("M-002", "Whole Foods", "5411"),
    ("M-003", "Uber", "4121"),
    ("M-004", "Apple Store", "5732"),
    ("M-005", "Shell", "5541"),
]
countries = ["US", "GB", "DE", "IN", "BR"]
sources = ["WEB", "MOBILE", "POS"]

now = dt.datetime.utcnow()
rows = []
# Make sure C-000000 has some transactions for the live path test.
for i in range(50):
    m = random.choice(merchants)
    rows.append((
        f"T-{i:06d}", "C-000000", round(random.lognormvariate(4, 1), 2), "USD",
        m[0], m[1], m[2], f"D-{random.randint(0, 5000):06d}",
        f"10.{random.randint(0,255)}.{random.randint(0,255)}.{random.randint(0,255)}",
        random.choice(countries),
        (now - dt.timedelta(minutes=random.randint(0, 60 * 24 * 30))).isoformat() + "Z",
        random.choice(sources),
    ))
# Plus a general pool for any other customer id.
for i in range(200):
    m = random.choice(merchants)
    rows.append((
        f"T-{i+50:06d}", f"C-{random.randint(0, 999):06d}",
        round(random.lognormvariate(4, 1), 2), "USD",
        m[0], m[1], m[2], f"D-{random.randint(0, 5000):06d}",
        f"10.{random.randint(0,255)}.{random.randint(0,255)}.{random.randint(0,255)}",
        random.choice(countries),
        (now - dt.timedelta(minutes=random.randint(0, 60 * 24 * 30))).isoformat() + "Z",
        random.choice(sources),
    ))

print(f"[seed] inserting {len(rows)} transactions")
# Chunk to keep query size manageable.
chunk = 50
for i in range(0, len(rows), chunk):
    batch = rows[i : i + chunk]
    values = ", ".join(
        f"('{r[0]}','{r[1]}',{r[2]},'{r[3]}','{r[4]}','{r[5]}','{r[6]}','{r[7]}','{r[8]}','{r[9]}','{r[10]}','{r[11]}')"
        for r in batch
    )
    sql = f"INSERT INTO `{DATASET}.transactions` (txn_id, customer_id, amount, currency, merchant_id, merchant_name, mcc, device_id, ip_address, geo_country, event_time, source) VALUES {values}"
    res = query(sql)
    if res.get("errors"):
        print(f"[seed] chunk {i} failed: {res['errors']}", file=sys.stderr)
        sys.exit(1)

print(f"[seed] done ({len(rows)} rows)")
