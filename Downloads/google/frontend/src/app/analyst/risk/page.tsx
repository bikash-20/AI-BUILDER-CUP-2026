import { RiskGauge } from "@/components/analyst/RiskGauge";

const MOCK: Array<{ customer_id: string; score: number; band: "LOW" | "MEDIUM" | "HIGH" | "CRITICAL" }> = [
  { customer_id: "C-001", score: 920, band: "CRITICAL" },
  { customer_id: "C-002", score: 760, band: "HIGH" },
  { customer_id: "C-003", score: 410, band: "MEDIUM" },
  { customer_id: "C-004", score: 130, band: "LOW" },
];

export default function RiskPage() {
  return (
    <section>
      <h1 className="text-2xl font-semibold mb-4">Portfolio Risk</h1>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {MOCK.map((row) => (
          <div key={row.customer_id} className="rounded bg-white p-4 shadow">
            <RiskGauge score={row.score} band={row.band} />
            <div className="mt-2 text-sm text-slate-500">{row.customer_id}</div>
          </div>
        ))}
      </div>
    </section>
  );
}
