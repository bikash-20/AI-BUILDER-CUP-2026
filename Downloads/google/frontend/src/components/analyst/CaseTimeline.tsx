interface TimelineEvent {
  ts: string;
  actor: string;
  action: string;
}

const MOCK: TimelineEvent[] = [
  { ts: "2026-10-15T09:00:00Z", actor: "system", action: "Alert created" },
  { ts: "2026-10-15T09:02:00Z", actor: "analyst@x", action: "Assigned to self" },
  { ts: "2026-10-15T09:18:00Z", actor: "analyst@x", action: "Reviewed SHAP plot" },
];

export function CaseTimeline({ caseId }: { caseId: string }) {
  return (
    <div className="rounded border border-slate-200 bg-white p-4 shadow-sm">
      <div className="text-sm text-slate-500 mb-3">Case ID: {caseId}</div>
      <ol className="space-y-3">
        {MOCK.map((e, i) => (
          <li key={i} className="flex gap-3">
            <div className="w-1.5 h-1.5 rounded-full bg-brand-500 mt-2" />
            <div>
              <div className="text-sm text-slate-800">{e.action}</div>
              <div className="text-xs text-slate-500">
                {new Date(e.ts).toLocaleString()} · {e.actor}
              </div>
            </div>
          </li>
        ))}
      </ol>
    </div>
  );
}
