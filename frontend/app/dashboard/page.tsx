"use client";

import { useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/lib/auth-store";
import { LogOut, Shield, User as UserIcon } from "lucide-react";

export default function DashboardPage() {
  const router = useRouter();
  const { user, initialized, bootstrap, logout } = useAuthStore();

  useEffect(() => {
    bootstrap();
  }, [bootstrap]);

  useEffect(() => {
    if (initialized && !user) {
      router.push("/login");
    }
  }, [initialized, user, router]);

  if (!initialized || !user) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-zinc-950 text-white">
        <div className="text-zinc-400">Loading...</div>
      </div>
    );
  }

  async function handleLogout() {
    await logout();
    router.push("/login");
  }

  return (
    <div className="min-h-screen bg-zinc-950 text-white">
      <nav className="border-b border-zinc-800 bg-zinc-900">
        <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Shield className="w-6 h-6 text-blue-500" />
            <span className="font-bold text-lg">DefendCore VPN</span>
          </div>
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
        <h1 className="text-2xl font-bold mb-6">Dashboard</h1>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6">
            <div className="flex items-center gap-3 mb-4">
              <UserIcon className="w-5 h-5 text-blue-500" />
              <h2 className="font-semibold">Account</h2>
            </div>
            <dl className="space-y-2 text-sm">
              <div className="flex justify-between">
                <dt className="text-zinc-400">Email</dt>
                <dd>{user.email}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-zinc-400">Role</dt>
                <dd className="capitalize">{user.role}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-zinc-400">User ID</dt>
                <dd className="font-mono text-xs">{user.id.slice(0, 8)}...</dd>
              </div>
            </dl>
          </div>

          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6">
            <h2 className="font-semibold mb-4">VPN Status</h2>
            <div className="text-sm text-zinc-400 mb-4">
              Manage your VPN devices and download configs.
            </div>
            <div className="flex gap-2">
              <Link
                href="/dashboard/devices"
                className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition text-sm"
              >
                Manage Devices
              </Link>
              <Link
                href="/dashboard/sessions"
                className="inline-flex items-center gap-2 px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded font-medium transition text-sm"
              >
                View Sessions
              </Link>
              <Link
                href="/dashboard/policies"
                className="inline-flex items-center gap-2 px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded font-medium transition text-sm"
              >
                Access Policies
              </Link>
              <Link
                href="/dashboard/vpn/services"
                className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition text-sm"
              >
                VPN Services
              </Link>
              <Link
                href="/dashboard/superadmin"
                className="inline-flex items-center gap-2 px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded font-medium transition text-sm"
              >
                SuperAdmin
              </Link>
              <Link
                href="/dashboard/org"
                className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition text-sm"
              >
                My Organization
              </Link>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
