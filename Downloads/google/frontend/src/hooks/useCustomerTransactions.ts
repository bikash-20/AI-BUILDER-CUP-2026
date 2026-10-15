"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/auth";
import type { Transaction } from "@/lib/types";

export interface UseCustomerTransactionsState {
  transactions: Transaction[];
  loading: boolean;
  error: string | null;
  /** True when we're rendering the embedded fallback (BFF unreachable). */
  simulated: boolean;
  /** Re-trigger the fetch. */
  refresh: () => void;
}

/**
 * Fetches the recent transactions for a customer through the BFF.
 * Returns loading / error / simulated flags so the consumer can render
 * appropriate states.
 */
export function useCustomerTransactions(
  customerId: string,
  opts: { limit?: number; useMocks?: boolean } = {}
): UseCustomerTransactionsState {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [simulated, setSimulated] = useState(false);
  const [tick, setTick] = useState(0);

  useEffect(() => {
    if (!customerId) {
      setTransactions([]);
      setLoading(false);
      setError("missing customer id");
      return;
    }
    let cancelled = false;
    setLoading(true);
    setError(null);
    api
      .listCustomerTransactions(customerId, {
        limit: opts.limit,
        useMocks: opts.useMocks,
      })
      .then((env) => {
        if (cancelled) return;
        setTransactions(env.transactions ?? []);
        setSimulated(Boolean(env.simulated));
        setLoading(false);
      })
      .catch((e) => {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : String(e));
        setLoading(false);
      });
    return () => {
      cancelled = true;
    };
    // We deliberately exclude opts.limit / opts.useMocks from deps so
    // callers can re-render freely without refetching. The fetch is
    // triggered by `customerId` or the manual `refresh` tick.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [customerId, tick]);

  return {
    transactions,
    loading,
    error,
    simulated,
    refresh: () => setTick((n) => n + 1),
  };
}
