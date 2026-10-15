"use client";

import { fmtCurrency, fmtDate } from "@/lib/format";
import { useCustomerTransactions } from "@/hooks/useCustomerTransactions";

interface TransactionListProps {
  /** Customer to fetch transactions for. Defaults to C-1001 to match
   *  the deterministic seed data from risk-service. */
  customerId?: string;
  /** Optional cap on the number of rows. Defaults to 10. */
  limit?: number;
}

/**
 * Recent transactions for a customer. Fetches via the BFF at
 * /api/customers/{id}/transactions. Falls back to deterministic seed
 * data when the BFF is unreachable so the customer dashboard always
 * renders something during local development.
 *
 * Loading, empty, error, and "simulated data" states are all surfaced
 * explicitly per the application design.
 */
export function TransactionList({
  customerId = "C-1001",
  limit = 10,
}: TransactionListProps = {}) {
  const { transactions, loading, error, simulated } =
    useCustomerTransactions(customerId, { limit });

  return (
    <div className="space-y-3">
      {simulated && (
        <div
          role="status"
          className="rounded border border-amber-300 bg-amber-50 px-3 py-2 text-xs text-amber-800"
        >
          Showing simulated data — backend unreachable, using embedded fallback.
        </div>
      )}

      {error && (
        <div
          role="alert"
          className="rounded border border-rose-300 bg-rose-50 px-3 py-2 text-xs text-rose-800"
        >
          Failed to load transactions: {error}
        </div>
      )}

      <div className="rounded border border-slate-200 bg-white shadow-sm overflow-hidden">
        <table className="w-full text-sm">
          <thead className="text-left text-slate-500 bg-slate-50">
            <tr>
              <th className="px-4 py-2">Merchant</th>
              <th className="px-4 py-2">Amount</th>
              <th className="px-4 py-2">When</th>
            </tr>
          </thead>
          <tbody>
            {loading && transactions.length === 0 ? (
              <tr>
                <td colSpan={3} className="px-4 py-8 text-center text-slate-500">
                  Loading transactions…
                </td>
              </tr>
            ) : transactions.length === 0 ? (
              <tr>
                <td colSpan={3} className="px-4 py-8 text-center text-slate-500">
                  No transactions for {customerId}.
                </td>
              </tr>
            ) : (
              transactions.map((t) => (
                <tr key={t.txn_id} className="border-t border-slate-100">
                  <td className="px-4 py-3">
                    <div className="font-medium text-slate-800">{t.merchant_id}</div>
                    {t.merchant_category && (
                      <div className="text-xs text-slate-500">{t.merchant_category}</div>
                    )}
                  </td>
                  <td className="px-4 py-3">{fmtCurrency(t.amount, t.currency)}</td>
                  <td className="px-4 py-3">{fmtDate(t.timestamp)}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
