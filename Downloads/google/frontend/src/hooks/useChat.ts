"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/auth";

interface Message {
  role: "user" | "assistant";
  content: string;
  citations?: string[];
}

export function useChat(sessionId?: string) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const send = async (msg: string) => {
    setMessages((m) => [...m, { role: "user", content: msg }]);
    setLoading(true);
    setError(null);
    try {
      const r = await api.sendChat(msg, sessionId);
      setMessages((m) => [
        ...m,
        { role: "assistant", content: r.reply, citations: r.citations },
      ]);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  return { messages, loading, error, send };
}
