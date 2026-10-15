"use client";

import { useRouter } from "next/navigation";

export function RoleSwitcher() {
  const router = useRouter();
  return (
    <div className="flex items-center gap-2">
      <button
        onClick={() => router.push("/analyst")}
        className="rounded border border-slate-300 px-3 py-1 text-sm hover:bg-slate-100"
      >
        Analyst
      </button>
      <button
        onClick={() => router.push("/customer")}
        className="rounded border border-slate-300 px-3 py-1 text-sm hover:bg-slate-100"
      >
        Customer
      </button>
    </div>
  );
}
