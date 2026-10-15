import { NextResponse } from "next/server";

/**
 * BFF proxy for chat-bff POST /v1/chat.
 *
 * Returns a clean 400 when the client body is empty or unparseable.
 */
export async function POST(req: Request) {
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

  const baseUrl = process.env.CHAT_SERVICE_URL || "http://localhost:8082";
  const url = `${baseUrl}/v1/chat`;

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
