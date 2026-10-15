import { cn } from "@/lib/format";
import * as React from "react";

type Variant = "default" | "ghost" | "danger";

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", ...props }, ref) => {
    const classes = {
      default: "bg-brand-600 text-white hover:bg-brand-700",
      ghost: "bg-transparent text-slate-700 hover:bg-slate-100",
      danger: "bg-danger text-white hover:bg-red-700",
    }[variant];
    return (
      <button
        ref={ref}
        className={cn(
          "inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium transition",
          classes,
          className
        )}
        {...props}
      />
    );
  }
);
Button.displayName = "Button";
