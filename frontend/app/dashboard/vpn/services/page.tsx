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
  Users,
  Activity,
  Settings,
  XCircle,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import {
  vpnServicesApi,
  VpnService,
  categoryColor,
  statusColor,
} from "@/lib/vpn-services-api";

export default function VpnServicesPage() {
  const router = useRouter();
  const { user, initialized, bootstrap, logout } = useAuthStore();

  const [services, setServices] = useState<VpnService[]>([]);
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
      loadServices();
    }
  }, [user]);

  async function loadServices() {
    setLoading(true);
    try {
      const list = await vpnServicesApi.list();
      setServices(list);
    } catch (err) {
      console.error("Failed to load services:", err);
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete(id: string, name: string) {
    if (!confirm(`Delete service "${name}"? This cannot be undone.`)) return;
    try {
      await vpnServicesApi.delete(id);
      setServices((prev) => prev.filter((s) => s.id !== id));
    } catch (err) {
      alert("Failed to delete service");
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
              <h1 className="text-2xl font-bold">VPN Services</h1>
              <p className="text-sm text-zinc-400">
                Manage multi-VPN service instances
              </p>
            </div>
          </div>
          <Link
            href="/dashboard/vpn/services/new"
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition"
          >
            <Plus className="w-4 h-4" />
            New Service
          </Link>
        </div>

        {/* Info box */}
        <div className="bg-blue-950/30 border border-blue-900 rounded-lg p-4 mb-6 text-sm text-blue-200">
          <strong>Multi-VPN Platform:</strong> Create multiple VPN services
          (Full Tunnel, Split Tunnel, Remote Access, etc.) and assign users to
          each. Clients can switch between services.
        </div>

        {/* Services List */}
        {loading ? (
          <div className="text-center text-zinc-400 py-12">
            Loading services...
          </div>
        ) : services.length === 0 ? (
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-12 text-center">
            <Shield className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
            <h3 className="font-semibold mb-2">No services yet</h3>
            <p className="text-zinc-400 text-sm mb-4">
              Create your first VPN service from 6 available types
            </p>
            <Link
              href="/dashboard/vpn/services/new"
              className="inline-block px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition"
            >
              Create First Service
            </Link>
          </div>
        ) : (
          <div className="space-y-3">
            {services.map((s) => (
              <div
                key={s.id}
                className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 hover:border-zinc-700 transition"
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-start gap-3 flex-1 min-w-0">
                    <div className="text-3xl flex-shrink-0">
                      {s.type_icon || "🔒"}
                    </div>

                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2 flex-wrap mb-1">
                        <h3 className="font-semibold text-lg">{s.name}</h3>
                        <span
                          className={`text-xs px-2 py-0.5 rounded border ${statusColor(
                            s.status
                          )}`}
                        >
                          {s.status}
                        </span>
                        <span className="text-xs text-zinc-500">
                          {s.type_name}
                        </span>
                      </div>

                      {s.description && (
                        <p className="text-sm text-zinc-400 mb-2">
                          {s.description}
                        </p>
                      )}

                      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs text-zinc-500">
                        <div>
                          <span className="text-zinc-600">Subnet: </span>
                          <span className="font-mono">{s.subnet}</span>
                        </div>
                        <div>
                          <span className="text-zinc-600">Server: </span>
                          <span className="font-mono">{s.server_ip}</span>
                        </div>
                        <div>
                          <span className="text-zinc-600">Clients: </span>
                          <span>
                            {s.current_clients}/{s.max_clients}
                          </span>
                        </div>
                        <div>
                          <span className="text-zinc-600">Slug: </span>
                          <span className="font-mono truncate">{s.slug}</span>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-1 ml-2">
                    <Link
                      href={`/dashboard/vpn/services/${s.id}/users`}
                      className="p-2 hover:bg-zinc-800 rounded text-zinc-500 hover:text-blue-400 transition"
                      title="Manage users"
                    >
                      <Users className="w-4 h-4" />
                    </Link>
                    <button
                      onClick={() => handleDelete(s.id, s.name)}
                      className="p-2 hover:bg-red-950 rounded text-zinc-500 hover:text-red-400 transition"
                      title="Delete service"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
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
