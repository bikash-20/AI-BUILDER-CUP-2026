import { TransactionList } from "@/components/customer/TransactionList";

// In a real app this would come from the session / OAuth subject.
// C-1001 matches the seed data served by risk-service in mock mode.
const DEMO_CUSTOMER_ID = "C-1001";

export default function CustomerHome() {
  return (
    <section className="space-y-6">
      <header>
        <h1 className="text-2xl font-semibold">Welcome back</h1>
        <p className="text-sm text-slate-500">
          Your recent activity and account security.
        </p>
      </header>
      <TransactionList customerId={DEMO_CUSTOMER_ID} limit={10} />
    </section>
  );
}
