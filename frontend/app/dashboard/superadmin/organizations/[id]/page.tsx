"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import {
  Shield,
  LogOut,
  ArrowLeft,
  Building2,
  Users,
  CreditCard,
  DollarSign,
  Mail,
  Phone,
  Globe,
  CheckCircle,
  Plus,
  Trash2,
  X,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import { SubscriptionModal } from "@/components/organizations/SubscriptionModal";
import { InvoiceModal } from "@/components/organizations/InvoiceModal";
import {
  organizationsApi,
  Organization,
  OrganizationUser,
  Subscription,
  Invoice,
  statusColor,
  planColor,
  invoiceStatusColor,
  formatPrice,
  formatDate,
} from "@/lib/organizations-api";

type Tab = "users" | "subscriptions" | "invoices";

export default function OrganizationDetailsPage() {
  const router = useRouter();
  const params = useParams();
  const orgId = params?.id as string;

  const { user, initialized, bootstrap, logout } = useAuthStore();

  const [org, setOrg] = useState<Organization | null>(null);
  const [users, setUsers] = useState<OrganizationUser[]>([]);
  const [subs, setSubs] = useState<Subscription[]>([]);
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<Tab>("users");
  const [showSubscriptionModal, setShowSubscriptionModal] = useState(false);
  const [showInvoiceModal, setShowInvoiceModal] = useState(false);

  useEffect(() => {
    bootstrap();
  }, [bootstrap]);

  useEffect(() => {
    if (initialized && !user) {
      router.push("/login");
    }
  }, [initialized, user, router]);

  useEffect(() => {
    if (user && orgId) {
      loadData();
    }
  }, [user, orgId]);

  async function loadData() {
    setLoading(true);
    try {
      const [o, u, s, i] = await Promise.all([
        organizationsApi.get(orgId),
        organizationsApi.listUsers(orgId),
        organizationsApi.listSubscriptions(orgId),
        organizationsApi.listInvoices(orgId),
      ]);
      setOrg(o);
      setUsers(u);
      setSubs(s);
      setInvoices(i);
    } catch (err) {
      console.error("Failed to load org:", err);
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

  if (!org) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-zinc-950 text-white">
        <div className="text-center">
          <h2 className="text-xl font-bold mb-2">Organization not found</h2>
          <Link
            href="/dashboard/superadmin/organizations"
            className="text-blue-400 hover:text-blue-300"
          >
            ← Back to organizations
          </Link>
        </div>
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
        <div className="flex items-center gap-3 mb-6">
          <Link
            href="/dashboard/superadmin/organizations"
            className="p-2 hover:bg-zinc-800 rounded transition"
          >
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div className="flex-1">
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold">{org.name}</h1>
              <span
                className={`text-xs px-2 py-0.5 rounded border ${statusColor(
                  org.status
                )}`}
              >
                {org.status}
              </span>
              <span
                className={`text-xs px-2 py-0.5 rounded border ${planColor(
                  org.plan
                )}`}
              >
                {org.plan}
              </span>
            </div>
            <p className="text-sm text-zinc-400 font-mono mt-1">{org.slug}</p>
          </div>
        </div>

        {/* Info cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          <InfoCard
            icon={<Users className="w-5 h-5" />}
            label="Users"
            value={`${(users ?? []).length} / ${org.max_users}`}
            color="blue"
          />
          <InfoCard
            icon={<CreditCard className="w-5 h-5" />}
            label="Services"
            value={`${(subs ?? []).length} / ${org.max_services}`}
            color="purple"
          />
          <InfoCard
            icon={<DollarSign className="w-5 h-5" />}
            label="MRR"
            value={formatPrice(org.monthly_price_cents)}
            color="green"
          />
          <InfoCard
            icon={<Building2 className="w-5 h-5" />}
            label="Created"
            value={formatDate(org.created_at)}
            color="orange"
          />
        </div>

        {/* Contact info */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 mb-8">
          <h2 className="font-semibold mb-3">Contact</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
            <div className="flex items-center gap-2 text-zinc-400">
              <Mail className="w-4 h-4 text-zinc-600" />
              <span>{org.email}</span>
            </div>
            {org.phone && (
              <div className="flex items-center gap-2 text-zinc-400">
                <Phone className="w-4 h-4 text-zinc-600" />
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

        {/* Tabs */}
        <div className="border-b border-zinc-800 mb-6">
          <div className="flex gap-1">
            <TabButton
              active={activeTab === "users"}
              onClick={() => setActiveTab("users")}
              icon={<Users className="w-4 h-4" />}
              label="Users"
              count={(users ?? []).length}
            />
            <TabButton
              active={activeTab === "subscriptions"}
              onClick={() => setActiveTab("subscriptions")}
              icon={<CreditCard className="w-4 h-4" />}
              label="Services"
              count={(subs ?? []).length}
            />
            <TabButton
              active={activeTab === "invoices"}
              onClick={() => setActiveTab("invoices")}
              icon={<DollarSign className="w-4 h-4" />}
              label="Invoices"
              count={(invoices ?? []).length}
            />
          </div>
        </div>

        {/* Tab content */}
        {activeTab === "users" && <UsersTab orgId={orgId} users={users} onChange={loadData} />}
        {activeTab === "subscriptions" && (
          <SubscriptionsTab orgId={orgId} subs={subs} onChange={loadData} onAdd={() => setShowSubscriptionModal(true)} />
        )}
        {activeTab === "invoices" && <InvoicesTab invoices={invoices} onChange={loadData} onAdd={() => setShowInvoiceModal(true)} />}
      </main>

      {/* Modals */}
      <SubscriptionModal
        orgId={orgId}
        open={showSubscriptionModal}
        onClose={() => setShowSubscriptionModal(false)}
        onSuccess={loadData}
      />

      <InvoiceModal
        orgId={orgId}
        suggestedAmount={org?.monthly_price_cents ?? 0}
        open={showInvoiceModal}
        onClose={() => setShowInvoiceModal(false)}
        onSuccess={loadData}
      />
    </div>
  );
}

// =====================================================
// Sub-components
// =====================================================

function InfoCard({
  icon,
  label,
  value,
  color,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
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
        <div
          className={`w-9 h-9 rounded-lg border flex items-center justify-center ${colorMap[color]}`}
        >
          {icon}
        </div>
        <span className="text-sm text-zinc-500">{label}</span>
      </div>
      <div className="text-xl font-bold">{value}</div>
    </div>
  );
}

function TabButton({
  active,
  onClick,
  icon,
  label,
  count,
}: {
  active: boolean;
  onClick: () => void;
  icon: React.ReactNode;
  label: string;
  count: number;
}) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-2 px-4 py-3 border-b-2 transition ${
        active
          ? "border-purple-500 text-white"
          : "border-transparent text-zinc-500 hover:text-zinc-300"
      }`}
    >
      {icon}
      <span className="font-medium">{label}</span>
      <span className="text-xs px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400">
        {count}
      </span>
    </button>
  );
}

// =====================================================
// Users Tab
// =====================================================

function UsersTab({
  orgId,
  users,
  onChange,
}: {
  orgId: string;
  users: OrganizationUser[];
  onChange: () => void;
}) {
  const [showAdd, setShowAdd] = useState(false);
  const [newUserId, setNewUserId] = useState("");
  const [newRole, setNewRole] = useState<"admin" | "user">("user");
  const [error, setError] = useState("");

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await organizationsApi.addUser(orgId, newUserId, newRole);
      setNewUserId("");
      setShowAdd(false);
      onChange();
    } catch (err: any) {
      setError(
        err?.response?.data?.message || err?.response?.data?.error || "Failed to add user"
      );
    }
  }

  async function handleRemove(userId: string) {
    if (!confirm("Remove this user from the organization?")) return;
    try {
      await organizationsApi.removeUser(orgId, userId);
      onChange();
    } catch (err) {
      alert("Failed to remove user");
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="font-semibold">Organization Members</h2>
        <button
          onClick={() => setShowAdd(true)}
          className="flex items-center gap-2 px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded font-medium transition"
        >
          <Plus className="w-4 h-4" />
          Add User
        </button>
      </div>

      {showAdd && (
        <form
          onSubmit={handleAdd}
          className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 mb-4 space-y-3"
        >
          {error && (
            <div className="p-2 bg-red-950 border border-red-800 rounded text-red-300 text-sm">
              {error}
            </div>
          )}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div className="md:col-span-2">
              <label className="block text-sm font-medium mb-1">User ID (UUID)</label>
              <input
                type="text"
                value={newUserId}
                onChange={(e) => setNewUserId(e.target.value)}
                required
                placeholder="54463435-f10f-45f0-b06f-..."
                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-purple-500 font-mono text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Role</label>
              <select
                value={newRole}
                onChange={(e) => setNewRole(e.target.value as "admin" | "user")}
                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-purple-500"
              >
                <option value="user">User</option>
                <option value="admin">Admin</option>
              </select>
            </div>
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setShowAdd(false)}
              className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded font-medium transition"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded font-medium transition"
            >
              Add User
            </button>
          </div>
        </form>
      )}

      {(users ?? []).length === 0 ? (
        <EmptyState
          icon={<Users className="w-12 h-12" />}
          title="No members yet"
          description="Add users to give them access to this organization"
        />
      ) : (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg divide-y divide-zinc-800">
          {(users ?? []).map((u) => (
            <div key={u.id} className="flex items-center justify-between p-4">
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-lg bg-zinc-800 flex items-center justify-center">
                  <Users className="w-4 h-4 text-zinc-400" />
                </div>
                <div>
                  <div className="font-mono text-sm">{u.user_id}</div>
                  <div className="text-xs text-zinc-500">
                    Role: <span className="text-zinc-400">{u.role}</span> · Joined{" "}
                    {formatDate(u.joined_at)}
                  </div>
                </div>
              </div>
              <button
                onClick={() => handleRemove(u.user_id)}
                className="p-2 hover:bg-red-950 rounded text-zinc-600 hover:text-red-400 transition"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// =====================================================
// Subscriptions Tab
// =====================================================

function SubscriptionsTab({
  orgId,
  subs,
  onChange,
  onAdd,
}: {
  orgId: string;
  subs: Subscription[];
  onChange: () => void;
  onAdd: () => void;
}) {
  async function handleDelete(id: string) {
    if (!confirm("Delete this subscription?")) return;
    try {
      await organizationsApi.deleteSubscription(id);
      onChange();
    } catch (err) {
      alert("Failed to delete subscription");
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="font-semibold">VPN Service Subscriptions</h2>
        <button
          onClick={onAdd}
          className="flex items-center gap-2 px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded font-medium transition"
        >
          <Plus className="w-4 h-4" />
          Add Service
        </button>
      </div>

      {(subs ?? []).length === 0 ? (
        <EmptyState
          icon={<CreditCard className="w-12 h-12" />}
          title="No subscriptions"
          description="This organization has not subscribed to any VPN service"
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {(subs ?? []).map((sub) => (
            <div
              key={sub.id}
              className="bg-zinc-900 border border-zinc-800 rounded-lg p-5"
            >
              <div className="flex items-start justify-between mb-3">
                <div>
                  <div className="font-semibold">{sub.service_type}</div>
                  <div className="text-xs text-zinc-500 mt-0.5">
                    Started {formatDate(sub.started_at)}
                  </div>
                </div>
                <span
                  className={`text-xs px-2 py-0.5 rounded border ${statusColor(
                    sub.status
                  )}`}
                >
                  {sub.status}
                </span>
              </div>
              <div className="grid grid-cols-2 gap-3 text-sm mb-3">
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
              <button
                onClick={() => handleDelete(sub.id)}
                className="text-xs text-red-400 hover:text-red-300 flex items-center gap-1"
              >
                <Trash2 className="w-3 h-3" />
                Remove
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// =====================================================
// Invoices Tab
// =====================================================

function InvoicesTab({
  invoices,
  onChange,
  onAdd,
}: {
  invoices: Invoice[];
  onChange: () => void;
  onAdd: () => void;
}) {
  async function handleMarkPaid(id: string) {
    try {
      await organizationsApi.markInvoicePaid(id);
      onChange();
    } catch (err) {
      alert("Failed to mark invoice as paid");
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="font-semibold">Invoices</h2>
        <button
          onClick={onAdd}
          className="flex items-center gap-2 px-4 py-2 bg-green-600 hover:bg-green-700 rounded font-medium transition"
        >
          <Plus className="w-4 h-4" />
          Create Invoice
        </button>
      </div>

      {(invoices ?? []).length === 0 ? (
        <EmptyState
          icon={<DollarSign className="w-12 h-12" />}
          title="No invoices"
          description="No billing history for this organization"
        />
      ) : (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
          <table className="w-full">
            <thead className="bg-zinc-950/50 border-b border-zinc-800">
              <tr className="text-left text-xs uppercase tracking-wider text-zinc-500">
                <th className="px-4 py-3">Invoice #</th>
                <th className="px-4 py-3">Amount</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Period</th>
                <th className="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-800">
              {(invoices ?? []).map((inv) => (
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
                  <td className="px-4 py-3 text-right">
                    {inv.status !== "paid" && (
                      <button
                        onClick={() => handleMarkPaid(inv.id)}
                        className="text-xs px-3 py-1 bg-green-950 hover:bg-green-900 text-green-400 rounded transition flex items-center gap-1 ml-auto"
                      >
                        <CheckCircle className="w-3 h-3" />
                        Mark Paid
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

function EmptyState({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
}) {
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-12 text-center">
      <div className="text-zinc-600 flex justify-center mb-4">{icon}</div>
      <h3 className="font-semibold mb-2">{title}</h3>
      <p className="text-sm text-zinc-400">{description}</p>
    </div>
  );
}
