"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Shield,
  LogOut,
  ArrowLeft,
  Plus,
  Building2,
  Search,
  ChevronRight,
  XCircle,
  Trash2,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import {
  organizationsApi,
  Organization,
  OrgStatus,
  OrgPlan,
  statusColor,
  planColor,
  formatPrice,
  formatDate,
} from "@/lib/organizations-api";

export default function OrganizationsListPage() {
  const router = useRouter();
  const { user, initialized, bootstrap, logout } = useAuthStore();

  const [orgs, setOrgs] = useState<Organization[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("");
  const [planFilter, setPlanFilter] = useState<string>("");

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
      loadOrgs();
    }
  }, [user, statusFilter, planFilter]);

  async function loadOrgs() {
    setLoading(true);
    try {
      const page = await organizationsApi.list({
        search: search || undefined,
        status: statusFilter || undefined,
        plan: planFilter || undefined,
        limit: 50,
      });
      setOrgs(page.data);
      setTotal(page.total);
    } catch (err) {
      console.error("Failed to load orgs:", err);
    } finally {
      setLoading(false);
    }
  }

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    loadOrgs();
  }

  async function handleDelete(id: string, name: string) {
    if (!confirm(`Delete organization "${name}"? This cannot be undone.`)) return;
    try {
      await organizationsApi.delete(id);
      setOrgs((prev) => prev.filter((o) => o.id !== id));
      setTotal((t) => t - 1);
    } catch (err) {
      alert("Failed to delete organization");
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
              href="/dashboard/superadmin"
              className="p-2 hover:bg-zinc-800 rounded transition"
            >
              <ArrowLeft className="w-5 h-5" />
            </Link>
            <div>
              <h1 className="text-2xl font-bold">Organizations</h1>
              <p className="text-sm text-zinc-400">
                {total} customer{total !== 1 ? "s" : ""} on the platform
              </p>
            </div>
          </div>
          <Link
            href="/dashboard/superadmin/organizations/new"
            className="flex items-center gap-2 px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded font-medium transition"
          >
            <Plus className="w-4 h-4" />
            New Customer
          </Link>
        </div>

        {/* Filters */}
        <form
          onSubmit={handleSearch}
          className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 mb-6 flex flex-wrap gap-3 items-center"
        >
          <div className="flex-1 min-w-[200px] relative">
            <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search by name or email..."
              className="w-full pl-9 pr-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500"
            />
          </div>

          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500"
          >
            <option value="">All status</option>
            <option value="active">Active</option>
            <option value="trial">Trial</option>
            <option value="suspended">Suspended</option>
            <option value="cancelled">Cancelled</option>
          </select>

          <select
            value={planFilter}
            onChange={(e) => setPlanFilter(e.target.value)}
            className="px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500"
          >
            <option value="">All plans</option>
            <option value="trial">Trial</option>
            <option value="basic">Basic</option>
            <option value="pro">Pro</option>
            <option value="enterprise">Enterprise</option>
          </select>

          <button
            type="submit"
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition"
          >
            Search
          </button>
        </form>

        {/* List */}
        {loading ? (
          <div className="text-center text-zinc-400 py-12">Loading...</div>
        ) : orgs.length === 0 ? (
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-12 text-center">
            <Building2 className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
            <h3 className="font-semibold mb-2">No customers found</h3>
            <p className="text-sm text-zinc-400 mb-4">
              {search || statusFilter || planFilter
                ? "Try a different search or filter"
                : "Create your first customer to get started"}
            </p>
            {!search && !statusFilter && !planFilter && (
              <Link
                href="/dashboard/superadmin/organizations/new"
                className="inline-block px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded font-medium transition"
              >
                Create First Customer
              </Link>
            )}
          </div>
        ) : (
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
            <table className="w-full">
              <thead className="bg-zinc-950/50 border-b border-zinc-800">
                <tr className="text-left text-xs uppercase tracking-wider text-zinc-500">
                  <th className="px-4 py-3">Organization</th>
                  <th className="px-4 py-3">Plan</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Users</th>
                  <th className="px-4 py-3">MRR</th>
                  <th className="px-4 py-3">Created</th>
                  <th className="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800">
                {orgs.map((org) => (
                  <tr
                    key={org.id}
                    className="hover:bg-zinc-800/50 transition cursor-pointer"
                    onClick={() =>
                      router.push(`/dashboard/superadmin/organizations/${org.id}`)
                    }
                  >
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-3">
                        <div className="w-9 h-9 rounded-lg bg-zinc-800 flex items-center justify-center flex-shrink-0">
                          <Building2 className="w-4 h-4 text-zinc-400" />
                        </div>
                        <div className="min-w-0">
                          <div className="font-medium truncate">{org.name}</div>
                          <div className="text-xs text-zinc-500 truncate">
                            {org.email}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`text-xs px-2 py-0.5 rounded border ${planColor(
                          org.plan
                        )}`}
                      >
                        {org.plan}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`text-xs px-2 py-0.5 rounded border ${statusColor(
                          org.status
                        )}`}
                      >
                        {org.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-400">
                      — / {org.max_users}
                    </td>
                    <td className="px-4 py-3 text-sm">
                      {formatPrice(org.monthly_price_cents)}
                    </td>
                    <td className="px-4 py-3 text-sm text-zinc-500">
                      {formatDate(org.created_at)}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDelete(org.id, org.name);
                        }}
                        className="p-2 hover:bg-red-950 rounded text-zinc-600 hover:text-red-400 transition"
                        title="Delete"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
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
