import { DisputeForm } from "@/components/customer/DisputeForm";

export default function CustomerDisputesPage() {
  return (
    <section className="max-w-xl">
      <h1 className="text-2xl font-semibold mb-4">Dispute a transaction</h1>
      <DisputeForm />
    </section>
  );
}
