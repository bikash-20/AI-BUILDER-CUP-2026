import { AlertQueue } from "@/components/analyst/AlertQueue";

export default function AnalystHome() {
  return (
    <section>
      <header className="mb-6">
        <h1 className="text-2xl font-semibold">Live Alerts</h1>
        <p className="text-sm text-slate-500">
          Real-time fraud alerts streamed from the alert-worker service.
        </p>
      </header>
      <AlertQueue />
    </section>
  );
}
