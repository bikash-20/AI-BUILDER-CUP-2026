import { NextResponse } from "next/server";

/**
 * BFF proxy for fraud-service POST /v1/score.
 *
 * Returns a clean 400 when the client body is empty or unparseable, and
 * forwards the upstream status / JSON for everything else.
 */
export async function POST(req: Request) {
  // 1. Read the body as text first so we can return 400 on bad JSON
  //    rather than letting Next.js turn a SyntaxError into a 500.
  let raw: string;
  try {
    raw = await req.text();
  } catch {
    return NextResponse.json(
      { error: { code: "invalid_request", message: "unable to read body" } },
      { status: 400 }
    );
  }
  if (!raw || raw.trim() === "") {
    return NextResponse.json(
      { error: { code: "invalid_request", message: "empty body" } },
      { status: 400 }
    );
  }

  let body: unknown;
  try {
    body = JSON.parse(raw);
  } catch {
    return NextResponse.json(
      { error: { code: "invalid_request", message: "body is not valid JSON" } },
      { status: 400 }
    );
  }

  const baseUrl = process.env.FRAUD_SERVICE_URL || "http://localhost:8080";
  const url = `${baseUrl}/v1/score`;

  try {
    const res = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      cache: "no-store",
    });
    const data = await res.json();
    return NextResponse.json(data, { status: res.status });
  } catch (e) {
    return NextResponse.json(
      { error: { code: "upstream", message: String(e) } },
      { status: 502 }
    );
  }
}
