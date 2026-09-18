"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import {
  Shield,
  LogOut,
  ArrowLeft,
  UserPlus,
  Trash2,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import {
  vpnServicesApi,
  VpnService,
  ServiceUser,
} from "@/lib/vpn-services-api";

export default function ServiceUsersPage() {
  const router = useRouter();
  const params = useParams();
  const serviceId = params?.id as string;

  const { user, initialized, bootstrap, logout } = useAuthStore();

  const [service, setService] = useState<VpnService | null>(null);
  const [users, setUsers] = useState<ServiceUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAdd, setShowAdd] = useState(false);
  const [newUserId, setNewUserId] = useState("");
  const [newRole, setNewRole] = useState("user");
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
    if (user && serviceId) {
      loadData();
    }
  }, [user, serviceId]);

  async function loadData() {
    setLoading(true);
    try {
      const [svc, userList] = await Promise.all([
        vpnServicesApi.get(serviceId),
        vpnServicesApi.listUsers(serviceId),
      ]);
      setService(svc);
      setUsers(userList);
    } catch (err) {
      console.error("Failed to load:", err);
    } finally {
      setLoading(false);
    }
  }

  async function handleAssign(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      const result = await vpnServicesApi.assignUser(
        serviceId,
        newUserId,
        newRole
      );
      setUsers((prev) => [...prev, result]);
      setNewUserId("");
      setShowAdd(false);
    } catch (err: any) {
      setError(
        err?.response?.data?.message ||
          err?.response?.data?.error ||
          "Failed to assign user"
      );
    }
  }

  async function handleUnassign(userId: string) {
    if (!confirm("Remove this user from the service?")) return;
    try {
      await vpnServicesApi.unassignUser(serviceId, userId);
      setUsers((prev) => prev.filter((u) => u.user_id !== userId));
    } catch (err) {
      alert("Failed to unassign user");
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
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <Link
              href="/dashboard/vpn/services"
              className="p-2 hover:bg-zinc-800 rounded transition"
            >
              <ArrowLeft className="w-5 h-5" />
            </Link>
            <div>
              <h1 className="text-2xl font-bold">
                {service?.name || "Service"} — Users
              </h1>
              <p className="text-sm text-zinc-400">
                Manage who can access this service
              </p>
            </div>
          </div>
          <button
            onClick={() => setShowAdd(true)}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition"
          >
            <UserPlus className="w-4 h-4" />
            Assign User
          </button>
        </div>

        {service && (
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 mb-6">
            <div className="flex items-center gap-3">
              <span className="text-2xl">{service.type_icon || "🔒"}</span>
              <div>
                <div className="font-semibold">{service.name}</div>
                <div className="text-xs text-zinc-500 font-mono">
                  {service.subnet} · {service.status}
                </div>
              </div>
            </div>
          </div>
        )}

        {error && (
          <div className="mb-4 p-3 bg-red-950 border border-red-800 rounded text-red-300 text-sm">
            {error}
          </div>
        )}

        {showAdd && (
          <form
            onSubmit={handleAssign}
            className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 mb-6 space-y-3"
          >
            <h3 className="font-semibold mb-2">Assign User</h3>
            <div>
              <label className="block text-sm font-medium mb-1">
                User ID (UUID) *
              </label>
              <input
                type="text"
                value={newUserId}
                onChange={(e) => setNewUserId(e.target.value)}
                required
                placeholder="54463435-f10f-45f0-b06f-..."
                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500 font-mono text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Role</label>
              <select
                value={newRole}
                onChange={(e) => setNewRole(e.target.value)}
                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500"
              >
                <option value="user">User</option>
                <option value="admin">Admin</option>
              </select>
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
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition"
              >
                Assign
              </button>
            </div>
          </form>
        )}

        {loading ? (
          <div className="text-center text-zinc-400 py-12">Loading...</div>
        ) : users.length === 0 ? (
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-12 text-center">
            <UserPlus className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
            <h3 className="font-semibold mb-2">No users assigned</h3>
            <p className="text-zinc-400 text-sm mb-4">
              Assign users to give them access to this VPN service
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {users.map((u) => (
              <div
                key={u.id}
                className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 flex items-center justify-between"
              >
                <div>
                  <div className="font-mono text-sm">{u.user_id}</div>
                  <div className="text-xs text-zinc-500 mt-0.5">
                    Role: <span className="text-zinc-400">{u.role}</span> ·
                    Granted:{" "}
                    {new Date(u.granted_at).toLocaleDateString()}
                  </div>
                </div>
                <button
                  onClick={() => handleUnassign(u.user_id)}
                  className="p-2 hover:bg-red-950 rounded text-zinc-500 hover:text-red-400 transition"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
