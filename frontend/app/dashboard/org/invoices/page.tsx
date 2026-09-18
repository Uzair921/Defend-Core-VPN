"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Shield,
  LogOut,
  ArrowLeft,
  DollarSign,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import {
  orgApi,
  Invoice,
  invoiceStatusColor,
  formatPrice,
  formatDate,
} from "@/lib/organizations-api";

export default function OrgInvoicesPage() {
  const router = useRouter();
  const { user, initialized, bootstrap, logout } = useAuthStore();

  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    bootstrap();
  }, [bootstrap]);

  useEffect(() => {
    if (initialized && !user) router.push("/login");
  }, [initialized, user, router]);

  useEffect(() => {
    if (user) loadData();
  }, [user]);

  async function loadData() {
    setLoading(true);
    try {
      const i = await orgApi.listMyInvoices();
      setInvoices(i ?? []);
    } catch (err) {
      console.error("Failed to load invoices:", err);
    } finally {
      setLoading(false);
    }
  }

  async function handleLogout() {
    await logout();
    router.push("/login");
  }

  if (!initialized || !user) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-zinc-950 text-white">
        <div className="text-zinc-400">Loading...</div>
      </div>
    );
  }

  const totalPaid = invoices
    .filter((i) => i.status === "paid")
    .reduce((sum, i) => sum + i.amount_cents, 0);
  const totalPending = invoices
    .filter((i) => i.status === "pending")
    .reduce((sum, i) => sum + i.amount_cents, 0);

  return (
    <div className="min-h-screen bg-zinc-950 text-white">
      <nav className="border-b border-zinc-800 bg-zinc-900">
        <div className="max-w-5xl mx-auto px-4 py-3 flex items-center justify-between">
          <Link href="/dashboard/org" className="flex items-center gap-2">
            <Shield className="w-6 h-6 text-blue-500" />
            <span className="font-bold text-lg">DefendCore VPN</span>
            <span className="text-xs px-2 py-0.5 rounded bg-blue-950 text-blue-400 border border-blue-800 ml-2">
              Org Admin
            </span>
          </Link>
          <button
            onClick={handleLogout}
            className="flex items-center gap-2 px-3 py-1.5 text-sm bg-zinc-800 hover:bg-zinc-700 rounded"
          >
            <LogOut className="w-4 h-4" />
            Logout
          </button>
        </div>
      </nav>

      <main className="max-w-5xl mx-auto px-4 py-8">
        <div className="flex items-center gap-3 mb-6">
          <Link
            href="/dashboard/org"
            className="p-2 hover:bg-zinc-800 rounded transition"
          >
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold">Invoices</h1>
            <p className="text-sm text-zinc-400">
              {invoices.length} invoice{invoices.length !== 1 ? "s" : ""}
            </p>
          </div>
        </div>

        {/* Summary cards */}
        {invoices.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
              <div className="text-xs text-zinc-500 mb-1">Total Invoices</div>
              <div className="text-xl font-bold">{invoices.length}</div>
            </div>
            <div className="bg-zinc-900 border border-green-900/50 rounded-lg p-4">
              <div className="text-xs text-green-500 mb-1">Total Paid</div>
              <div className="text-xl font-bold text-green-400">
                {formatPrice(totalPaid)}
              </div>
            </div>
            <div className="bg-zinc-900 border border-yellow-900/50 rounded-lg p-4">
              <div className="text-xs text-yellow-500 mb-1">Pending</div>
              <div className="text-xl font-bold text-yellow-400">
                {formatPrice(totalPending)}
              </div>
            </div>
          </div>
        )}

        {loading ? (
          <div className="text-center text-zinc-400 py-12">Loading...</div>
        ) : invoices.length === 0 ? (
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-12 text-center">
            <DollarSign className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
            <h3 className="font-semibold mb-2">No invoices</h3>
            <p className="text-sm text-zinc-400">
              Your billing history will appear here.
            </p>
          </div>
        ) : (
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
            <table className="w-full">
              <thead className="bg-zinc-950/50 border-b border-zinc-800">
                <tr className="text-left text-xs uppercase tracking-wider text-zinc-500">
                  <th className="px-4 py-3">Invoice #</th>
                  <th className="px-4 py-3">Amount</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Period</th>
                  <th className="px-4 py-3">Due</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800">
                {invoices.map((inv) => (
                  <tr key={inv.id} className="hover:bg-zinc-800/50">
                    <td className="px-4 py-3 font-mono text-sm">
                      {inv.invoice_number}
                    </td>
                    <td className="px-4 py-3 font-medium">
                      {formatPrice(inv.amount_cents)}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`text-xs px-2 py-0.5 rounded border ${invoiceStatusColor(
                          inv.status
                        )}`}
                      >
                        {inv.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-zinc-500">
                      {formatDate(inv.period_start)} → {formatDate(inv.period_end)}
                    </td>
                    <td className="px-4 py-3 text-xs text-zinc-500">
                      {formatDate(inv.due_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </main>
    </div>
  );
}
