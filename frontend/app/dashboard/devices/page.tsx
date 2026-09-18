"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Shield,
  LogOut,
  Plus,
  Trash2,
  Download,
  Laptop,
  ArrowLeft,
  X,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import { devicesApi, Device, CreateDeviceResponse } from "@/lib/devices-api";

const PLATFORMS = [
  { value: "windows", label: "Windows" },
  { value: "linux", label: "Linux" },
  { value: "macos", label: "macOS" },
  { value: "android", label: "Android" },
  { value: "ios", label: "iOS" },
];

export default function DevicesPage() {
  const router = useRouter();
  const { user, initialized, bootstrap, logout } = useAuthStore();

  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAdd, setShowAdd] = useState(false);
  const [newName, setNewName] = useState("");
  const [newPlatform, setNewPlatform] = useState("windows");
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState("");
  const [newDevice, setNewDevice] = useState<CreateDeviceResponse | null>(null);

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
      loadDevices();
    }
  }, [user]);

  async function loadDevices() {
    setLoading(true);
    try {
      const list = await devicesApi.list();
      setDevices(list);
    } catch (err) {
      console.error("Failed to load devices:", err);
    } finally {
      setLoading(false);
    }
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setCreating(true);
    try {
      const resp = await devicesApi.create(newName, newPlatform);
      setNewDevice(resp);
      setDevices((prev) => [resp.device, ...prev]);
      setNewName("");
      setShowAdd(false);
    } catch (err: any) {
      setError(
        err?.response?.data?.message ||
          err?.response?.data?.error ||
          "Failed to create device"
      );
    } finally {
      setCreating(false);
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Delete this device? This cannot be undone.")) return;
    try {
      await devicesApi.delete(id);
      setDevices((prev) => prev.filter((d) => d.id !== id));
    } catch (err) {
      alert("Failed to delete device");
    }
  }

  function downloadConfig() {
    if (!newDevice) return;
    const blob = new Blob([newDevice.config], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `defendcore-${newDevice.device.name.replace(/\s+/g, "-")}.json`;
    a.click();
    URL.revokeObjectURL(url);
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

      <main className="max-w-7xl mx-auto px-4 py-8">
        {/* Breadcrumb + Add */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <Link
              href="/dashboard"
              className="p-2 hover:bg-zinc-800 rounded transition"
            >
              <ArrowLeft className="w-5 h-5" />
            </Link>
            <h1 className="text-2xl font-bold">Devices</h1>
          </div>
          <button
            onClick={() => setShowAdd(true)}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition"
          >
            <Plus className="w-4 h-4" />
            Add Device
          </button>
        </div>

        {/* Add Device Modal */}
        {showAdd && (
          <div className="fixed inset-0 bg-black/70 flex items-center justify-center p-4 z-50">
            <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 w-full max-w-md">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-bold">Add Device</h2>
                <button
                  onClick={() => setShowAdd(false)}
                  className="p-1 hover:bg-zinc-800 rounded"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              <form onSubmit={handleCreate} className="space-y-4">
                {error && (
                  <div className="p-3 bg-red-950 border border-red-800 rounded text-red-300 text-sm">
                    {error}
                  </div>
                )}

                <div>
                  <label className="block text-sm font-medium mb-1">
                    Device Name
                  </label>
                  <input
                    type="text"
                    value={newName}
                    onChange={(e) => setNewName(e.target.value)}
                    required
                    maxLength={64}
                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500"
                    placeholder="My Work Laptop"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">
                    Platform
                  </label>
                  <select
                    value={newPlatform}
                    onChange={(e) => setNewPlatform(e.target.value)}
                    className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-blue-500"
                  >
                    {PLATFORMS.map((p) => (
                      <option key={p.value} value={p.value}>
                        {p.label}
                      </option>
                    ))}
                  </select>
                </div>

                <button
                  type="submit"
                  disabled={creating}
                  className="w-full py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-zinc-700 rounded font-medium transition"
                >
                  {creating ? "Creating..." : "Create Device"}
                </button>
              </form>
            </div>
          </div>
        )}

        {/* New device config modal */}
        {newDevice && (
          <div className="fixed inset-0 bg-black/70 flex items-center justify-center p-4 z-50">
            <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 w-full max-w-lg">
              <h2 className="text-xl font-bold mb-2">Device Created ✅</h2>
              <p className="text-sm text-zinc-400 mb-4">
                Save this config now. Private key is shown only once.
              </p>

              <div className="bg-zinc-950 border border-zinc-800 rounded p-3 mb-4 max-h-64 overflow-auto">
                <pre className="text-xs font-mono text-zinc-300 whitespace-pre-wrap">
                  {newDevice.config}
                </pre>
              </div>

              <div className="flex gap-2">
                <button
                  onClick={downloadConfig}
                  className="flex-1 flex items-center justify-center gap-2 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition"
                >
                  <Download className="w-4 h-4" />
                  Download Config
                </button>
                <button
                  onClick={() => setNewDevice(null)}
                  className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded font-medium transition"
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Devices List */}
        {loading ? (
          <div className="text-center text-zinc-400 py-12">Loading devices...</div>
        ) : devices.length === 0 ? (
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-12 text-center">
            <Laptop className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
            <h3 className="font-semibold mb-2">No devices yet</h3>
            <p className="text-zinc-400 text-sm mb-4">
              Add your first device to get started.
            </p>
            <button
              onClick={() => setShowAdd(true)}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition"
            >
              Add Device
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {devices.map((d) => (
              <div
                key={d.id}
                className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 hover:border-zinc-700 transition"
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <Laptop className="w-5 h-5 text-blue-500" />
                    <h3 className="font-semibold">{d.name}</h3>
                  </div>
                  <button
                    onClick={() => handleDelete(d.id)}
                    className="p-1 hover:bg-red-950 rounded text-zinc-500 hover:text-red-400 transition"
                    title="Delete device"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
                <dl className="text-xs space-y-1 text-zinc-400">
                  <div className="flex justify-between">
                    <dt>Platform</dt>
                    <dd className="capitalize text-zinc-300">{d.platform}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt>Status</dt>
                    <dd className="text-green-400">{d.status}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt>Public Key</dt>
                    <dd className="font-mono text-zinc-500 truncate ml-2 max-w-[120px]">
                      {d.public_key.slice(0, 12)}...
                    </dd>
                  </div>
                </dl>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
