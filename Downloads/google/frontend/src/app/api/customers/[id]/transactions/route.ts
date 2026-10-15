import { NextResponse } from "next/server";

/**
 * BFF proxy for risk-service GET /v1/customers/{id}/transactions.
 *
 * Validates the customer id before forwarding so a bad URL gets a clean
 * 400 instead of relying on the upstream. Preserves the simulated/source
 * fields so the UI can show a "simulated data" badge.
 */
export async function GET(
  req: Request,
  ctx: { params: { id: string } }
) {
  const id = ctx.params.id?.trim() ?? "";
  if (!id) {
    return NextResponse.json(
      { error: { code: "invalid_request", message: "missing customer id" } },
      { status: 400 }
    );
  }
  // Defensive: the id is interpolated into a URL path, so reject anything
  // that could break out of the segment (whitespace, slash, percent, etc.).
  // Only allow simple, opaque ids matching the FinGuard customer-id format
  // (a single letter "C" followed by 1-32 alphanumerics). The risk-service
  // accepts the same shape and the alloydb seed data uses values like
  // "C-1001", "C-1002".
  if (!/^C[A-Za-z0-9_-]{0,32}$/.test(id)) {
    return NextResponse.json(
      { error: { code: "invalid_request", message: "invalid customer id" } },
      { status: 400 }
    );
  }

  const baseUrl = process.env.RISK_SERVICE_URL || "http://localhost:8081";
  const incoming = new URL(req.url);
  const upstream = new URL(`/v1/customers/${encodeURIComponent(id)}/transactions`, baseUrl);
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
