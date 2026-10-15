"use client";

import { useState } from "react";

export function DisputeForm() {
  const [submitted, setSubmitted] = useState(false);
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        setSubmitted(true);
      }}
      className="rounded border border-slate-200 bg-white p-6 shadow-sm space-y-4"
    >
      <div>
        <label className="text-sm font-medium text-slate-600" htmlFor="txn">
          Transaction ID
        </label>
        <input
          name="txn"
          id="txn"
          className="mt-1 block w-full rounded border border-slate-300 px-3 py-2"
        />
      </div>
      <div>
        <label className="text-sm font-medium text-slate-600" htmlFor="reason">
          Reason
        </label>
        <textarea
          name="reason"
          id="reason"
          rows={4}
          className="mt-1 block w-full rounded border border-slate-300 px-3 py-2"
        />
      </div>
      <button
        type="submit"
        className="rounded bg-brand-600 px-4 py-2 text-white text-sm font-medium hover:bg-brand-700"
      >
        Submit dispute
      </button>
      {submitted && (
        <p className="rounded bg-emerald-50 text-emerald-700 px-3 py-2 text-sm">
          Dispute submitted. We&apos;ll email you within 2 business days.
        </p>
      )}
    </form>
  );
}
