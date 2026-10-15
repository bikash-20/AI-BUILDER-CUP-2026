export default function DocumentsPage() {
  return (
    <section>
      <h1 className="text-2xl font-semibold mb-4">Document Review Queue</h1>
      <p className="text-slate-500">
        Documents flagged for human review because Document AI confidence was
        below 0.9 or extraction failed.
      </p>
      <div className="mt-4 rounded bg-white shadow">
        <table className="w-full text-sm">
          <thead className="text-left text-slate-500 border-b">
            <tr>
              <th className="px-4 py-2">Doc ID</th>
              <th className="px-4 py-2">Customer</th>
              <th className="px-4 py-2">Type</th>
              <th className="px-4 py-2">Confidence</th>
              <th className="px-4 py-2">Status</th>
            </tr>
          </thead>
          <tbody>
            <tr className="border-b last:border-0">
              <td className="px-4 py-3">DOC-123</td>
              <td className="px-4 py-3">C-001</td>
              <td className="px-4 py-3">KYC</td>
              <td className="px-4 py-3">0.81</td>
              <td className="px-4 py-3 text-amber-600">Needs Review</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  );
}
