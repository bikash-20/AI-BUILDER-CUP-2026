// Formatting helpers.
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/** Combine class names with Tailwind-aware merging. */
export const cn = (...inputs: ClassValue[]): string => twMerge(clsx(inputs));

export const fmtCurrency = (n: number, currency = "INR"): string =>
  new Intl.NumberFormat("en-IN", { style: "currency", currency }).format(n);

export const fmtPercent = (n: number): string =>
  new Intl.NumberFormat("en", { style: "percent", minimumFractionDigits: 1 }).format(n);

export const fmtDate = (iso: string): string => {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString();
};

export const bandToColor = (band: "LOW" | "MEDIUM" | "HIGH" | "CRITICAL"): string => {
  switch (band) {
    case "LOW":
      return "bg-emerald-100 text-emerald-700";
    case "MEDIUM":
      return "bg-amber-100 text-amber-700";
    case "HIGH":
      return "bg-orange-100 text-orange-700";
    case "CRITICAL":
      return "bg-rose-100 text-rose-700";
  }
};
