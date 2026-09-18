"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Shield,
  LogOut,
  ArrowLeft,
  CreditCard,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import {
  orgApi,
  Subscription,
  statusColor,
  formatPrice,
  formatDate,
} from "@/lib/organizations-api";

export default function OrgSubscriptionsPage() {
  const router = useRouter();
  const { user, initialized, bootstrap, logout } = useAuthStore();

  const [subs, setSubs] = useState<Subscription[]>([]);
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
      const s = await orgApi.listMySubscriptions();
      setSubs(s ?? []);
    } catch (err) {
      console.error("Failed to load subscriptions:", err);
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
            <h1 className="text-2xl font-bold">VPN Services</h1>
            <p className="text-sm text-zinc-400">
              {subs.length} active service{subs.length !== 1 ? "s" : ""}
            </p>
          </div>
        </div>

        {loading ? (
          <div className="text-center text-zinc-400 py-12">Loading...</div>
        ) : subs.length === 0 ? (
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-12 text-center">
            <CreditCard className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
            <h3 className="font-semibold mb-2">No services yet</h3>
            <p className="text-sm text-zinc-400">
              Contact your platform administrator to subscribe.
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {subs.map((sub) => (
              <div
                key={sub.id}
                className="bg-zinc-900 border border-zinc-800 rounded-lg p-5"
              >
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <div className="font-semibold capitalize">
                      {sub.service_type.replace("_", " ")}
                    </div>
                    <div className="text-xs text-zinc-500 mt-0.5">
                      Started {formatDate(sub.started_at)}
                    </div>
                  </div>
                  <span
                    className={`text-xs px-2 py-0.5 rounded border ${statusColor(sub.status)}`}
                  >
                    {sub.status}
                  </span>
                </div>
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div>
                    <div className="text-xs text-zinc-500">Max Users</div>
                    <div className="font-medium">{sub.max_users}</div>
                  </div>
                  <div>
                    <div className="text-xs text-zinc-500">Price</div>
                    <div className="font-medium">
                      {formatPrice(sub.price_cents_per_month)}/mo
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
