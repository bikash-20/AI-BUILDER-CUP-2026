"use client";

import { useChat } from "@/hooks/useChat";

export function ChatWindow() {
  const { messages, send, loading } = useChat();
  return (
    <div className="rounded-lg border border-slate-200 bg-white shadow-sm">
      <div className="h-[420px] overflow-y-auto p-4 space-y-3">
        {messages.length === 0 && (
          <div className="text-sm text-slate-400">
            Ask things like "Why was my card declined?"
          </div>
        )}
        {messages.map((m, i) => (
          <div
            key={i}
            className={
              m.role === "user"
                ? "rounded-lg bg-brand-600 text-white p-3 max-w-[80%] ml-auto"
                : "rounded-lg bg-slate-100 text-slate-900 p-3 max-w-[80%]"
            }
          >
            {m.content}
            {m.citations && m.citations.length > 0 && (
              <div className="mt-2 text-xs opacity-75">
                Sources: {m.citations.join(", ")}
              </div>
            )}
          </div>
        ))}
      </div>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          const fd = new FormData(e.currentTarget);
          const msg = String(fd.get("message") || "");
          if (msg) {
            send(msg);
            e.currentTarget.reset();
          }
        }}
        className="flex gap-2 p-3 border-t border-slate-200"
      >
        <input
          type="text"
          name="message"
          placeholder="Ask a question…"
          className="flex-1 rounded border border-slate-300 px-3 py-2 text-sm"
        />
        <button
          type="submit"
          className="rounded bg-brand-600 px-4 py-2 text-white text-sm font-medium hover:bg-brand-700 disabled:opacity-50"
          disabled={loading}
        >
          {loading ? "…" : "Send"}
        </button>
      </form>
    </div>
  );
}
