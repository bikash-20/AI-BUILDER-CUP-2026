"use client";

import { useEffect } from "react";
import { useAlerts } from "@/hooks/useAlerts";
import { AlertCard } from "./AlertCard";
import { defaultAlertStream } from "@/lib/ws";

export function AlertQueue() {
  const { alerts, addAlert, loading, error, simulated } = useAlerts();

  useEffect(() => {
    const stream = defaultAlertStream();
    stream.connect();
    const unsub = stream.subscribe((msg) => {
      addAlert({
        alert_id: msg.alert_id,
        txn_id: msg.txn_id,
        customer_id: msg.customer_id,
        severity: msg.severity,
        reason: `live_score=${msg.score.toFixed(2)}`,
        created_at: new Date().toISOString(),
      });
    });
    return () => {
      unsub();
      stream.close();
    };
  }, [addAlert]);

  return (
    <div className="space-y-3">
      {simulated && (
        <div
          role="status"
          className="rounded border border-amber-300 bg-amber-50 px-3 py-2 text-xs text-amber-800"
        >
          Showing simulated data — backend returned no live alerts.
        </div>
      )}

      {error && (
        <div
          role="alert"
          className="rounded border border-rose-300 bg-rose-50 px-3 py-2 text-xs text-rose-800"
        >
          Failed to load alerts: {error}
        </div>
      )}

      {loading && !alerts.length ? (
        <div
          role="status"
          className="rounded border border-dashed border-slate-300 bg-white p-12 text-center text-slate-500"
        >
          Loading alerts…
        </div>
      ) : !alerts.length ? (
        <div className="rounded border border-dashed border-slate-300 bg-white p-12 text-center text-slate-500">
          No alerts. Listening…
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          {alerts.map((a) => (
            <AlertCard key={a.alert_id} alert={a} />
          ))}
        </div>
      )}
    </div>
  );
}
