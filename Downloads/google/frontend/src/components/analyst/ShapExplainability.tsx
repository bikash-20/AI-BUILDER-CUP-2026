interface ShapPoint {
  feature: string;
  value: number; // SHAP contribution (negative is "reduces risk")
}

const MOCK: ShapPoint[] = [
  { feature: "geo_velocity", value: 0.32 },
  { feature: "amount_zscore", value: 0.21 },
  { feature: "device_risk", value: 0.18 },
  { feature: "merchant_risk", value: -0.05 },
  { feature: "time_of_day", value: -0.11 },
];

export function ShapExplainability() {
  const max = Math.max(...MOCK.map((m) => Math.abs(m.value)));
  return (
    <div className="rounded border border-slate-200 bg-white p-4 shadow-sm">
      <h2 className="text-sm font-semibold text-slate-700">SHAP explanation</h2>
      <p className="text-xs text-slate-500 mb-3">Contribution to predicted fraud probability.</p>
      <ul className="space-y-2">
        {MOCK.map((m) => (
          <li key={m.feature} className="text-sm">
            <div className="flex justify-between">
              <span className="font-mono">{m.feature}</span>
              <span
                className={
                  m.value >= 0 ? "text-rose-600" : "text-emerald-600"
                }
              >
                {m.value.toFixed(3)}
              </span>
            </div>
            <div className="h-2 rounded bg-slate-100 overflow-hidden mt-1">
              <div
                className={m.value >= 0 ? "h-full bg-rose-500" : "h-full bg-emerald-500"}
                style={{
                  width: `${(Math.abs(m.value) / max) * 100}%`,
                  marginLeft: m.value < 0 ? `${100 - (Math.abs(m.value) / max) * 100}%` : undefined,
                }}
              />
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
