"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Shield,
  LogOut,
  ArrowLeft,
  Building2,
  Users,
  CreditCard,
  DollarSign,
  Activity,
  TrendingUp,
  ChevronRight,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import { organizationsApi, Organization, formatPrice } from "@/lib/organizations-api";

export default function SuperAdminPage() {
  const router = useRouter();
  const { user, initialized, bootstrap, logout } = useAuthStore();

  const [orgs, setOrgs] = useState<Organization[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    bootstrap();
  }, [bootstrap]);

  useEffect(() => {
    if (initialized && !user) {
      router.push("/login");
    }
  }, [initialized, user, router]);

  useEffect(() => {
    if (user) {
      loadData();
    }
  }, [user]);

  async function loadData() {
    setLoading(true);
    try {
      const page = await organizationsApi.list({ limit: 5 });
      setOrgs(page.data);
      setTotal(page.total);
    } catch (err) {
      console.error("Failed to load orgs:", err);
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

  const activeOrgs = orgs.filter((o) => o.status === "active").length;
  const mrr = orgs.reduce((sum, o) => sum + (o.monthly_price_cents || 0), 0);

  return (
    <div className="min-h-screen bg-zinc-950 text-white">
      {/* Nav */}
      <nav className="border-b border-zinc-800 bg-zinc-900">
        <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
          <Link href="/dashboard" className="flex items-center gap-2">
            <Shield className="w-6 h-6 text-blue-500" />
            <span className="font-bold text-lg">DefendCore VPN</span>
            <span className="text-xs px-2 py-0.5 rounded bg-purple-950 text-purple-400 border border-purple-800 ml-2">
              SuperAdmin
            </span>
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

      <main className="max-w-7xl mx-auto px-4 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <Link
              href="/dashboard"
              className="p-2 hover:bg-zinc-800 rounded transition"
            >
              <ArrowLeft className="w-5 h-5" />
            </Link>
            <div>
              <h1 className="text-2xl font-bold">Platform Overview</h1>
              <p className="text-sm text-zinc-400">
                Manage customers, subscriptions, and billing
              </p>
            </div>
          </div>
          <Link
            href="/dashboard/superadmin/organizations/new"
            className="flex items-center gap-2 px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded font-medium transition"
          >
            <Building2 className="w-4 h-4" />
            New Customer
          </Link>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          <StatCard
            icon={<Building2 className="w-5 h-5" />}
            label="Total Customers"
            value={total}
            color="blue"
          />
          <StatCard
            icon={<Activity className="w-5 h-5" />}
            label="Active"
            value={activeOrgs}
            color="green"
          />
          <StatCard
            icon={<DollarSign className="w-5 h-5" />}
            label="MRR"
            value={formatPrice(mrr)}
            color="purple"
          />
          <StatCard
            icon={<TrendingUp className="w-5 h-5" />}
            label="Growth"
            value="—"
            color="orange"
          />
        </div>

        {/* Quick Links */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          <QuickLink
            href="/dashboard/superadmin/organizations"
            icon={<Building2 className="w-6 h-6 text-blue-400" />}
            title="Organizations"
            description="Manage all customers"
          />
          <QuickLink
            href="/dashboard/superadmin/subscriptions"
            icon={<CreditCard className="w-6 h-6 text-purple-400" />}
            title="Subscriptions"
            description="VPN service allotments"
          />
          <QuickLink
            href="/dashboard/superadmin/billing"
            icon={<DollarSign className="w-6 h-6 text-green-400" />}
            title="Billing"
            description="Invoices & payments"
          />
        </div>

        {/* Recent Organizations */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg">
          <div className="flex items-center justify-between px-5 py-4 border-b border-zinc-800">
            <h2 className="font-semibold">Recent Customers</h2>
            <Link
              href="/dashboard/superadmin/organizations"
              className="text-sm text-blue-400 hover:text-blue-300 flex items-center gap-1"
            >
              View all
              <ChevronRight className="w-4 h-4" />
            </Link>
          </div>

          {loading ? (
            <div className="p-8 text-center text-zinc-400">Loading...</div>
          ) : orgs.length === 0 ? (
            <div className="p-12 text-center">
              <Building2 className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
              <h3 className="font-semibold mb-2">No customers yet</h3>
              <p className="text-sm text-zinc-400 mb-4">
                Create your first customer to get started
              </p>
              <Link
                href="/dashboard/superadmin/organizations/new"
                className="inline-block px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded font-medium transition"
              >
                Create First Customer
              </Link>
            </div>
          ) : (
            <div className="divide-y divide-zinc-800">
              {orgs.map((org) => (
                <Link
                  key={org.id}
                  href={`/dashboard/superadmin/organizations/${org.id}`}
                  className="flex items-center justify-between p-4 hover:bg-zinc-800/50 transition"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <div className="w-10 h-10 rounded-lg bg-zinc-800 flex items-center justify-center flex-shrink-0">
                      <Building2 className="w-5 h-5 text-zinc-400" />
                    </div>
                    <div className="min-w-0">
                      <div className="font-medium truncate">{org.name}</div>
                      <div className="text-xs text-zinc-500 truncate">
                        {org.email} · {org.plan}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-xs text-zinc-500">
                      {formatPrice(org.monthly_price_cents)}/mo
                    </span>
                    <ChevronRight className="w-4 h-4 text-zinc-600" />
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

// =====================================================
// Sub-components
// =====================================================

function StatCard({
  icon,
  label,
  value,
  color,
}: {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  color: "blue" | "green" | "purple" | "orange";
}) {
  const colorMap = {
    blue: "text-blue-400 bg-blue-950/50 border-blue-900",
    green: "text-green-400 bg-green-950/50 border-green-900",
    purple: "text-purple-400 bg-purple-950/50 border-purple-900",
    orange: "text-orange-400 bg-orange-950/50 border-orange-900",
  };

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5">
      <div className="flex items-center gap-3 mb-3">
        <div className={`w-9 h-9 rounded-lg border flex items-center justify-center ${colorMap[color]}`}>
          {icon}
        </div>
        <span className="text-sm text-zinc-500">{label}</span>
      </div>
      <div className="text-2xl font-bold">{value}</div>
    </div>
  );
}

function QuickLink({
  href,
  icon,
  title,
  description,
}: {
  href: string;
  icon: React.ReactNode;
  title: string;
  description: string;
}) {
  return (
    <Link
      href={href}
      className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 hover:border-zinc-700 transition group"
    >
      <div className="flex items-center justify-between mb-3">
        <div className="w-11 h-11 rounded-lg bg-zinc-800 flex items-center justify-center">
          {icon}
        </div>
        <ChevronRight className="w-5 h-5 text-zinc-600 group-hover:text-zinc-400 transition" />
      </div>
      <h3 className="font-semibold mb-1">{title}</h3>
      <p className="text-sm text-zinc-500">{description}</p>
    </Link>
  );
}
