import { api } from "./api";

export type PolicyType = "cidr" | "domain" | "url" | "ip" | "port";
export type PolicyAction = "allow" | "deny";

export interface AccessPolicy {
  id: string;
  user_id: string;
  device_id?: string;
  type: PolicyType;
  value: string;
  action: PolicyAction;
  priority: number;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface CreatePolicyRequest {
  user_id: string;
  device_id?: string;
  type: PolicyType;
  value: string;
  action: PolicyAction;
  priority?: number;
  description?: string;
}

export const policiesApi = {
  listByUser: (userId: string) =>
    api
      .get<{ policies: AccessPolicy[] }>(`/admin/policies/user/${userId}`)
      .then((r) => r.data.policies),

  create: (req: CreatePolicyRequest) =>
    api.post<AccessPolicy>("/admin/policies", req).then((r) => r.data),

  update: (id: string, patch: Partial<CreatePolicyRequest>) =>
    api.patch<AccessPolicy>(`/admin/policies/${id}`, patch).then((r) => r.data),

  delete: (id: string) =>
    api.delete(`/admin/policies/${id}`).then((r) => r.data),
};

export function actionColor(action: PolicyAction) {
  return action === "allow"
    ? "bg-green-950 text-green-400 border-green-800"
    : "bg-red-950 text-red-400 border-red-800";
}
