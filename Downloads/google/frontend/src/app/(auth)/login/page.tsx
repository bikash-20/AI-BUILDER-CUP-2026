export default function LoginPage() {
  return (
    <main className="min-h-screen flex items-center justify-center bg-slate-50">
      <form className="w-full max-w-sm rounded-lg bg-white p-8 shadow">
        <h1 className="text-2xl font-semibold mb-6">Sign in</h1>
        <label className="block text-sm font-medium text-slate-600">Email</label>
        <input
          type="email"
          name="email"
          className="mt-1 block w-full rounded border border-slate-300 px-3 py-2"
        />
        <label className="block mt-4 text-sm font-medium text-slate-600">
          Password
        </label>
        <input
          type="password"
          name="password"
          className="mt-1 block w-full rounded border border-slate-300 px-3 py-2"
        />
        <button
          type="submit"
          className="mt-6 w-full rounded bg-brand-600 px-4 py-2 text-white font-medium hover:bg-brand-700"
        >
          Continue
        </button>
        <p className="mt-4 text-xs text-slate-500">
          (Stub login — wire to your IdP via NextAuth or similar.)
        </p>
      </form>
    </main>
  );
}
