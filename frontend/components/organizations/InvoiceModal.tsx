"use client";

import { useState } from "react";
import { X, DollarSign, Loader2 } from "lucide-react";
import {
  organizationsApi,
  CreateInvoiceInput,
  formatPrice,
} from "@/lib/organizations-api";

interface Props {
  orgId: string;
  suggestedAmount?: number;
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export function InvoiceModal({
  orgId,
  suggestedAmount = 0,
  open,
  onClose,
  onSuccess,
}: Props) {
  const today = new Date();
  const nextMonth = new Date();
  nextMonth.setMonth(nextMonth.getMonth() + 1);
  const dueDate = new Date();
  dueDate.setDate(dueDate.getDate() + 30);

  const [form, setForm] = useState<CreateInvoiceInput>({
    amount_cents: suggestedAmount || 9900,
    currency: "USD",
    period_start: today.toISOString().split("T")[0],
    period_end: nextMonth.toISOString().split("T")[0],
    due_at: dueDate.toISOString().split("T")[0],
  });
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState("");

  if (!open) return null;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setCreating(true);

    try {
      // Convert date-only strings to RFC3339 (ISO 8601) format
      const payload = {
        ...form,
        period_start: new Date(form.period_start + "T00:00:00Z").toISOString(),
        period_end: new Date(form.period_end + "T23:59:59Z").toISOString(),
        due_at: form.due_at
          ? new Date(form.due_at + "T00:00:00Z").toISOString()
          : undefined,
      };
      await organizationsApi.createInvoice(orgId, payload);
      onSuccess();
      onClose();
    } catch (err: any) {
      setError(
        err?.response?.data?.message ||
          err?.response?.data?.error ||
          "Failed to create invoice"
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
            <div className="w-9 h-9 rounded-lg bg-green-950 border border-green-900 flex items-center justify-center">
              <DollarSign className="w-5 h-5 text-green-400" />
            </div>
            <div>
              <h2 className="font-semibold">Create Invoice</h2>
              <p className="text-xs text-zinc-500">
                Generate a new billing invoice
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
            <label className="block text-sm font-medium mb-1">
              Amount (cents) *
            </label>
            <input
              type="number"
              value={form.amount_cents}
              onChange={(e) =>
                setForm((f) => ({
                  ...f,
                  amount_cents: parseInt(e.target.value) || 0,
                }))
              }
              required
              min={0}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-green-500"
            />
            <p className="text-xs text-zinc-500 mt-1">
              {formatPrice(form.amount_cents)}
            </p>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium mb-1">
                Period Start *
              </label>
              <input
                type="date"
                value={form.period_start}
                onChange={(e) =>
                  setForm((f) => ({ ...f, period_start: e.target.value }))
                }
                required
                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-green-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">
                Period End *
              </label>
              <input
                type="date"
                value={form.period_end}
                onChange={(e) =>
                  setForm((f) => ({ ...f, period_end: e.target.value }))
                }
                required
                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-green-500"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Due Date</label>
            <input
              type="date"
              value={form.due_at || ""}
              onChange={(e) =>
                setForm((f) => ({ ...f, due_at: e.target.value }))
              }
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded focus:outline-none focus:border-green-500"
            />
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
              className="flex-1 px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-zinc-700 rounded font-medium transition flex items-center justify-center gap-2"
            >
              {creating ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Creating...
                </>
              ) : (
                "Create Invoice"
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
