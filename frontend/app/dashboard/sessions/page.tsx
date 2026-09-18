"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Shield,
  LogOut,
  ArrowLeft,
  RefreshCw,
  Activity,
  Server,
  Globe,
  Clock,
  ArrowDown,
  ArrowUp,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import {
  vpnApi,
  VpnSession,
  VpnServer,
  formatBytes,
  timeAgo,
  duration,
} from "@/lib/sessions-api";

export default function SessionsPage() {
  const router = useRouter();
  const { user, initialized, bootstrap, logout } = useAuthStore();

  const [sessions, setSessions] = useState<VpnSession[]>([]);
  const [servers, setServers] = useState<VpnServer[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date());

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
      const interval = setInterval(() => {
        loadData(true);
      }, 30000);
      return () => clearInterval(interval);
    }
  }, [user]);

  async function loadData(silent = false) {
    if (!silent) setLoading(true);
    else setRefreshing(true);

    try {
      const [sess, srv] = await Promise.all([
        vpnApi.listSessions(),
        vpnApi.listServers(),
      ]);
      setSessions(sess);
      setServers(srv);
      setLastRefresh(new Date());
    } catch (err) {
      console.error("Failed to load data:", err);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }

  async function handleLogout() {
    await logout();
    router.push("/login");
  }

  const serverById = new Map(servers.map((s) => [s.id, s]));

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
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <Link
              href="/dashboard"
              className="p-2 hover:bg-zinc-800 rounded transition"
            >
              <ArrowLeft className="w-5 h-5" />
            </Link>
            <h1 className="text-2xl font-bold">VPN Sessions</h1>
          </div>
          <div className="flex items-center gap-3">
            <div className="text-xs text-zinc-500">
              Last refresh: {lastRefresh.toLocaleTimeString()}
            </div>
            <button
              onClick={() => loadData(true)}
              disabled={refreshing}
              className="flex items-center gap-2 px-3 py-1.5 text-sm bg-zinc-800 hover:bg-zinc-700 disabled:opacity-50 rounded transition"
            >
              <RefreshCw
                className={`w-4 h-4 ${refreshing ? "animate-spin" : ""}`}
              />
              Refresh
            </button>
          </div>
        </div>

        <div className="mb-6">
          <h2 className="text-sm font-medium text-zinc-400 mb-3 flex items-center gap-2">
            <Server className="w-4 h-4" />
            VPN Servers ({servers.length})
          </h2>
          {servers.length === 0 ? (
            <div className="text-sm text-zinc-500 bg-zinc-900 border border-zinc-800 rounded-lg p-4">
              No servers registered yet.
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {servers.map((s) => (
                <div
                  key={s.id}
                  className="bg-zinc-900 border border-zinc-800 rounded-lg p-3"
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-medium text-sm">{s.name}</span>
                    <span
                      className={`text-xs px-2 py-0.5 rounded ${
                        s.status === "online"
                          ? "bg-green-950 text-green-400"
                          : "bg-red-950 text-red-400"
                      }`}
                    >
                      {s.status}
                    </span>
                  </div>
                  <div className="text-xs text-zinc-500 space-y-1">
                    <div className="flex items-center gap-1.5">
                      <Globe className="w-3 h-3" />
                      {s.public_ip}
                    </div>
                    <div className="flex items-center gap-1.5">
                      <Clock className="w-3 h-3" />
                      {timeAgo(s.last_seen)}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div>
          <h2 className="text-sm font-medium text-zinc-400 mb-3 flex items-center gap-2">
            <Activity className="w-4 h-4" />
            Active Sessions ({sessions.length})
          </h2>

          {loading ? (
            <div className="text-center text-zinc-400 py-12">
              Loading sessions...
            </div>
          ) : sessions.length === 0 ? (
            <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-12 text-center">
              <Activity className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
              <h3 className="font-semibold mb-2">No active sessions</h3>
              <p className="text-zinc-400 text-sm">
                Connect a device to see active VPN sessions here.
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {sessions.map((sess) => {
                const server = sess.server_id
                  ? serverById.get(sess.server_id)
                  : undefined;
                return (
                  <div
                    key={sess.id}
                    className="bg-zinc-900 border border-zinc-800 rounded-lg p-4"
                  >
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
                        <span className="font-medium">
                          {server?.name || "Unknown server"}
                        </span>
                        <span className="text-xs text-zinc-500 font-mono">
                          {sess.id.slice(0, 8)}
                        </span>
                      </div>
                      <span className="text-xs text-zinc-500">
                        {timeAgo(sess.started_at)}
                      </span>
                    </div>

                    <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs">
                      <div>
                        <div className="text-zinc-500 mb-0.5">Client IP</div>
                        <div className="font-mono">{sess.client_ip}</div>
                      </div>
                      <div>
                        <div className="text-zinc-500 mb-0.5">Assigned IP</div>
                        <div className="font-mono">
                          {sess.assigned_ip || "N/A"}
                        </div>
                      </div>
                      <div>
                        <div className="text-zinc-500 mb-0.5">Duration</div>
                        <div>{duration(sess.started_at, sess.ended_at)}</div>
                      </div>
                      <div>
                        <div className="text-zinc-500 mb-0.5">Transfer</div>
                        <div className="flex items-center gap-2">
                          <span className="flex items-center gap-1 text-green-400">
                            <ArrowDown className="w-3 h-3" />
                            {formatBytes(sess.bytes_in)}
                          </span>
                          <span className="flex items-center gap-1 text-blue-400">
                            <ArrowUp className="w-3 h-3" />
                            {formatBytes(sess.bytes_out)}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
