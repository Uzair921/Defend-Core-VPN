"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Shield,
  LogOut,
  Building2,
  Users,
  CreditCard,
  DollarSign,
  Activity,
  ChevronRight,
  Mail,
  Globe,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import {
  orgApi,
  Organization,
  OrganizationUser,
  Subscription,
  Invoice,
  statusColor,
  planColor,
  formatPrice,
} from "@/lib/organizations-api";

export default function OrgDashboardPage() {
  const router = useRouter();
  const { user, initialized, bootstrap, logout } = useAuthStore();

  const [org, setOrg] = useState<Organization | null>(null);
  const [users, setUsers] = useState<OrganizationUser[]>([]);
  const [subs, setSubs] = useState<Subscription[]>([]);
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

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
    setError("");
    try {
      const [o, u, s, i] = await Promise.all([
        orgApi.getMyOrganization(),
        orgApi.listMyUsers(),
        orgApi.listMySubscriptions(),
        orgApi.listMyInvoices(),
      ]);
      setOrg(o);
      setUsers(u);
      setSubs(s);
      setInvoices(i);
    } catch (err: any) {
      if (err?.response?.status === 403) {
        setError("You are not a member of any organization. Contact your administrator.");
      } else {
        setError("Failed to load organization data");
      }
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

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-zinc-950 text-white">
        <div className="text-zinc-400">Loading organization...</div>
      </div>
    );
  }

  if (error || !org) {
    return (
      <div className="min-h-screen bg-zinc-950 text-white">
        <nav className="border-b border-zinc-800 bg-zinc-900">
          <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
            <Link href="/dashboard" className="flex items-center gap-2">
              <Shield className="w-6 h-6 text-blue-500" />
              <span className="font-bold text-lg">DefendCore VPN</span>
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
        <main className="max-w-4xl mx-auto px-4 py-16 text-center">
          <Building2 className="w-16 h-16 text-zinc-700 mx-auto mb-4" />
          <h1 className="text-2xl font-bold mb-2">No Organization</h1>
          <p className="text-zinc-400 mb-6">{error || "You are not a member of any organization"}</p>
          <Link
            href="/dashboard"
            className="inline-block px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium"
          >
            Back to Dashboard
          </Link>
        </main>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-zinc-950 text-white">
      <nav className="border-b border-zinc-800 bg-zinc-900">
        <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
          <Link href="/dashboard" className="flex items-center gap-2">
            <Shield className="w-6 h-6 text-blue-500" />
            <span className="font-bold text-lg">DefendCore VPN</span>
            <span className="text-xs px-2 py-0.5 rounded bg-blue-950 text-blue-400 border border-blue-800 ml-2">
              {org.name}
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
        <div className="flex items-start justify-between mb-6">
          <div>
            <div className="flex items-center gap-3 mb-1">
              <h1 className="text-2xl font-bold">{org.name}</h1>
              <span className={`text-xs px-2 py-0.5 rounded border ${statusColor(org.status)}`}>
                {org.status}
              </span>
              <span className={`text-xs px-2 py-0.5 rounded border ${planColor(org.plan)}`}>
                {org.plan}
              </span>
            </div>
            <p className="text-sm text-zinc-500 font-mono">{org.slug}</p>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          <StatCard
            icon={<Users className="w-5 h-5" />}
            label="Team Members"
            value={`${users.length} / ${org.max_users}`}
            color="blue"
          />
          <StatCard
            icon={<CreditCard className="w-5 h-5" />}
            label="VPN Services"
            value={`${subs.length} / ${org.max_services}`}
            color="purple"
          />
          <StatCard
            icon={<DollarSign className="w-5 h-5" />}
            label="Monthly Cost"
            value={formatPrice(org.monthly_price_cents)}
            color="green"
          />
          <StatCard
            icon={<Activity className="w-5 h-5" />}
            label="Invoices"
            value={invoices.length}
            color="orange"
          />
        </div>

        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 mb-8">
          <h2 className="font-semibold mb-3">Contact Information</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
            <div className="flex items-center gap-2 text-zinc-400">
              <Mail className="w-4 h-4 text-zinc-600" />
              <span>{org.email}</span>
            </div>
            {org.phone && (
              <div className="flex items-center gap-2 text-zinc-400">
                <span className="text-zinc-600">📞</span>
                <span>{org.phone}</span>
              </div>
            )}
            {org.website && (
              <div className="flex items-center gap-2 text-zinc-400">
                <Globe className="w-4 h-4 text-zinc-600" />
                <a
                  href={org.website}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-blue-400"
                >
                  {org.website}
                </a>
              </div>
            )}
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <QuickLink
            href="/dashboard/org/users"
            icon={<Users className="w-6 h-6 text-blue-400" />}
            title="Team Members"
            description={`${users.length} users`}
          />
          <QuickLink
            href="/dashboard/org/subscriptions"
            icon={<CreditCard className="w-6 h-6 text-purple-400" />}
            title="VPN Services"
            description={`${subs.length} services`}
          />
          <QuickLink
            href="/dashboard/org/invoices"
            icon={<DollarSign className="w-6 h-6 text-green-400" />}
            title="Invoices"
            description={`${invoices.length} invoices`}
          />
        </div>
      </main>
    </div>
  );
}

function StatCard({
  icon,
  label,
  value,
  color,
}: {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  color: "blue" | "purple" | "green" | "orange";
}) {
  const colorMap = {
    blue: "text-blue-400 bg-blue-950/50 border-blue-900",
    purple: "text-purple-400 bg-purple-950/50 border-purple-900",
    green: "text-green-400 bg-green-950/50 border-green-900",
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
      <div className="text-xl font-bold">{value}</div>
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
