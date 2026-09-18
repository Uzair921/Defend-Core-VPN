import { api } from "./api";

// =====================================================
// Types
// =====================================================

export type OrgStatus = "active" | "suspended" | "trial" | "cancelled";
export type OrgPlan = "trial" | "basic" | "pro" | "enterprise";
export type OrgRole = "admin" | "user";
export type SubscriptionStatus = "active" | "suspended" | "cancelled";
export type InvoiceStatus = "pending" | "paid" | "overdue" | "cancelled";

export interface Organization {
  id: string;
  name: string;
  slug: string;
  email: string;
  phone?: string;
  website?: string;
  status: OrgStatus;
  plan: OrgPlan;
  max_users: number;
  max_services: number;
  bandwidth_limit_gb?: number;
  billing_email?: string;
  subscription_start?: string;
  subscription_end?: string;
  monthly_price_cents?: number;
  created_at: string;
  updated_at: string;
}

export interface OrganizationUser {
  id: string;
  organization_id: string;
  user_id: string;
  role: OrgRole;
  status: string;
  joined_at: string;
}

export interface Subscription {
  id: string;
  organization_id: string;
  service_type: string;
  quantity: number;
  max_users: number;
  status: SubscriptionStatus;
  price_cents_per_month?: number;
  started_at: string;
  expires_at?: string;
}

export interface Invoice {
  id: string;
  organization_id: string;
  invoice_number: string;
  amount_cents: number;
  currency: string;
  status: InvoiceStatus;
  period_start: string;
  period_end: string;
  paid_at?: string;
  due_at?: string;
  created_at: string;
}

export interface CreateOrganizationInput {
  name: string;
  email: string;
  plan: OrgPlan;
  phone?: string;
  website?: string;
  max_users?: number;
  max_services?: number;
  billing_email?: string;
  monthly_price_cents?: number;
}

export interface UpdateOrganizationInput {
  name?: string;
  email?: string;
  status?: OrgStatus;
  plan?: OrgPlan;
  max_users?: number;
  max_services?: number;
}

export interface CreateSubscriptionInput {
  service_type: string;
  quantity?: number;
  max_users?: number;
  price_cents_per_month?: number;
  expires_at?: string;
}

export interface CreateInvoiceInput {
  amount_cents: number;
  currency?: string;
  period_start: string;
  period_end: string;
  due_at?: string;
  items?: any;
}

export interface Paginated<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

// =====================================================
// API Client
// =====================================================

export const organizationsApi = {
  // List organizations (SuperAdmin)
  list: (params?: { status?: string; plan?: string; search?: string; limit?: number; offset?: number }) => {
    const q = new URLSearchParams();
    if (params?.status) q.set("status", params.status);
    if (params?.plan) q.set("plan", params.plan);
    if (params?.search) q.set("search", params.search);
    if (params?.limit) q.set("limit", String(params.limit));
    if (params?.offset) q.set("offset", String(params.offset));
    const qs = q.toString();
    return api
      .get<Paginated<Organization>>(`/superadmin/organizations${qs ? "?" + qs : ""}`)
      .then((r) => r.data);
  },

  // Get single org
  get: (id: string) =>
    api.get<Organization>(`/superadmin/organizations/${id}`).then((r) => r.data),

  // Create org
  create: (input: CreateOrganizationInput) =>
    api.post<Organization>("/superadmin/organizations", input).then((r) => r.data),

  // Update org
  update: (id: string, input: UpdateOrganizationInput) =>
    api.patch<Organization>(`/superadmin/organizations/${id}`, input).then((r) => r.data),

  // Delete org
  delete: (id: string) =>
    api.delete(`/superadmin/organizations/${id}`).then((r) => r.data),

  // Users
  listUsers: (orgId: string) =>
    api
      .get<{ users: OrganizationUser[] }>(`/superadmin/organizations/${orgId}/users`)
      .then((r) => r.data.users),

  addUser: (orgId: string, userId: string, role: OrgRole = "user") =>
    api
      .post<OrganizationUser>(`/superadmin/organizations/${orgId}/users`, {
        user_id: userId,
        role,
      })
      .then((r) => r.data),

  removeUser: (orgId: string, userId: string) =>
    api
      .delete(`/superadmin/organizations/${orgId}/users/${userId}`)
      .then((r) => r.data),

  // Subscriptions
  listSubscriptions: (orgId: string) =>
    api
      .get<{ subscriptions: Subscription[] }>(`/superadmin/organizations/${orgId}/subscriptions`)
      .then((r) => r.data.subscriptions),

  createSubscription: (orgId: string, input: CreateSubscriptionInput) =>
    api
      .post<Subscription>(`/superadmin/organizations/${orgId}/subscriptions`, input)
      .then((r) => r.data),

  updateSubscriptionStatus: (subId: string, status: SubscriptionStatus) =>
    api
      .patch<Subscription>(`/superadmin/subscriptions/${subId}`, { status })
      .then((r) => r.data),

  deleteSubscription: (subId: string) =>
    api.delete(`/superadmin/subscriptions/${subId}`).then((r) => r.data),

  // Invoices
  listInvoices: (orgId: string, limit = 20, offset = 0) =>
    api
      .get<{ invoices: Invoice[] }>(
        `/superadmin/organizations/${orgId}/invoices?limit=${limit}&offset=${offset}`
      )
      .then((r) => r.data.invoices),

  createInvoice: (orgId: string, input: CreateInvoiceInput) =>
    api
      .post<Invoice>(`/superadmin/organizations/${orgId}/invoices`, input)
      .then((r) => r.data),

  markInvoicePaid: (invoiceId: string) =>
    api.post<Invoice>(`/superadmin/invoices/${invoiceId}/paid`).then((r) => r.data),
};

// =====================================================
// Org Admin (self-service) API
// =====================================================

export const orgApi = {
  getMyOrganization: () =>
    api.get<Organization>("/org/me").then((r) => r.data),

  listMyUsers: () =>
    api.get<{ users: OrganizationUser[] }>("/org/me/users").then((r) => r.data.users),

  listMySubscriptions: () =>
    api
      .get<{ subscriptions: Subscription[] }>("/org/me/subscriptions")
      .then((r) => r.data.subscriptions),

  listMyInvoices: (limit = 20, offset = 0) =>
    api
      .get<{ invoices: Invoice[] }>(`/org/me/invoices?limit=${limit}&offset=${offset}`)
      .then((r) => r.data.invoices),
};

// =====================================================
// Display helpers
// =====================================================

export function planColor(plan: OrgPlan): string {
  switch (plan) {
    case "trial":
      return "bg-zinc-800 text-zinc-300 border-zinc-700";
    case "basic":
      return "bg-blue-950 text-blue-400 border-blue-800";
    case "pro":
      return "bg-purple-950 text-purple-400 border-purple-800";
    case "enterprise":
      return "bg-orange-950 text-orange-400 border-orange-800";
    default:
      return "bg-zinc-800 text-zinc-400";
  }
}

export function statusColor(status: OrgStatus | SubscriptionStatus): string {
  switch (status) {
    case "active":
      return "bg-green-950 text-green-400 border-green-800";
    case "trial":
      return "bg-yellow-950 text-yellow-400 border-yellow-800";
    case "suspended":
      return "bg-red-950 text-red-400 border-red-800";
    case "cancelled":
      return "bg-zinc-800 text-zinc-500 border-zinc-700";
    default:
      return "bg-zinc-800 text-zinc-400";
  }
}

export function invoiceStatusColor(status: InvoiceStatus): string {
  switch (status) {
    case "paid":
      return "bg-green-950 text-green-400 border-green-800";
    case "pending":
      return "bg-yellow-950 text-yellow-400 border-yellow-800";
    case "overdue":
      return "bg-red-950 text-red-400 border-red-800";
    case "cancelled":
      return "bg-zinc-800 text-zinc-500 border-zinc-700";
    default:
      return "bg-zinc-800 text-zinc-400";
  }
}

export function formatPrice(cents?: number): string {
  if (cents === undefined || cents === null) return "—";
  return `$${(cents / 100).toFixed(2)}`;
}

export function formatDate(iso?: string): string {
  if (!iso) return "—";
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}
