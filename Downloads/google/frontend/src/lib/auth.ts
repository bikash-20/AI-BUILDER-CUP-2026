// Tiny client to the backend.
import type {
  ScoreRequest,
  ScoreResponse,
  RiskResponse,
  ChatResponse,
  Alert,
  AlertsEnvelope,
  CustomerTransactions,
} from "./types";

const base = process.env.NEXT_PUBLIC_API_BASE_URL || "";

async function jget<T>(path: string): Promise<T> {
  const res = await fetch(`${base}${path}`, { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`GET ${path} failed: ${res.status}`);
  }
  return (await res.json()) as T;
}

async function jpost<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${base}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`POST ${path} failed: ${res.status}`);
  }
  return (await res.json()) as T;
}

export const api = {
  scoreTransaction: (req: ScoreRequest) => jpost<ScoreResponse>("/api/score", req),
  getRiskScore: (customerId: string) => jget<RiskResponse>(`/api/risk/${customerId}`),
  sendChat: (message: string, sessionId?: string) =>
    jpost<ChatResponse>("/api/chat", { message, session_id: sessionId }),

  // Live data path: hits the BFF at /api/alerts which proxies to
  // fraud-service GET /v1/alerts. Returns the envelope so callers can
  // detect the `simulated` flag and surface a "simulated data" badge.
  // On any network/HTTP error, falls back to MOCK_ALERTS so the analyst
  // console still renders something during local development. Callers
  // can pass `setUseMocks(false)` to disable the fallback.
  listAlerts: async (
    opts: { useMocks?: boolean } = {}
  ): Promise<AlertsEnvelope> => {
    if (opts.useMocks === false) {
      // Skip fallback; propagate any error to the caller.
      return jget<AlertsEnvelope>("/api/alerts");
    }
    try {
      return await jget<AlertsEnvelope>("/api/alerts");
    } catch {
      return {
        alerts: MOCK_ALERTS,
        count: MOCK_ALERTS.length,
        simulated: true,
        source: "frontend:MOCK_ALERTS",
      };
    }
  },

  // Live data path: hits the BFF at /api/customers/{id}/transactions
  // which proxies to risk-service. Returns the envelope so callers can
  // surface the `simulated` flag. Same fallback semantics as listAlerts.
  listCustomerTransactions: async (
    customerId: string,
    opts: { limit?: number; useMocks?: boolean } = {}
  ): Promise<CustomerTransactions> => {
    const qs = new URLSearchParams();
    if (opts.limit && opts.limit > 0) qs.set("limit", String(opts.limit));

    if (opts.useMocks === false) {
      return jget<CustomerTransactions>(
        `/api/customers/${encodeURIComponent(customerId)}/transactions${qs.toString() ? `?${qs}` : ""}`
      );
    }
    try {
      return await jget<CustomerTransactions>(
        `/api/customers/${encodeURIComponent(customerId)}/transactions${qs.toString() ? `?${qs}` : ""}`
      );
    } catch {
      return {
        customer_id: customerId,
        count: MOCK_TXNS.length,
        simulated: true,
        source: "frontend:MOCK_TXNS",
        transactions: MOCK_TXNS.map((t) => ({ ...t, customer_id: customerId })),
      };
    }
  },
};

// Mock data kept ONLY as a fallback for local dev when the BFF is
// unreachable. The analyst/customer UIs always prefer the live endpoint
// and only render the mock when the live fetch fails.
const MOCK_ALERTS: Alert[] = [
  {
    alert_id: "A-1001",
    txn_id: "T-987",
    customer_id: "C-42",
    severity: "HIGH",
    reason: "unusual_geo_velocity,high_amount_zscore",
    created_at: new Date(Date.now() - 1000 * 60 * 2).toISOString(),
    simulated: true,
  },
  {
    alert_id: "A-1002",
    txn_id: "T-988",
    customer_id: "C-77",
    severity: "MEDIUM",
    reason: "new_device",
    created_at: new Date(Date.now() - 1000 * 60 * 7).toISOString(),
    simulated: true,
  },
];

const MOCK_TXNS = [
  {
    txn_id: "T-1001",
    merchant_id: "M-blue-bottle",
    merchant_category: "FOOD",
    amount: 8.5,
    currency: "USD",
    channel: "CARD" as const,
    timestamp: new Date().toISOString(),
  },
  {
    txn_id: "T-1002",
    merchant_id: "M-whole-foods",
    merchant_category: "GROCERY",
    amount: 64.21,
    currency: "USD",
    channel: "CARD" as const,
    timestamp: new Date(Date.now() - 1000 * 60 * 60).toISOString(),
  },
  {
    txn_id: "T-1003",
    merchant_id: "M-uber",
    merchant_category: "TRANSPORT",
    amount: 22.0,
    currency: "USD",
    channel: "MOBILE" as const,
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 26).toISOString(),
  },
];
