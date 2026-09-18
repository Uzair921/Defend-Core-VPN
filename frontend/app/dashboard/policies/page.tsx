"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Shield,
  LogOut,
  ArrowLeft,
  Plus,
  Trash2,
  CheckCircle,
  XCircle,
  Globe,
  Network,
  Clock,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import {
  policiesApi,
  AccessPolicy,
  PolicyType,
  PolicyAction,
} from "@/lib/policies-api";

export default function PoliciesPage() {
  const router = useRouter();
  const { user, initialized, bootstrap, logout } = useAuthStore();

  const [policies, setPolicies] = useState<AccessPolicy[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAdd, setShowAdd] = useState(false);
  const [error, setError] = useState("");
  const [creating, setCreating] = useState(false);

  // New policy form
  const [newType, setNewType] = useState<PolicyType>("cidr");
  const [newValue, setNewValue] = useState("");
  const [newAction, setNewAction] = useState<PolicyAction>("allow");
  const [newPriority, setNewPriority] = useState(100);
  const [newDescription, setNewDescription] = useState("");

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
      loadPolicies();
    }
  }, [user]);

  async function loadPolicies() {
    if (!user) return;
    setLoading(true);
    try {
      const list = await policiesApi.listByUser(user.id);
      setPolicies(list);
    } catch (err) {
      console.error("Failed to load policies:", err);
    } finally {
      setLoading(false);
    }
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!user) return;
    setError("");
    setCreating(true);
    try {
      const created = await policiesApi.create({
        user_id: user.id,
        type: newType,
        value: newValue.trim(),
        action: newAction,
        priority: newPriority,
        description: newDescription.trim() || undefined,
      });
      setPolicies((prev) =>
        [...prev, created].sort((a, b) => a.priority - b.priority)
      );
      setNewValue("");
      setNewDescription("");
      setNewPriority(100);
      setShowAdd(false);
    } catch (err: any) {
      setError(
        err?.response?.data?.message ||
          err?.response?.data?.error ||
          "Failed to create policy"
      );
    } finally {
      setCreating(false);
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Delete this policy?")) return;
    try {
      await policiesApi.delete(id);
      setPolicies((prev) => prev.filter((p) => p.id !== id));
    } catch (err) {
      alert("Failed to delete policy");
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

  const allowCount = policies.filter((p) => p.action === "allow").length;
  const denyCount = policies.filter((p) => p.action === "deny").length;

  return (
    <div className="min-h-screen bg-zinc-950 text-white">
      {/* Nav */}
      <nav className="border-b border-zinc-800 bg-zinc-900">
        <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
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

      <main className="max-w-5xl mx-auto px-4 py-8">
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
              <h1 className="text-2xl font-bold">Access Policies</h1>
              <p className="text-sm text-zinc-400">
                Control what {user.email} can access through the VPN
              </p>
            </div>
          </div>
          <button
            onClick={() => setShowAdd(true)}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition"
          >
            <Plus className="w-4 h-4" />
            Add Policy
          </button>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-3 gap-4 mb-6">
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
            <div className="text-xs text-zinc-500 mb-1">Total Rules</div>
            <div className="text-2xl font-bold">{policies.length}</div>
          </div>
          <div className="bg-zinc-900 border border-green-900/50 rounded-lg p-4">
            <div className="text-xs text-green-500 mb-1">Allowed</div>
            <div className="text-2xl font-bold text-green-400">{allowCount}</div>
          </div>
          <div className="bg-zinc-900 border border-red-900/50 rounded-lg p-4">
            <div className="text-xs text-red-500 mb-1">Denied</div>
            <div className="text-2xl font-bold text-red-400">{denyCount}</div>
          </div>
        </div>

        {/* Info box */}
        <div className="bg-blue-950/30 border border-blue-900 rounded-lg p-4 mb-6 text-sm text-blue-200">
          <strong>Zero-trust default:</strong> All traffic is blocked unless a
          matching <span className="font-mono">allow</span> rule exists.
          Higher priority (lower number) rules are evaluated first.
        </div>

        {/* Add Policy Modal */}
        {showAdd && (
          <div className="fixed inset-0 bg-black/70 flex items-center justify-center p-4 z-50">
            <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 w-full max-w-lg">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-bold">Add Access Policy</h2>
                <button
                  onClick={() => setShowAdd(false)}
                  className="p-1 hover:bg-zinc-800 rounded"
                >
                  <XCircle className="w-5 h-5" />
                </button>
              </div>

              <form onSubmit={handleCreate} className="space-y-4">
                {error && (
                  <div className="p-3 bg-red-950 border border-red-800 rounded text-red-300 text-sm">
                    {error}
                  </div>
                )}

                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-sm font-medium mb-1">
                      Type
                    </label>
                    <select
                      value={newType}
                      onChange={(e) => setNewType(e.target.value as PolicyType)}
                      className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500"
                    >
                      <option value="cidr">CIDR (10.0.0.0/24)</option>
                      <option value="ip">IP (10.0.0.5)</option>
                      <option value="domain">Domain (git.acme.com)</option>
                      <option value="url">URL (https://jira.acme.com)</option>
                      <option value="port">Port (443)</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium mb-1">
                      Action
                    </label>
                    <select
                      value={newAction}
                      onChange={(e) =>
                        setNewAction(e.target.value as PolicyAction)
                      }
                      className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500"
                    >
                      <option value="allow">Allow</option>
                      <option value="deny">Deny</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">
                    Value
                  </label>
                  <input
                    type="text"
                    value={newValue}
                    onChange={(e) => setNewValue(e.target.value)}
                    required
                    placeholder={
                      newType === "cidr"
                        ? "10.0.0.0/24"
                        : newType === "ip"
                        ? "10.0.0.5"
                        : newType === "domain"
                        ? "git.acme.com"
                        : newType === "url"
                        ? "https://jira.acme.com"
                        : "443"
                    }
                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500 font-mono text-sm"
                  />
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-sm font-medium mb-1">
                      Priority (lower = higher)
                    </label>
                    <input
                      type="number"
                      value={newPriority}
                      onChange={(e) => setNewPriority(parseInt(e.target.value) || 100)}
                      min={1}
                      max={9999}
                      className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium mb-1">
                      Description
                    </label>
                    <input
                      type="text"
                      value={newDescription}
                      onChange={(e) => setNewDescription(e.target.value)}
                      placeholder="Dev network"
                      className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500"
                    />
                  </div>
                </div>

                <button
                  type="submit"
                  disabled={creating}
                  className="w-full py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-zinc-700 rounded font-medium transition"
                >
                  {creating ? "Creating..." : "Create Policy"}
                </button>
              </form>
            </div>
          </div>
        )}

        {/* Policies List */}
        {loading ? (
          <div className="text-center text-zinc-400 py-12">
            Loading policies...
          </div>
        ) : policies.length === 0 ? (
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-12 text-center">
            <Shield className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
            <h3 className="font-semibold mb-2">No policies yet</h3>
            <p className="text-zinc-400 text-sm mb-4">
              By default, all traffic is <strong>denied</strong>. Add allow rules
              to grant access.
            </p>
            <button
              onClick={() => setShowAdd(true)}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition"
            >
              Add First Policy
            </button>
          </div>
        ) : (
          <div className="space-y-2">
            {policies.map((p) => (
              <div
                key={p.id}
                className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 hover:border-zinc-700 transition"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3 flex-1 min-w-0">
                    <div className="flex-shrink-0">
                      {p.action === "allow" ? (
                        <CheckCircle className="w-5 h-5 text-green-400" />
                      ) : (
                        <XCircle className="w-5 h-5 text-red-400" />
                      )}
                    </div>

                    <div className="flex-shrink-0">
                      {p.type === "cidr" || p.type === "ip" ? (
                        <Network className="w-4 h-4 text-zinc-500" />
                      ) : (
                        <Globe className="w-4 h-4 text-zinc-500" />
                      )}
                    </div>

                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <span className="text-xs uppercase tracking-wider text-zinc-500">
                          {p.type}
                        </span>
                        <span className="font-mono text-sm truncate">
                          {p.value}
                        </span>
                        <span
                          className={`text-xs px-2 py-0.5 rounded ${
                            p.action === "allow"
                              ? "bg-green-950 text-green-400"
                              : "bg-red-950 text-red-400"
                          }`}
                        >
                          {p.action}
                        </span>
                      </div>
                      {p.description && (
                        <div className="text-xs text-zinc-500 mt-0.5">
                          {p.description}
                        </div>
                      )}
                    </div>

                    <div className="flex-shrink-0 flex items-center gap-2 text-xs text-zinc-500">
                      <Clock className="w-3 h-3" />
                      <span>Priority {p.priority}</span>
                    </div>
                  </div>

                  <button
                    onClick={() => handleDelete(p.id)}
                    className="p-2 hover:bg-red-950 rounded text-zinc-500 hover:text-red-400 transition ml-2"
                    title="Delete policy"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
