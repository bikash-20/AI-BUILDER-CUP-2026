import { RoleSwitcher } from "./RoleSwitcher";

export function Topbar() {
  return (
    <header className="h-14 border-b border-slate-200 bg-white flex items-center justify-between px-6">
      <div className="text-sm text-slate-500">Signed in as analyst@finguard.example</div>
      <RoleSwitcher />
    </header>
  );
}
