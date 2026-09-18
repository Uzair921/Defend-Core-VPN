package organizations

import (
"context"
"errors"
"fmt"
"strings"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"
)

// =====================================================
// SQL Constants
// =====================================================

const queryOrganizationColumns = `
id, name, slug, email, phone, website, status, plan,
max_users, max_services, bandwidth_limit_gb, billing_email,
subscription_start, subscription_end, monthly_price_cents,
metadata, created_at, updated_at
`

const queryCreateOrganization = `
INSERT INTO organizations (
name, slug, email, phone, website, status, plan,
max_users, max_services, bandwidth_limit_gb, billing_email,
subscription_start, subscription_end, monthly_price_cents, metadata
) VALUES (
$1, $2, $3, $4, $5, $6, $7,
$8, $9, $10, $11,
$12, $13, $14, $15
)
RETURNING ` + queryOrganizationColumns

const queryGetOrganizationByID = `
SELECT ` + queryOrganizationColumns + `
FROM organizations
WHERE id = $1
`

const queryGetOrganizationBySlug = `
SELECT ` + queryOrganizationColumns + `
FROM organizations
WHERE slug = $1
`

const queryListOrganizations = `
SELECT ` + queryOrganizationColumns + `
FROM organizations
WHERE ($1::text IS NULL OR status = $1)
  AND ($2::text IS NULL OR plan = $2)
  AND ($3::text IS NULL OR name ILIKE '%' || $3 || '%' OR email ILIKE '%' || $3 || '%')
ORDER BY created_at DESC
LIMIT $4 OFFSET $5
`

const queryCountOrganizations = `
SELECT COUNT(*)
FROM organizations
WHERE ($1::text IS NULL OR status = $1)
  AND ($2::text IS NULL OR plan = $2)
  AND ($3::text IS NULL OR name ILIKE '%' || $3 || '%' OR email ILIKE '%' || $3 || '%')
`

const queryUpdateOrganization = `
UPDATE organizations SET
name = COALESCE($2, name),
email = COALESCE($3, email),
phone = COALESCE($4, phone),
website = COALESCE($5, website),
status = COALESCE($6, status),
plan = COALESCE($7, plan),
max_users = COALESCE($8, max_users),
max_services = COALESCE($9, max_services),
bandwidth_limit_gb = COALESCE($10, bandwidth_limit_gb),
billing_email = COALESCE($11, billing_email),
subscription_end = COALESCE($12, subscription_end),
monthly_price_cents = COALESCE($13, monthly_price_cents),
updated_at = now()
WHERE id = $1
RETURNING ` + queryOrganizationColumns

const queryDeleteOrganization = `
DELETE FROM organizations WHERE id = $1
`

// Organization users

const queryOrgUserColumns = `
id, organization_id, user_id, role, status, joined_at
`

const queryAddOrgUser = `
INSERT INTO organization_users (organization_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING ` + queryOrgUserColumns

const queryRemoveOrgUser = `
DELETE FROM organization_users
WHERE organization_id = $1 AND user_id = $2
`

const queryListOrgUsers = `
SELECT ` + queryOrgUserColumns + `
FROM organization_users
WHERE organization_id = $1
ORDER BY joined_at ASC
`

const queryCountOrgUsers = `
SELECT COUNT(*) FROM organization_users WHERE organization_id = $1
`

const queryGetOrgUserRole = `
SELECT role FROM organization_users
WHERE organization_id = $1 AND user_id = $2
`

const queryIsOrgMember = `
SELECT EXISTS(
SELECT 1 FROM organization_users
WHERE organization_id = $1 AND user_id = $2
)
`

// Subscriptions

const querySubscriptionColumns = `
id, organization_id, service_type, quantity, max_users,
status, price_cents_per_month, started_at, expires_at
`

const queryCreateSubscription = `
INSERT INTO organization_subscriptions (
organization_id, service_type, quantity, max_users,
price_cents_per_month, expires_at
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING ` + querySubscriptionColumns

const queryListSubscriptions = `
SELECT ` + querySubscriptionColumns + `
FROM organization_subscriptions
WHERE organization_id = $1
ORDER BY started_at DESC
`

const queryGetSubscription = `
SELECT ` + querySubscriptionColumns + `
FROM organization_subscriptions
WHERE id = $1
`

const queryUpdateSubscriptionStatus = `
UPDATE organization_subscriptions
SET status = $2
WHERE id = $1
RETURNING ` + querySubscriptionColumns

const queryDeleteSubscription = `
DELETE FROM organization_subscriptions WHERE id = $1
`

const queryCountSubscriptions = `
SELECT COUNT(*) FROM organization_subscriptions WHERE organization_id = $1
`

// Invoices

const queryInvoiceColumns = `
id, organization_id, invoice_number, amount_cents, currency,
status, period_start, period_end, paid_at, due_at, items, created_at
`

const queryCreateInvoice = `
INSERT INTO invoices (
organization_id, invoice_number, amount_cents, currency,
status, period_start, period_end, due_at, items
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING ` + queryInvoiceColumns

const queryListInvoices = `
SELECT ` + queryInvoiceColumns + `
FROM invoices
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`

const queryGetInvoiceByID = `
SELECT ` + queryInvoiceColumns + `
FROM invoices WHERE id = $1
`

const queryMarkInvoicePaid = `
UPDATE invoices
SET status = 'paid', paid_at = now()
WHERE id = $1
RETURNING ` + queryInvoiceColumns

// =====================================================
// Repository
// =====================================================

type Repository struct {
pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
return &Repository{pool: pool}
}

// =====================================================
// Organization CRUD
// =====================================================

// Create inserts a new organization and returns the created row.
func (r *Repository) Create(ctx context.Context, org *Organization) error {
err := r.pool.QueryRow(ctx, queryCreateOrganization,
org.Name, org.Slug, org.Email, org.Phone, org.Website,
org.Status, org.Plan, org.MaxUsers, org.MaxServices,
org.BandwidthLimitGB, org.BillingEmail,
org.SubscriptionStart, org.SubscriptionEnd,
org.MonthlyPriceCents, org.Metadata,
).Scan(
&org.ID, &org.Name, &org.Slug, &org.Email, &org.Phone, &org.Website,
&org.Status, &org.Plan, &org.MaxUsers, &org.MaxServices,
&org.BandwidthLimitGB, &org.BillingEmail,
&org.SubscriptionStart, &org.SubscriptionEnd,
&org.MonthlyPriceCents, &org.Metadata,
&org.CreatedAt, &org.UpdatedAt,
)
if err != nil {
if isUniqueViolation(err, "organizations_slug_key") {
return ErrSlugExists
}
return fmt.Errorf("create organization: %w", err)
}
return nil
}

// GetByID retrieves an organization by its ID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Organization, error) {
return scanOrganization(r.pool.QueryRow(ctx, queryGetOrganizationByID, id))
}

// GetBySlug retrieves an organization by its slug.
func (r *Repository) GetBySlug(ctx context.Context, slug string) (*Organization, error) {
return scanOrganization(r.pool.QueryRow(ctx, queryGetOrganizationBySlug, slug))
}

// List returns a paginated list of organizations matching the filter.
func (r *Repository) List(ctx context.Context, f ListOrganizationsFilter) ([]Organization, int, error) {
var status, plan, search *string
if f.Status != nil {
s := string(*f.Status)
status = &s
}
if f.Plan != nil {
p := string(*f.Plan)
plan = &p
}
if f.Search != "" {
search = &f.Search
}

rows, err := r.pool.Query(ctx, queryListOrganizations,
status, plan, search, f.Limit, f.Offset)
if err != nil {
return nil, 0, fmt.Errorf("list organizations: %w", err)
}
defer rows.Close()

orgs, err := scanOrganizationRows(rows)
if err != nil {
return nil, 0, err
}

var total int
if err := r.pool.QueryRow(ctx, queryCountOrganizations,
status, plan, search).Scan(&total); err != nil {
return nil, 0, fmt.Errorf("count organizations: %w", err)
}

return orgs, total, nil
}

// Update applies partial updates to an organization.
func (r *Repository) Update(ctx context.Context, id uuid.UUID, req UpdateOrganizationRequest) (*Organization, error) {
var name, email, phone, website *string
var status, plan *string

if req.Name != nil {
name = req.Name
}
if req.Email != nil {
email = req.Email
}
if req.Phone != nil {
phone = req.Phone
}
if req.Website != nil {
website = req.Website
}
if req.Status != nil {
s := string(*req.Status)
status = &s
}
if req.Plan != nil {
p := string(*req.Plan)
plan = &p
}

org, err := scanOrganization(r.pool.QueryRow(ctx, queryUpdateOrganization,
id, name, email, phone, website, status, plan,
req.MaxUsers, req.MaxServices, req.BandwidthLimitGB,
req.BillingEmail, req.SubscriptionEnd, req.MonthlyPriceCents,
))
if err != nil {
return nil, err
}
return org, nil
}

// Delete removes an organization permanently.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
tag, err := r.pool.Exec(ctx, queryDeleteOrganization, id)
if err != nil {
return fmt.Errorf("delete organization: %w", err)
}
if tag.RowsAffected() == 0 {
return ErrNotFound
}
return nil
}

// =====================================================
// Organization Users
// =====================================================

// AddUser adds a user as a member of an organization.
func (r *Repository) AddUser(ctx context.Context, orgID, userID uuid.UUID, role Role) (*OrganizationUser, error) {
m := &OrganizationUser{}
err := r.pool.QueryRow(ctx, queryAddOrgUser, orgID, userID, role).Scan(
&m.ID, &m.OrganizationID, &m.UserID, &m.Role, &m.Status, &m.JoinedAt,
)
if err != nil {
if isUniqueViolation(err, "") {
return nil, ErrAlreadyMember
}
return nil, fmt.Errorf("add organization user: %w", err)
}
return m, nil
}

// RemoveUser removes a user from an organization.
func (r *Repository) RemoveUser(ctx context.Context, orgID, userID uuid.UUID) error {
tag, err := r.pool.Exec(ctx, queryRemoveOrgUser, orgID, userID)
if err != nil {
return fmt.Errorf("remove organization user: %w", err)
}
if tag.RowsAffected() == 0 {
return ErrNotMember
}
return nil
}

// ListUsers returns all members of an organization.
func (r *Repository) ListUsers(ctx context.Context, orgID uuid.UUID) ([]OrganizationUser, error) {
rows, err := r.pool.Query(ctx, queryListOrgUsers, orgID)
if err != nil {
return nil, fmt.Errorf("list organization users: %w", err)
}
defer rows.Close()

var users []OrganizationUser
for rows.Next() {
var u OrganizationUser
if err := rows.Scan(
&u.ID, &u.OrganizationID, &u.UserID, &u.Role, &u.Status, &u.JoinedAt,
); err != nil {
return nil, err
}
users = append(users, u)
}
return users, rows.Err()
}

// CountUsers returns the number of users in an organization.
func (r *Repository) CountUsers(ctx context.Context, orgID uuid.UUID) (int, error) {
var count int
if err := r.pool.QueryRow(ctx, queryCountOrgUsers, orgID).Scan(&count); err != nil {
return 0, fmt.Errorf("count organization users: %w", err)
}
return count, nil
}

// GetUserRole returns a user's role within an organization.
func (r *Repository) GetUserRole(ctx context.Context, orgID, userID uuid.UUID) (Role, error) {
var role Role
err := r.pool.QueryRow(ctx, queryGetOrgUserRole, orgID, userID).Scan(&role)
if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return "", ErrNotMember
}
return "", fmt.Errorf("get user role: %w", err)
}
return role, nil
}

// IsMember reports whether a user belongs to an organization.
func (r *Repository) IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
var exists bool
if err := r.pool.QueryRow(ctx, queryIsOrgMember, orgID, userID).Scan(&exists); err != nil {
return false, fmt.Errorf("check membership: %w", err)
}
return exists, nil
}

// =====================================================
// Subscriptions
// =====================================================

// CreateSubscription allots a VPN service type to an organization.
func (r *Repository) CreateSubscription(ctx context.Context, orgID uuid.UUID, req CreateSubscriptionRequest) (*Subscription, error) {
s := &Subscription{}
err := r.pool.QueryRow(ctx, queryCreateSubscription,
orgID, req.ServiceType, req.Quantity, req.MaxUsers,
req.PriceCentsPerMonth, req.ExpiresAt,
).Scan(
&s.ID, &s.OrganizationID, &s.ServiceType, &s.Quantity, &s.MaxUsers,
&s.Status, &s.PriceCentsPerMonth, &s.StartedAt, &s.ExpiresAt,
)
if err != nil {
if isUniqueViolation(err, "") {
return nil, ErrServiceLimitReached
}
return nil, fmt.Errorf("create subscription: %w", err)
}
return s, nil
}

// ListSubscriptions returns all subscriptions for an organization.
func (r *Repository) ListSubscriptions(ctx context.Context, orgID uuid.UUID) ([]Subscription, error) {
rows, err := r.pool.Query(ctx, queryListSubscriptions, orgID)
if err != nil {
return nil, fmt.Errorf("list subscriptions: %w", err)
}
defer rows.Close()

var subs []Subscription
for rows.Next() {
var s Subscription
if err := rows.Scan(
&s.ID, &s.OrganizationID, &s.ServiceType, &s.Quantity, &s.MaxUsers,
&s.Status, &s.PriceCentsPerMonth, &s.StartedAt, &s.ExpiresAt,
); err != nil {
return nil, err
}
subs = append(subs, s)
}
return subs, rows.Err()
}

// GetSubscription retrieves a subscription by ID.
func (r *Repository) GetSubscription(ctx context.Context, id uuid.UUID) (*Subscription, error) {
s := &Subscription{}
err := r.pool.QueryRow(ctx, queryGetSubscription, id).Scan(
&s.ID, &s.OrganizationID, &s.ServiceType, &s.Quantity, &s.MaxUsers,
&s.Status, &s.PriceCentsPerMonth, &s.StartedAt, &s.ExpiresAt,
)
if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, ErrNotFound
}
return nil, fmt.Errorf("get subscription: %w", err)
}
return s, nil
}

// UpdateSubscriptionStatus changes a subscription's status.
func (r *Repository) UpdateSubscriptionStatus(ctx context.Context, id uuid.UUID, status SubscriptionStatus) (*Subscription, error) {
s := &Subscription{}
err := r.pool.QueryRow(ctx, queryUpdateSubscriptionStatus, id, status).Scan(
&s.ID, &s.OrganizationID, &s.ServiceType, &s.Quantity, &s.MaxUsers,
&s.Status, &s.PriceCentsPerMonth, &s.StartedAt, &s.ExpiresAt,
)
if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, ErrNotFound
}
return nil, fmt.Errorf("update subscription status: %w", err)
}
return s, nil
}

// DeleteSubscription removes a subscription.
func (r *Repository) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
tag, err := r.pool.Exec(ctx, queryDeleteSubscription, id)
if err != nil {
return fmt.Errorf("delete subscription: %w", err)
}
if tag.RowsAffected() == 0 {
return ErrNotFound
}
return nil
}

// CountSubscriptions returns the number of active subscriptions for an org.
func (r *Repository) CountSubscriptions(ctx context.Context, orgID uuid.UUID) (int, error) {
var count int
if err := r.pool.QueryRow(ctx, queryCountSubscriptions, orgID).Scan(&count); err != nil {
return 0, fmt.Errorf("count subscriptions: %w", err)
}
return count, nil
}

// =====================================================
// Invoices
// =====================================================

// CreateInvoice generates a new invoice for an organization.
func (r *Repository) CreateInvoice(ctx context.Context, inv *Invoice) error {
err := r.pool.QueryRow(ctx, queryCreateInvoice,
inv.OrganizationID, inv.InvoiceNumber, inv.AmountCents, inv.Currency,
inv.Status, inv.PeriodStart, inv.PeriodEnd, inv.DueAt, inv.Items,
).Scan(
&inv.ID, &inv.OrganizationID, &inv.InvoiceNumber, &inv.AmountCents,
&inv.Currency, &inv.Status, &inv.PeriodStart, &inv.PeriodEnd,
&inv.PaidAt, &inv.DueAt, &inv.Items, &inv.CreatedAt,
)
if err != nil {
return fmt.Errorf("create invoice: %w", err)
}
return nil
}

// ListInvoices returns paginated invoices for an organization.
func (r *Repository) ListInvoices(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]Invoice, error) {
rows, err := r.pool.Query(ctx, queryListInvoices, orgID, limit, offset)
if err != nil {
return nil, fmt.Errorf("list invoices: %w", err)
}
defer rows.Close()

var invoices []Invoice
for rows.Next() {
var inv Invoice
if err := rows.Scan(
&inv.ID, &inv.OrganizationID, &inv.InvoiceNumber, &inv.AmountCents,
&inv.Currency, &inv.Status, &inv.PeriodStart, &inv.PeriodEnd,
&inv.PaidAt, &inv.DueAt, &inv.Items, &inv.CreatedAt,
); err != nil {
return nil, err
}
invoices = append(invoices, inv)
}
return invoices, rows.Err()
}

// GetInvoiceByID retrieves an invoice by ID.
func (r *Repository) GetInvoiceByID(ctx context.Context, id uuid.UUID) (*Invoice, error) {
inv := &Invoice{}
err := r.pool.QueryRow(ctx, queryGetInvoiceByID, id).Scan(
&inv.ID, &inv.OrganizationID, &inv.InvoiceNumber, &inv.AmountCents,
&inv.Currency, &inv.Status, &inv.PeriodStart, &inv.PeriodEnd,
&inv.PaidAt, &inv.DueAt, &inv.Items, &inv.CreatedAt,
)
if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, ErrNotFound
}
return nil, fmt.Errorf("get invoice: %w", err)
}
return inv, nil
}

// MarkInvoicePaid marks an invoice as paid.
func (r *Repository) MarkInvoicePaid(ctx context.Context, id uuid.UUID) (*Invoice, error) {
inv := &Invoice{}
err := r.pool.QueryRow(ctx, queryMarkInvoicePaid, id).Scan(
&inv.ID, &inv.OrganizationID, &inv.InvoiceNumber, &inv.AmountCents,
&inv.Currency, &inv.Status, &inv.PeriodStart, &inv.PeriodEnd,
&inv.PaidAt, &inv.DueAt, &inv.Items, &inv.CreatedAt,
)
if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, ErrNotFound
}
return nil, fmt.Errorf("mark invoice paid: %w", err)
}
return inv, nil
}

// =====================================================
// Scan Helpers
// =====================================================

func scanOrganization(row pgx.Row) (*Organization, error) {
org := &Organization{}
err := row.Scan(
&org.ID, &org.Name, &org.Slug, &org.Email, &org.Phone, &org.Website,
&org.Status, &org.Plan, &org.MaxUsers, &org.MaxServices,
&org.BandwidthLimitGB, &org.BillingEmail,
&org.SubscriptionStart, &org.SubscriptionEnd,
&org.MonthlyPriceCents, &org.Metadata,
&org.CreatedAt, &org.UpdatedAt,
)
if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, ErrNotFound
}
return nil, fmt.Errorf("scan organization: %w", err)
}
return org, nil
}

func scanOrganizationRows(rows pgx.Rows) ([]Organization, error) {
orgs := make([]Organization, 0, 16)
for rows.Next() {
var org Organization
if err := rows.Scan(
&org.ID, &org.Name, &org.Slug, &org.Email, &org.Phone, &org.Website,
&org.Status, &org.Plan, &org.MaxUsers, &org.MaxServices,
&org.BandwidthLimitGB, &org.BillingEmail,
&org.SubscriptionStart, &org.SubscriptionEnd,
&org.MonthlyPriceCents, &org.Metadata,
&org.CreatedAt, &org.UpdatedAt,
); err != nil {
return nil, err
}
orgs = append(orgs, org)
}
if err := rows.Err(); err != nil {
return nil, err
}
return orgs, nil
}

// isUniqueViolation checks whether the error is a PostgreSQL unique violation.
// If constraint is empty, any unique violation matches.
func isUniqueViolation(err error, constraint string) bool {
if err == nil {
return false
}
msg := err.Error()
if !strings.Contains(msg, "duplicate key") && !strings.Contains(msg, "unique constraint") {
return false
}
if constraint == "" {
return true
}
return strings.Contains(msg, constraint)
}

// PrimaryOrgForUser returns the first organization the user belongs to.
// This is used by the tenant middleware to select an active organization
// when a user is a member of exactly one (or when we don't need multi-org
// selection). It returns ErrNotMember if the user is not a member of any
// organization.
func (r *Repository) PrimaryOrgForUser(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
const q = `SELECT organization_id FROM organization_users
WHERE user_id = $1
ORDER BY joined_at ASC
LIMIT 1`

var orgID uuid.UUID
err := r.pool.QueryRow(ctx, q, userID).Scan(&orgID)
if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return uuid.Nil, ErrNotMember
}
return uuid.Nil, fmt.Errorf("primary org for user: %w", err)
}
return orgID, nil
}
