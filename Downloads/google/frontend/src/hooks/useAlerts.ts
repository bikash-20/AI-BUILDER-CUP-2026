"use client";

import { useCallback, useEffect, useState } from "react";
import { api } from "@/lib/auth";
import type { Alert } from "@/lib/types";

export interface UseAlertsState {
  alerts: Alert[];
  loading: boolean;
  /** Set when the BFF returned a non-2xx response. The fallback path
   *  in api.listAlerts will swallow network errors and return mock data;
   *  this only fires when callers opt out of the fallback. */
  error: string | null;
  /** True when the live data path is unreachable and we're showing the
   *  embedded fallback so the UI can surface a "simulated data" badge. */
  simulated: boolean;
  /** Re-trigger the fetch (e.g. after a manual refresh). */
  refresh: () => void;
  /** Append a single alert to the in-memory list (used by the WebSocket
   *  subscriber to prepend live alerts to the historical list). */
  addAlert: (a: Alert) => void;
}

export function useAlerts(opts: { useMocks?: boolean } = {}): UseAlertsState {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [simulated, setSimulated] = useState(false);
  const [tick, setTick] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    api
      .listAlerts({ useMocks: opts.useMocks })
      .then((env) => {
        if (cancelled) return;
        setAlerts(env.alerts ?? []);
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
  }, [tick, opts.useMocks]);

  const addAlert = useCallback((a: Alert) => {
    setAlerts((prev) => [a, ...prev].slice(0, 100));
  }, []);

  const refresh = useCallback(() => setTick((n) => n + 1), []);

  return { alerts, loading, error, simulated, refresh, addAlert };
}
