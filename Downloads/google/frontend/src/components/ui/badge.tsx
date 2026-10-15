import { cn } from "@/lib/format";
import * as React from "react";

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  tone?: "neutral" | "success" | "warning" | "danger";
}

export function Badge({ tone = "neutral", className, ...props }: BadgeProps) {
  const t = {
    neutral: "bg-slate-100 text-slate-700",
    success: "bg-emerald-100 text-emerald-700",
    warning: "bg-amber-100 text-amber-700",
    danger: "bg-rose-100 text-rose-700",
  }[tone];
  return (
    <span
      className={cn(
        "inline-block rounded px-2 py-0.5 text-xs font-medium",
        t,
        className
      )}
      {...props}
    />
  );
}
