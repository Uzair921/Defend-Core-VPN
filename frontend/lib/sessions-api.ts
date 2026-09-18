import { api } from "./api";

export interface VpnServer {
  id: string;
  name: string;
  public_ip: string;
  region?: string;
  status: string;
  version?: string;
  last_seen: string;
  created_at: string;
}

export interface VpnSession {
  id: string;
  user_id: string;
  device_id?: string;
  server_id?: string;
  client_ip: string;
  assigned_ip: string;
  started_at: string;
  ended_at?: string;
  bytes_in: number;
  bytes_out: number;
  status: string;
}

export const vpnApi = {
  listServers: () =>
    api.get<{ servers: VpnServer[] }>("/vpn/servers").then((r) => r.data.servers),

  listSessions: () =>
    api.get<{ sessions: VpnSession[] }>("/vpn/sessions").then((r) => r.data.sessions),

  getSession: (id: string) =>
    api.get<VpnSession>(`/vpn/sessions/${id}`).then((r) => r.data),
};

export function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
}

export function timeAgo(iso: string): string {
  const now = Date.now();
  const then = new Date(iso).getTime();
  const diff = Math.floor((now - then) / 1000);
  if (diff < 60) return `${diff}s ago`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

export function duration(startISO: string, endISO?: string): string {
  const start = new Date(startISO).getTime();
  const end = endISO ? new Date(endISO).getTime() : Date.now();
  const diff = Math.floor((end - start) / 1000);
  const h = Math.floor(diff / 3600);
  const m = Math.floor((diff % 3600) / 60);
  const s = diff % 60;
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m ${s}s`;
  return `${s}s`;
}
