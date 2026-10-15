import { NextResponse } from "next/server";

/**
 * BFF proxy for fraud-service GET /v1/alerts.
 *
 * Returns the upstream JSON untouched (preserving the simulated/source
 * fields so the UI can show a "simulated data" badge). Forwards the
 * upstream status code for 4xx/5xx responses, and converts network
 * failures into a clean 502 with a structured error body.
 */
export async function GET(req: Request) {
  const baseUrl = process.env.FRAUD_SERVICE_URL || "http://localhost:8080";

  // Forward query params (customer_id, limit) to the upstream so the
  // fraud-service can apply the filter. Strip nothing — the Go service
  // is responsible for rejecting unknown params.
  const incoming = new URL(req.url);
  const upstream = new URL("/v1/alerts", baseUrl);
  incoming.searchParams.forEach((v, k) => upstream.searchParams.set(k, v));

  const controller = new AbortController();
  const t = setTimeout(() => controller.abort(), 3000);
  try {
    const res = await fetch(upstream.toString(), {
      method: "GET",
      headers: { Accept: "application/json" },
      cache: "no-store",
      signal: controller.signal,
    });
    const text = await res.text();
    const data = text ? safeJSON(text) : {};
    return NextResponse.json(data, { status: res.status });
  } catch (e) {
    return NextResponse.json(
      { error: { code: "upstream", message: String(e) } },
      { status: 502 }
    );
  } finally {
    clearTimeout(t);
  }
}

function safeJSON(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return { error: { code: "bad_upstream", message: "non-JSON response" } };
  }
}
