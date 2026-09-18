import { api } from "./api";

export interface VpnServiceType {
  id: string;
  code: string;
  name: string;
  description: string;
  icon: string;
  category: "personal" | "corporate" | "enterprise";
  supports_split_tunnel: boolean;
  supports_kill_switch: boolean;
  supports_mfa: boolean;
  supports_policies: boolean;
  supports_ztna: boolean;
  default_routes: string[];
  default_dns: string[];
  default_mtu: number;
  display_order: number;
}

export interface VpnService {
  id: string;
  name: string;
  slug: string;
  service_type: string;
  description: string;
  subnet: string;
  server_ip: string;
  dns_servers: string[];
  max_clients: number;
  current_clients: number;
  status: "active" | "disabled" | "maintenance";
  is_public: boolean;
  custom_routes?: string[];
  type_name?: string;
  type_icon?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateServiceRequest {
  name: string;
  service_type: string;
  description?: string;
  subnet: string;
  server_ip: string;
  dns_servers?: string[];
  max_clients?: number;
  is_public?: boolean;
}

export interface ServiceUser {
  id: string;
  user_id: string;
  service_id: string;
  role: string;
  granted_at: string;
  expires_at?: string;
}

export const vpnServicesApi = {
  // Service types (public)
  listTypes: () =>
    api.get<{ types: VpnServiceType[] }>("/vpn/service-types").then((r) => r.data.types),

  getType: (code: string) =>
    api.get<VpnServiceType>(`/vpn/service-types/${code}`).then((r) => r.data),

  // Services (admin)
  list: () =>
    api.get<{ services: VpnService[] }>("/admin/vpn/services").then((r) => r.data.services),

  get: (id: string) =>
    api.get<VpnService>(`/admin/vpn/services/${id}`).then((r) => r.data),

  create: (req: CreateServiceRequest) =>
    api.post<VpnService>("/admin/vpn/services", req).then((r) => r.data),

  update: (id: string, req: Partial<CreateServiceRequest>) =>
    api.patch<VpnService>(`/admin/vpn/services/${id}`, req).then((r) => r.data),

  delete: (id: string) =>
    api.delete(`/admin/vpn/services/${id}`).then((r) => r.data),

  // User assignments
  listUsers: (serviceId: string) =>
    api.get<{ users: ServiceUser[] }>(`/admin/vpn/services/${serviceId}/users`).then((r) => r.data.users),

  assignUser: (serviceId: string, userId: string, role: string = "user") =>
    api.post(`/admin/vpn/services/${serviceId}/users`, { user_id: userId, role }).then((r) => r.data),

  unassignUser: (serviceId: string, userId: string) =>
    api.delete(`/admin/vpn/services/${serviceId}/users/${userId}`).then((r) => r.data),

  // Client
  myServices: () =>
    api.get<{ services: VpnService[] }>("/vpn/services").then((r) => r.data.services),

  getConfig: (serviceId: string) =>
    api.get(`/vpn/services/${serviceId}/config`).then((r) => r.data),
};

export function categoryColor(cat: string) {
  switch (cat) {
    case "personal":
      return "bg-blue-950 text-blue-400 border-blue-800";
    case "corporate":
      return "bg-purple-950 text-purple-400 border-purple-800";
    case "enterprise":
      return "bg-orange-950 text-orange-400 border-orange-800";
    default:
      return "bg-zinc-800 text-zinc-400";
  }
}

export function statusColor(status: string) {
  switch (status) {
    case "active":
      return "bg-green-950 text-green-400 border-green-800";
    case "disabled":
      return "bg-zinc-800 text-zinc-400 border-zinc-700";
    case "maintenance":
      return "bg-yellow-950 text-yellow-400 border-yellow-800";
    default:
      return "bg-zinc-800 text-zinc-400";
  }
}
