package organizations

import (
"time"

"github.com/google/uuid"
)

// CreateOrganizationRequest is the payload for creating a new organization.
type CreateOrganizationRequest struct {
Name              string `json:"name" validate:"required,min=2,max=120"`
Slug              string `json:"slug,omitempty" validate:"omitempty,min=2,max=60"`
Email             string `json:"email" validate:"required,email"`
Phone             string `json:"phone,omitempty" validate:"omitempty,max=32"`
Website           string `json:"website,omitempty" validate:"omitempty,url"`
Plan              Plan   `json:"plan" validate:"required,oneof=trial basic pro enterprise"`
MaxUsers          int    `json:"max_users" validate:"omitempty,min=1,max=100000"`
MaxServices       int    `json:"max_services" validate:"omitempty,min=1,max=50"`
BandwidthLimitGB  *int   `json:"bandwidth_limit_gb,omitempty" validate:"omitempty,min=1"`
BillingEmail      string `json:"billing_email,omitempty" validate:"omitempty,email"`
MonthlyPriceCents *int   `json:"monthly_price_cents,omitempty" validate:"omitempty,min=0"`
}

// UpdateOrganizationRequest is the payload for updating an organization.
type UpdateOrganizationRequest struct {
Name              *string    `json:"name,omitempty" validate:"omitempty,min=2,max=120"`
Email             *string    `json:"email,omitempty" validate:"omitempty,email"`
Phone             *string    `json:"phone,omitempty" validate:"omitempty,max=32"`
Website           *string    `json:"website,omitempty" validate:"omitempty,url"`
Status            *Status    `json:"status,omitempty" validate:"omitempty,oneof=active suspended trial cancelled"`
Plan              *Plan      `json:"plan,omitempty" validate:"omitempty,oneof=trial basic pro enterprise"`
MaxUsers          *int       `json:"max_users,omitempty" validate:"omitempty,min=1,max=100000"`
MaxServices       *int       `json:"max_services,omitempty" validate:"omitempty,min=1,max=50"`
BandwidthLimitGB  *int       `json:"bandwidth_limit_gb,omitempty"`
BillingEmail      *string    `json:"billing_email,omitempty" validate:"omitempty,email"`
SubscriptionEnd   *time.Time `json:"subscription_end,omitempty"`
MonthlyPriceCents *int       `json:"monthly_price_cents,omitempty"`
}

// AddOrganizationUserRequest is the payload for adding a user to an org.
type AddOrganizationUserRequest struct {
UserID uuid.UUID `json:"user_id" validate:"required"`
Role   Role      `json:"role" validate:"omitempty,oneof=admin user"`
}

// InviteOrganizationAdminRequest creates a new user and adds them as admin.
type InviteOrganizationAdminRequest struct {
Email    string `json:"email" validate:"required,email"`
FullName string `json:"full_name,omitempty" validate:"omitempty,max=120"`
Role     Role   `json:"role" validate:"omitempty,oneof=admin user"`
}

// CreateSubscriptionRequest is the payload for allotting a VPN service.
type CreateSubscriptionRequest struct {
ServiceType        string     `json:"service_type" validate:"required"`
Quantity           int        `json:"quantity" validate:"omitempty,min=1,max=100"`
MaxUsers           int        `json:"max_users" validate:"omitempty,min=1"`
PriceCentsPerMonth *int       `json:"price_cents_per_month,omitempty"`
ExpiresAt          *time.Time `json:"expires_at,omitempty"`
}

// ListOrganizationsFilter is used for paginated listing.
type ListOrganizationsFilter struct {
Status *Status
Plan   *Plan
Search string
Limit  int
Offset int
}

// OrganizationSummary is a lightweight view for list endpoints.
type OrganizationSummary struct {
ID           uuid.UUID `json:"id"`
Name         string    `json:"name"`
Slug         string    `json:"slug"`
Status       Status    `json:"status"`
Plan         Plan      `json:"plan"`
MaxUsers     int       `json:"max_users"`
CurrentUsers int       `json:"current_users"`
CreatedAt    time.Time `json:"created_at"`
}

// PaginatedResponse wraps a list with pagination metadata.
type PaginatedResponse[T any] struct {
Data   []T `json:"data"`
Total  int `json:"total"`
Limit  int `json:"limit"`
Offset int `json:"offset"`
}
