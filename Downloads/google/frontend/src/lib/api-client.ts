// Shared TypeScript types.

export type Decision = "APPROVE" | "REVIEW" | "BLOCK";
export type RiskBand = "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";
export type Severity = "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";

export interface Geo {
  lat: number;
  lon: number;
}

export interface ScoreRequest {
  txn_id: string;
  customer_id: string;
  amount: number;
  currency: string;
  merchant_id: string;
  geo: Geo;
  device_id?: string;
  channel: "CARD" | "MOBILE" | "WEB" | "ATM" | "UPI";
  timestamp: string;
}

export interface ScoreResponse {
  decision: Decision;
  fraud_probability: number;
  risk_score: number;
  reasons: string[];
  model_version: string;
}

export interface RiskResponse {
  customer_id: string;
  score: number;
  band: RiskBand;
  breakdown: {
    fraud: number;
    credit: number;
    behavioral: number;
    compliance: number;
  };
  updated_at: string;
}

export interface ChatResponse {
  session_id: string;
  reply: string;
  citations: string[];
  actions: string[];
}

export interface Alert {
  alert_id: string;
  txn_id: string;
  customer_id: string;
  severity: Severity;
  reason: string;
  created_at: string;
}
