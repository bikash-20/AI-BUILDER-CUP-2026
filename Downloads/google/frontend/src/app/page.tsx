import Link from "next/link";

export default function HomePage() {
  return (
    <main className="min-h-screen flex flex-col items-center justify-center p-8 bg-slate-50">
      <div className="max-w-2xl text-center">
        <h1 className="text-5xl font-extrabold tracking-tight text-brand-700">
          FinGuard AI
        </h1>
        <p className="mt-4 text-lg text-slate-600">
          Real-time fraud detection, risk intelligence, and personalized
          financial guidance — powered by Google Cloud AI.
        </p>
        <div className="mt-10 flex gap-4 justify-center">
          <Link
            href="/analyst"
            className="rounded-md bg-brand-600 px-6 py-3 text-white font-medium shadow hover:bg-brand-700"
          >
            Analyst Console
          </Link>
          <Link
            href="/customer"
            className="rounded-md bg-white border border-slate-300 px-6 py-3 text-slate-700 font-medium hover:bg-slate-100"
          >
            Customer App
          </Link>
        </div>
      </div>
    </main>
  );
}
