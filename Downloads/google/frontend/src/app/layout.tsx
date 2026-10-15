import type { Metadata } from "next";
import "../styles/globals.css";

export const metadata: Metadata = {
  title: "FinGuard AI",
  description:
    "Real-time fraud detection, risk intelligence, and personalized financial guidance.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>{children}</body>
    </html>
  );
}
