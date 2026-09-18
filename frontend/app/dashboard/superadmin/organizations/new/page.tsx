"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Shield,
  LogOut,
  ArrowLeft,
  Building2,
  CheckCircle,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import {
  organizationsApi,
  CreateOrganizationInput,
  OrgPlan,
  planColor,
} from "@/lib/organizations-api";

const PLANS: { value: OrgPlan; label: string; description: string; price: number; maxUsers: number }[] = [
  { value: "trial", label: "Trial", description: "14-day free trial", price: 0, maxUsers: 5 },
  { value: "basic", label: "Basic", description: "Small teams", price: 9900, maxUsers: 50 },
  { value: "pro", label: "Pro", description: "Growing companies", price: 29900, maxUsers: 200 },
  { value: "enterprise", label: "Enterprise", description: "Large organizations", price: 99900, maxUsers: 1000 },
];

export default function NewOrganizationPage() {
  const router = useRouter();
  const { user, initialized, bootstrap, logout } = useAuthStore();

  const [form, setForm] = useState<CreateOrganizationInput>({
    name: "",
    email: "",
    plan: "basic",
    phone: "",
    website: "",
    max_users: 50,
    max_services: 2,
    monthly_price_cents: 9900,
  });
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    bootstrap();
  }, [bootstrap]);

  useEffect(() => {
    if (initialized && !user) {
      router.push("/login");
    }
  }, [initialized, user, router]);

  // Update defaults when plan changes
  useEffect(() => {
    const plan = PLANS.find((p) => p.value === form.plan);
    if (plan) {
      setForm((f) => ({
        ...f,
        max_users: plan.maxUsers,
        monthly_price_cents: plan.price,
      }));
    }
  }, [form.plan]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setCreating(true);

    try {
      const org = await organizationsApi.create(form);
      router.push(`/dashboard/superadmin/organizations/${org.id}`);
    } catch (err: any) {
      setError(
        err?.response?.data?.message ||
          err?.response?.data?.error ||
          "Failed to create organization"
      );
    } finally {
      setCreating(false);
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
        <div className="max-w-4xl mx-auto px-4 py-3 flex items-center justify-between">
          <Link href="/dashboard" className="flex items-center gap-2">
            <Shield className="w-6 h-6 text-blue-500" />
            <span className="font-bold text-lg">DefendCore VPN</span>
          </Link>
          <button
            onClick={handleLogout}
            className="flex items-center gap-2 px-3 py-1.5 text-sm bg-zinc-800 hover:bg-zinc-700 rounded transition"
          >
            <LogOut className="w-4 h-4" />
            Logout
          </button>
        </div>
      </nav>

      <main className="max-w-4xl mx-auto px-4 py-8">
        <div className="flex items-center gap-3 mb-6">
          <Link
            href="/dashboard/superadmin/organizations"
            className="p-2 hover:bg-zinc-800 rounded transition"
          >
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold">New Customer</h1>
            <p className="text-sm text-zinc-400">
              Create a new organization on the platform
            </p>
          </div>
        </div>

        {error && (
          <div className="mb-4 p-3 bg-red-950 border border-red-800 rounded text-red-300 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          {/* Plan selection */}
          <div className="mb-6">
            <h2 className="text-lg font-semibold mb-3">Step 1: Choose Plan</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {PLANS.map((plan) => (
                <button
                  key={plan.value}
                  type="button"
                  onClick={() => setForm((f) => ({ ...f, plan: plan.value }))}
                  className={`text-left p-4 rounded-lg border-2 transition ${
                    form.plan === plan.value
                      ? "border-purple-500 bg-purple-950/20"
                      : "border-zinc-800 bg-zinc-900 hover:border-zinc-700"
                  }`}
                >
                  <div className="flex items-start justify-between mb-2">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-semibold">{plan.label}</span>
                        {form.plan === plan.value && (
                          <CheckCircle className="w-4 h-4 text-purple-400" />
                        )}
                      </div>
                      <p className="text-xs text-zinc-400 mt-0.5">
                        {plan.description}
                      </p>
                    </div>
                    <span
                      className={`text-xs px-2 py-0.5 rounded border ${planColor(
                        plan.value
                      )}`}
                    >
                      ${(plan.price / 100).toFixed(0)}/mo
                    </span>
                  </div>
                  <div className="text-xs text-zinc-500">
                    Up to {plan.maxUsers} users
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Organization details */}
          <div className="mb-6">
            <h2 className="text-lg font-semibold mb-3">Step 2: Organization Details</h2>
            <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">
                  Organization Name *
                </label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  required
                  placeholder="ACME Corp"
                  className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-purple-500"
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">
                    Contact Email *
                  </label>
                  <input
                    type="email"
                    value={form.email}
                    onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
                    required
                    placeholder="admin@acme.com"
                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-purple-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Phone</label>
                  <input
                    type="tel"
                    value={form.phone || ""}
                    onChange={(e) => setForm((f) => ({ ...f, phone: e.target.value }))}
                    placeholder="+1 555 0100"
                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-purple-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Website</label>
                <input
                  type="url"
                  value={form.website || ""}
                  onChange={(e) => setForm((f) => ({ ...f, website: e.target.value }))}
                  placeholder="https://acme.com"
                  className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-purple-500"
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">
                    Max Users
                  </label>
                  <input
                    type="number"
                    value={form.max_users}
                    onChange={(e) =>
                      setForm((f) => ({
                        ...f,
                        max_users: parseInt(e.target.value) || 0,
                      }))
                    }
                    min={1}
                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-purple-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">
                    Monthly Price (cents)
                  </label>
                  <input
                    type="number"
                    value={form.monthly_price_cents}
                    onChange={(e) =>
                      setForm((f) => ({
                        ...f,
                        monthly_price_cents: parseInt(e.target.value) || 0,
                      }))
                    }
                    min={0}
                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-purple-500"
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-3">
            <Link
              href="/dashboard/superadmin/organizations"
              className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded font-medium transition"
            >
              Cancel
            </Link>
            <button
              type="submit"
              disabled={creating}
              className="px-6 py-2 bg-purple-600 hover:bg-purple-700 disabled:bg-zinc-700 rounded font-medium transition flex items-center gap-2"
            >
              {creating ? (
                "Creating..."
              ) : (
                <>
                  <Building2 className="w-4 h-4" />
                  Create Customer
                </>
              )}
            </button>
          </div>
        </form>
      </main>
    </div>
  );
}
