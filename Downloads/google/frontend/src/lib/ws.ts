// WebSocket subscriber for live alerts.

"use client";

export interface WsAlertMessage {
  alert_id: string;
  txn_id: string;
  customer_id: string;
  severity: "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";
  score: number;
}

export class AlertStream {
  private ws: WebSocket | null = null;
  private listeners = new Set<(m: WsAlertMessage) => void>();
  private url: string;

  constructor(url: string) {
    this.url = url;
  }

  connect(): void {
    if (this.ws) return;
    if (typeof window === "undefined") return;
    try {
      this.ws = new WebSocket(this.url);
      this.ws.onmessage = (ev) => {
        try {
          const data = JSON.parse(ev.data) as WsAlertMessage;
          this.listeners.forEach((cb) => cb(data));
        } catch {
          // ignore
        }
      };
      this.ws.onerror = () => {
        // Reconnect after a delay
        setTimeout(() => this.connect(), 5000);
      };
    } catch {
      setTimeout(() => this.connect(), 5000);
    }
  }

  subscribe(cb: (m: WsAlertMessage) => void): () => void {
    this.listeners.add(cb);
    return () => this.listeners.delete(cb);
  }

  close(): void {
    this.ws?.close();
    this.ws = null;
    this.listeners.clear();
  }
}

export const defaultAlertStream = () =>
  new AlertStream(process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8083/ws/alerts");
