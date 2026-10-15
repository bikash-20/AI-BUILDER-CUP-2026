// k6 load test for the /v1/score endpoint.
//
// Usage:
//   k6 run \
//     -e BASE_URL=https://gateway.dev.finguard.example.com \
//     -e TOKEN=eyJ... \
//     scripts/load_test/k6_score.js
//
// The script ramps to 200 RPS over 60s and asserts the p95 latency
// stays under 450ms — the SLO from terraform/modules/monitoring/.

import http from "k6/http";
import { check } from "k6";
import { Rate, Trend } from "k6/metrics";
import { SharedArray } from "k6/data";

const errorRate = new Rate("errors");
const scoreLatency = new Trend("score_latency_ms", true);

const customers = new SharedArray("customers", function () {
  // In a real test we'd seed this from BigQuery; for the smoke test
  // we generate 1000 fixed IDs.
  const out = [];
  for (let i = 0; i < 1000; i++) {
    out.push(`C-${i.toString().padStart(6, "0")}`);
  }
  return out;
});

export const options = {
  stages: [
    { duration: "30s", target: 50 },  // warm up
    { duration: "30s", target: 200 }, // ramp
    { duration: "60s", target: 200 }, // hold
    { duration: "30s", target: 0 },   // cool down
  ],
  thresholds: {
    http_req_failed: ["rate<0.01"],         // <1% errors
    "score_latency_ms": ["p(95)<450"],      // SLO budget
    errors: ["rate<0.01"],
  },
};

export default function () {
  const url = `${__ENV.BASE_URL}/v1/score`;
  const headers = {
    "Content-Type": "application/json",
    Authorization: `Bearer ${__ENV.TOKEN}`,
  };

  const payload = JSON.stringify({
    txn_id: `T-${__VU}-${__ITER}`,
    customer_id: customers[Math.floor(Math.random() * customers.length)],
    amount: Math.random() * 1000,
    currency: "USD",
    merchant_id: "M-1",
    merchant_name: "Acme",
    event_time: new Date().toISOString(),
    source: "WEB",
  });

  const res = http.post(url, payload, { headers, tags: { name: "score" } });

  scoreLatency.add(res.timings.duration);
  errorRate.add(res.status !== 200);

  check(res, {
    "status 200": (r) => r.status === 200,
    "decision present": (r) => r.json("decision") !== undefined,
  });
}
