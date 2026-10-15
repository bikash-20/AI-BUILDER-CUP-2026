import { Badge } from "@/components/ui/badge";

export function RiskGauge({
  score,
  band,
}: {
  score: number;
  band: "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";
}) {
  const pct = Math.max(0, Math.min(1000, score)) / 10; // 0..100
  const tone =
    band === "CRITICAL" ? "danger" : band === "HIGH" ? "warning" : "neutral";
  return (
    <div>
      <div className="flex items-center justify-between text-sm">
        <span className="text-slate-500">Risk score</span>
        <Badge tone={tone}>{band}</Badge>
      </div>
      <div className="mt-2 h-2 rounded bg-slate-100 overflow-hidden">
        <div
          className="h-full bg-gradient-to-r from-emerald-400 via-amber-400 to-rose-500"
          style={{ width: `${pct}%` }}
        />
      </div>
      <div className="mt-1 text-right text-lg font-semibold">{score}</div>
    </div>
  );
}
