import { Badge } from "@/components/ui/badge";
import type { Alert } from "@/lib/types";

const severityTone = {
  LOW: "neutral",
  MEDIUM: "warning",
  HIGH: "warning",
  CRITICAL: "danger",
} as const;

export function AlertCard({ alert }: { alert: Alert }) {
  return (
    <div className="rounded border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex items-center justify-between">
        <div className="font-mono text-xs text-slate-400">{alert.alert_id}</div>
        <div className="flex items-center gap-2">
          {alert.simulated && (
            <span
              title="This alert is from simulated/seed data, not a live AlloyDB query."
              className="rounded bg-amber-100 px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-amber-800"
            >
              Simulated
            </span>
          )}
          <Badge tone={severityTone[alert.severity]}>{alert.severity}</Badge>
        </div>
      </div>
      <div className="mt-2 text-sm text-slate-700">{alert.reason}</div>
      <div className="mt-2 text-xs text-slate-500">
        {alert.customer_id} · {alert.txn_id} · {new Date(alert.created_at).toLocaleString()}
      </div>
    </div>
  );
}
