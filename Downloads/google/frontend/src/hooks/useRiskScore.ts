"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/auth";
import type { RiskResponse } from "@/lib/types";

export function useRiskScore(customerId: string | null) {
  const [risk, setRisk] = useState<RiskResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(false);

  useEffect(() => {
    if (!customerId) return;
    let cancelled = false;
    setLoading(true);
    api
      .getRiskScore(customerId)
      .then((r) => !cancelled && setRisk(r))
      .catch((e) => !cancelled && setError(String(e)))
      .finally(() => !cancelled && setLoading(false));
    return () => {
      cancelled = true;
    };
  }, [customerId]);

  return { risk, error, loading };
}
