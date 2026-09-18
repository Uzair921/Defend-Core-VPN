"use client";

import { useState } from "react";
import { X, CreditCard, Loader2 } from "lucide-react";
import {
  organizationsApi,
  CreateSubscriptionInput,
} from "@/lib/organizations-api";

const SERVICE_TYPES = [
  { code: "full_tunnel", name: "Full Tunnel VPN", icon: "🔒" },
  { code: "split_tunnel", name: "Split Tunnel VPN", icon: "⚡" },
  { code: "remote_access", name: "Remote Access VPN", icon: "👤" },
  { code: "site_to_site", name: "Site-to-Site VPN", icon: "🏢" },
  { code: "zero_trust", name: "Zero Trust VPN", icon: "🔐" },
  { code: "stealth", name: "Stealth VPN", icon: "🎭" },
];

interface Props {
  orgId: string;
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export function SubscriptionModal({ orgId, open, onClose, onSuccess }: Props) {
  const [form, setForm] = useState<CreateSubscriptionInput>({
    service_type: "split_tunnel",
    quantity: 1,
    max_users: 50,
    price_cents_per_month: 9900,
  });
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState("");

  if (!open) return null;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setCreating(true);

    try {
      await organizationsApi.createSubscription(orgId, form);
      onSuccess();
      onClose();
      setForm({
        service_type: "split_tunnel",
        quantity: 1,
        max_users: 50,
        price_cents_per_month: 9900,
      });
    } catch (err: any) {
      setError(
        err?.response?.data?.message ||
          err?.response?.data?.error ||
          "Failed to create subscription"
      );
    } finally {
      setCreating(false);
    }
  }

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 z-50">
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg w-full max-w-lg">
        <div className="flex items-center justify-between px-5 py-4 border-b border-zinc-800">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-purple-950 border border-purple-900 flex items-center justify-center">
              <CreditCard className="w-5 h-5 text-purple-400" />
            </div>
            <div>
              <h2 className="font-semibold">Add VPN Service</h2>
              <p className="text-xs text-zinc-500">
                Allot a VPN service to this organization
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 hover:bg-zinc-800 rounded transition"
          >
            <X className="w-5 h-5 text-zinc-500" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          {error && (
            <div className="p-3 bg-red-950 border border-red-800 rounded text-red-300 text-sm">
              {error}
            </div>
          )}

          <div>
            <label className="block text-sm font-medium mb-2">
              Service Type *
            </label>
            <div className="grid grid-cols-2 gap-2">
              {SERVICE_TYPES.map((type) => (
                <button
                  key={type.code}
                  type="button"
                  onClick={() =>
                    setForm((f) => ({ ...f, service_type: type.code }))
                  }
                  className={`text-left p-3 rounded-lg border-2 transition ${
                    form.service_type === type.code
                      ? "border-purple-500 bg-purple-950/20"
                      : "border-zinc-800 bg-zinc-950/50 hover:border-zinc-700"
                  }`}
                >
                  <div className="text-xl mb-1">{type.icon}</div>
                  <div className="text-sm font-medium">{type.name}</div>
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">
              Max Users *
            </label>
            <input
              type="number"
              value={form.max_users}
              onChange={(e) =>
                setForm((f) => ({
                  ...f,
                  max_users: parseInt(e.target.value) || 0,
                }))
              }
              required
              min={1}
              max={10000}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-purple-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">
              Price per Month (cents) *
            </label>
            <input
              type="number"
              value={form.price_cents_per_month}
              onChange={(e) =>
                setForm((f) => ({
                  ...f,
                  price_cents_per_month: parseInt(e.target.value) || 0,
                }))
              }
              required
              min={0}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-purple-500"
            />
            <p className="text-xs text-zinc-500 mt-1">
              ${((form.price_cents_per_month || 0) / 100).toFixed(2)} per month
            </p>
          </div>

          <div className="flex gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded font-medium transition"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={creating}
              className="flex-1 px-4 py-2 bg-purple-600 hover:bg-purple-700 disabled:bg-zinc-700 rounded font-medium transition flex items-center justify-center gap-2"
            >
              {creating ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Creating...
                </>
              ) : (
                "Add Service"
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
