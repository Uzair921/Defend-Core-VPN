// Package organizations implements the multi-tenant SaaS layer. Each
// organization represents a customer that subscribes to one or more
// VPN services and manages its own users.
package organizations

import (
"encoding/json"
"time"

"github.com/google/uuid"
)

// Status represents the lifecycle state of an organization.
type Status string

const (
StatusActive    Status = "active"
StatusSuspended Status = "suspended"
StatusTrial     Status = "trial"
StatusCancelled Status = "cancelled"
)

// Plan represents the subscription tier of an organization.
type Plan string

const (
PlanTrial      Plan = "trial"
PlanBasic      Plan = "basic"
PlanPro        Plan = "pro"
PlanEnterprise Plan = "enterprise"
)

// Role represents a user's role within an organization.
type Role string

const (
RoleAdmin Role = "admin"
RoleUser  Role = "user"
)

// SubscriptionStatus represents the lifecycle state of a subscription.
type SubscriptionStatus string

const (
SubscriptionActive    SubscriptionStatus = "active"
SubscriptionSuspended SubscriptionStatus = "suspended"
SubscriptionCancelled SubscriptionStatus = "cancelled"
)

// InvoiceStatus represents the lifecycle state of an invoice.
type InvoiceStatus string

const (
InvoicePending   InvoiceStatus = "pending"
InvoicePaid      InvoiceStatus = "paid"
InvoiceOverdue   InvoiceStatus = "overdue"
InvoiceCancelled InvoiceStatus = "cancelled"
)

// Organization represents a customer (tenant) on the platform.
type Organization struct {
ID                uuid.UUID       `json:"id"`
Name              string          `json:"name"`
Slug              string          `json:"slug"`
Email             string          `json:"email"`
Phone             string          `json:"phone,omitempty"`
Website           string          `json:"website,omitempty"`
Status            Status          `json:"status"`
Plan              Plan            `json:"plan"`
MaxUsers          int             `json:"max_users"`
MaxServices       int             `json:"max_services"`
BandwidthLimitGB  *int            `json:"bandwidth_limit_gb,omitempty"`
BillingEmail      string          `json:"billing_email,omitempty"`
SubscriptionStart *time.Time      `json:"subscription_start,omitempty"`
SubscriptionEnd   *time.Time      `json:"subscription_end,omitempty"`
MonthlyPriceCents *int            `json:"monthly_price_cents,omitempty"`
Metadata          json.RawMessage `json:"metadata,omitempty"`
CreatedAt         time.Time       `json:"created_at"`
UpdatedAt         time.Time       `json:"updated_at"`
}

// IsActive reports whether the organization can currently use the platform.
func (o *Organization) IsActive() bool {
return o.Status == StatusActive || o.Status == StatusTrial
}

// OrganizationUser links a platform user to an organization.
type OrganizationUser struct {
ID             uuid.UUID `json:"id"`
OrganizationID uuid.UUID `json:"organization_id"`
UserID         uuid.UUID `json:"user_id"`
Role           Role      `json:"role"`
Status         string    `json:"status"`
JoinedAt       time.Time `json:"joined_at"`
}

// Subscription represents a VPN service allotment to an organization.
type Subscription struct {
ID                 uuid.UUID          `json:"id"`
OrganizationID     uuid.UUID          `json:"organization_id"`
ServiceType        string             `json:"service_type"`
Quantity           int                `json:"quantity"`
MaxUsers           int                `json:"max_users"`
Status             SubscriptionStatus `json:"status"`
PriceCentsPerMonth *int               `json:"price_cents_per_month,omitempty"`
StartedAt          time.Time          `json:"started_at"`
ExpiresAt          *time.Time         `json:"expires_at,omitempty"`
}

// Invoice represents a billing record for an organization.
type Invoice struct {
ID             uuid.UUID       `json:"id"`
OrganizationID uuid.UUID       `json:"organization_id"`
InvoiceNumber  string          `json:"invoice_number"`
AmountCents    int             `json:"amount_cents"`
Currency       string          `json:"currency"`
Status         InvoiceStatus   `json:"status"`
PeriodStart    time.Time       `json:"period_start"`
PeriodEnd      time.Time       `json:"period_end"`
PaidAt         *time.Time      `json:"paid_at,omitempty"`
DueAt          *time.Time      `json:"due_at,omitempty"`
Items          json.RawMessage `json:"items,omitempty"`
CreatedAt      time.Time       `json:"created_at"`
}
