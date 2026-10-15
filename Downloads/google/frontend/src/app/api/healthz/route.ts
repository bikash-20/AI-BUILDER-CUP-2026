import { NextResponse } from "next/server";

/**
 * BFF health endpoint. Pings each downstream Go service and returns a
 * per-service status. Useful for /api/healthz smoke-tests.
 */
export async function GET() {
  const targets: Array<[string, string]> = [
    ["fraud", process.env.FRAUD_SERVICE_URL || "http://localhost:8080"],
    ["risk", process.env.RISK_SERVICE_URL || "http://localhost:8081"],
    ["chat", process.env.CHAT_SERVICE_URL || "http://localhost:8082"],
    ["document", process.env.DOCUMENT_SERVICE_URL || "http://localhost:8083"],
    ["compliance", process.env.COMPLIANCE_SERVICE_URL || "http://localhost:8084"],
  ];

  const results = await Promise.all(
    targets.map(async ([name, base]) => {
      const controller = new AbortController();
      const t = setTimeout(() => controller.abort(), 1500);
      try {
        const res = await fetch(`${base}/healthz`, {
          cache: "no-store",
          signal: controller.signal,
        });
        const ok = res.ok;
        return [name, { ok, status: res.status, url: base }];
      } catch (e) {
        return [name, { ok: false, error: String(e), url: base }];
      } finally {
        clearTimeout(t);
      }
    })
  );

  const allOk = results.every(([, v]) => (v as { ok: boolean }).ok);
  return NextResponse.json(
    {
      status: allOk ? "ok" : "degraded",
      services: Object.fromEntries(results),
    },
    { status: allOk ? 200 : 503 }
  );
}
