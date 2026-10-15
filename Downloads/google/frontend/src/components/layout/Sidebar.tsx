import Link from "next/link";
import { cn } from "@/lib/format";

type Role = "analyst" | "customer";

const ANALYST_LINKS = [
  { href: "/analyst", label: "Live Alerts" },
  { href: "/analyst/risk", label: "Risk" },
  { href: "/analyst/documents", label: "Documents" },
];

const CUSTOMER_LINKS = [
  { href: "/customer", label: "Dashboard" },
  { href: "/customer/chat", label: "Chat" },
  { href: "/customer/disputes", label: "Disputes" },
];

export function Sidebar({ role }: { role: Role }) {
  const links = role === "analyst" ? ANALYST_LINKS : CUSTOMER_LINKS;
  return (
    <aside className="w-60 bg-slate-900 text-slate-100 p-4 hidden md:flex flex-col gap-2">
      <div className="text-xl font-bold mb-6">FinGuard AI</div>
      <nav className="flex flex-col gap-1">
        {links.map((l) => (
          <Link
            key={l.href}
            href={l.href}
            className={cn("rounded px-3 py-2 hover:bg-slate-800")}
          >
            {l.label}
          </Link>
        ))}
      </nav>
    </aside>
  );
}
