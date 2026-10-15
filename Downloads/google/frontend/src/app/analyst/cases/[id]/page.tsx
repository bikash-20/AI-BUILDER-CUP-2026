import { CaseTimeline } from "@/components/analyst/CaseTimeline";
import { ShapExplainability } from "@/components/analyst/ShapExplainability";

export default function CaseDetailPage({ params }: { params: { id: string } }) {
  return (
    <section className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div className="lg:col-span-2 space-y-6">
        <header>
          <h1 className="text-2xl font-semibold">Case {params.id}</h1>
          <p className="text-sm text-slate-500">
            Combined alert timeline and analyst notes.
          </p>
        </header>
        <CaseTimeline caseId={params.id} />
      </div>
      <aside className="space-y-6">
        <ShapExplainability />
      </aside>
    </section>
  );
}
