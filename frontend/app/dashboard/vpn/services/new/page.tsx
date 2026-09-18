"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Shield, LogOut, ArrowLeft, CheckCircle } from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import {
  vpnServicesApi,
  VpnServiceType,
  categoryColor,
} from "@/lib/vpn-services-api";

export default function NewServicePage() {
  const router = useRouter();
  const { user, initialized, bootstrap, logout } = useAuthStore();

  const [types, setTypes] = useState<VpnServiceType[]>([]);
  const [selectedType, setSelectedType] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState("");

  // Form fields
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [subnet, setSubnet] = useState("10.9.0.0/24");
  const [serverIp, setServerIp] = useState("10.9.0.1");
  const [maxClients, setMaxClients] = useState(100);
  const [isPublic, setIsPublic] = useState(false);

  useEffect(() => {
    bootstrap();
  }, [bootstrap]);

  useEffect(() => {
    if (initialized && !user) {
      router.push("/login");
    }
  }, [initialized, user, router]);

  useEffect(() => {
    loadTypes();
  }, []);

  async function loadTypes() {
    try {
      const list = await vpnServicesApi.listTypes();
      setTypes(list);
    } catch (err) {
      console.error("Failed to load types:", err);
    } finally {
      setLoading(false);
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedType) {
      setError("Please select a service type");
      return;
    }

    setError("");
    setCreating(true);

    try {
      const created = await vpnServicesApi.create({
        name,
        service_type: selectedType,
        description: description || undefined,
        subnet,
        server_ip: serverIp,
        max_clients: maxClients,
        is_public: isPublic,
      });
      router.push("/dashboard/vpn/services");
    } catch (err: any) {
      setError(
        err?.response?.data?.message ||
          err?.response?.data?.error ||
          "Failed to create service"
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
      {/* Nav */}
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
        <div className="flex items-center gap-3 mb-6">
          <Link
            href="/dashboard/vpn/services"
            className="p-2 hover:bg-zinc-800 rounded transition"
          >
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold">Create VPN Service</h1>
            <p className="text-sm text-zinc-400">
              Choose a type and configure your service
            </p>
          </div>
        </div>

        {error && (
          <div className="mb-4 p-3 bg-red-950 border border-red-800 rounded text-red-300 text-sm">
            {error}
          </div>
        )}

        {/* Step 1: Type Selection */}
        <div className="mb-8">
          <h2 className="text-lg font-semibold mb-3">Step 1: Choose Type</h2>
          {loading ? (
            <div className="text-zinc-400">Loading types...</div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {types.map((t) => (
                <button
                  key={t.code}
                  type="button"
                  onClick={() => setSelectedType(t.code)}
                  className={`text-left p-4 rounded-lg border-2 transition ${
                    selectedType === t.code
                      ? "border-blue-500 bg-blue-950/20"
                      : "border-zinc-800 bg-zinc-900 hover:border-zinc-700"
                  }`}
                >
                  <div className="flex items-start gap-3">
                    <div className="text-2xl">{t.icon}</div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-semibold">{t.name}</span>
                        {selectedType === t.code && (
                          <CheckCircle className="w-4 h-4 text-blue-400" />
                        )}
                      </div>
                      <p className="text-xs text-zinc-400 mb-2">
                        {t.description}
                      </p>
                      <span
                        className={`text-xs px-2 py-0.5 rounded border ${categoryColor(
                          t.category
                        )}`}
                      >
                        {t.category}
                      </span>
                    </div>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Step 2: Configuration */}
        <form onSubmit={handleSubmit}>
          <h2 className="text-lg font-semibold mb-3">Step 2: Configure</h2>
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 space-y-4 mb-6">
            <div>
              <label className="block text-sm font-medium mb-1">
                Service Name *
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                placeholder="Corporate Split VPN"
                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">
                Description
              </label>
              <input
                type="text"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Dev team split tunnel"
                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500"
              />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium mb-1">
                  Subnet *
                </label>
                <input
                  type="text"
                  value={subnet}
                  onChange={(e) => setSubnet(e.target.value)}
                  required
                  className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500 font-mono text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">
                  Server IP *
                </label>
                <input
                  type="text"
                  value={serverIp}
                  onChange={(e) => setServerIp(e.target.value)}
                  required
                  className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500 font-mono text-sm"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">
                Max Clients
              </label>
              <input
                type="number"
                value={maxClients}
                onChange={(e) => setMaxClients(parseInt(e.target.value) || 100)}
                min={1}
                max={10000}
                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500"
              />
            </div>

            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={isPublic}
                onChange={(e) => setIsPublic(e.target.checked)}
                className="w-4 h-4"
              />
              <span className="text-sm">
                Public service (visible to all users)
              </span>
            </label>
          </div>

          <div className="flex gap-3">
            <Link
              href="/dashboard/vpn/services"
              className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded font-medium transition"
            >
              Cancel
            </Link>
            <button
              type="submit"
              disabled={creating || !selectedType}
              className="px-6 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-zinc-700 rounded font-medium transition"
            >
              {creating ? "Creating..." : "Create Service"}
            </button>
          </div>
        </form>
      </main>
    </div>
  );
}
