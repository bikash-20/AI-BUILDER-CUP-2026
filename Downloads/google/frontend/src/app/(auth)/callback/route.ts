import { NextResponse } from "next/server";

export async function GET(request: Request) {
  // In production: exchange code for tokens, set session cookie, redirect to /analyst.
  // Stub: just redirect.
  return NextResponse.redirect(new URL("/analyst", request.url));
}
