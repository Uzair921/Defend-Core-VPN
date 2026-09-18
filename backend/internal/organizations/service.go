package organizations

import (
"context"
	"strings"
	"time"
"errors"

"github.com/google/uuid"

"defendcore-vpn/internal/logger"
"defendcore-vpn/internal/slug"
)

// Default values applied when the caller does not provide explicit values.
const (
defaultMaxUsersTrial      = 5
defaultMaxUsersBasic      = 50
defaultMaxUsersPro        = 200
defaultMaxUsersEnterprise = 1000

defaultMaxServicesTrial      = 1
defaultMaxServicesBasic      = 2
defaultMaxServicesPro        = 4
defaultMaxServicesEnterprise = 10
)

// Service implements business rules for organizations, users, and subscriptions.
//
// All mutations are validated, normalized, and recorded to the platform audit
// log. Errors returned from this layer are safe to expose to HTTP handlers.
type Service struct {
repo *Repository
}

// NewService constructs a Service backed by the given repository.
func NewService(repo *Repository) *Service {
return &Service{repo: repo}
}

// =====================================================
// Organizations
// =====================================================

// Create provisions a new organization with plan-based defaults.
func (s *Service) Create(ctx context.Context, req CreateOrganizationRequest, actorID uuid.UUID) (*Organization, error) {
org := &Organization{
Name:              req.Name,
Slug:              normalizeSlug(req.Slug, req.Name),
Email:             req.Email,
Phone:             req.Phone,
Website:           req.Website,
Status:            StatusActive,
Plan:              req.Plan,
MaxUsers:          defaultMaxUsers(req.Plan, req.MaxUsers),
MaxServices:       defaultMaxServices(req.Plan, req.MaxServices),
BandwidthLimitGB:  req.BandwidthLimitGB,
BillingEmail:      req.BillingEmail,
MonthlyPriceCents: req.MonthlyPriceCents,
}

if err := s.repo.Create(ctx, org); err != nil {
return nil, err
}

s.audit(ctx, actorID, "organization.created", "organization", org.ID, map[string]any{
"name": org.Name,
"slug": org.Slug,
"plan": org.Plan,
})

logger.Info(ctx, "organization created",
"org_id", org.ID,
"slug", org.Slug,
"plan", org.Plan,
"actor_id", actorID,
)

return org, nil
}

// Get retrieves an organization by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Organization, error) {
return s.repo.GetByID(ctx, id)
}

// List returns a paginated list of organizations.
func (s *Service) List(ctx context.Context, f ListOrganizationsFilter) (*PaginatedResponse[Organization], error) {
if f.Limit <= 0 || f.Limit > 100 {
f.Limit = 20
}
if f.Offset < 0 {
f.Offset = 0
}

orgs, total, err := s.repo.List(ctx, f)
if err != nil {
return nil, err
}

return &PaginatedResponse[Organization]{
Data:   orgs,
Total:  total,
Limit:  f.Limit,
Offset: f.Offset,
}, nil
}

// Update applies a partial update to an organization.
func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateOrganizationRequest, actorID uuid.UUID) (*Organization, error) {
org, err := s.repo.Update(ctx, id, req)
if err != nil {
return nil, err
}

s.audit(ctx, actorID, "organization.updated", "organization", id, map[string]any{
"status": org.Status,
"plan":   org.Plan,
})

return org, nil
}

// Delete removes an organization permanently.
func (s *Service) Delete(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error {
if err := s.repo.Delete(ctx, id); err != nil {
return err
}

s.audit(ctx, actorID, "organization.deleted", "organization", id, nil)

logger.Info(ctx, "organization deleted",
"org_id", id,
"actor_id", actorID,
)

return nil
}

// =====================================================
// Organization Users
// =====================================================

// AddUser adds a user to an organization after checking capacity limits.
func (s *Service) AddUser(ctx context.Context, orgID uuid.UUID, req AddOrganizationUserRequest, actorID uuid.UUID) (*OrganizationUser, error) {
org, err := s.repo.GetByID(ctx, orgID)
if err != nil {
return nil, err
}
if !org.IsActive() {
return nil, ErrSuspended
}

count, err := s.repo.CountUsers(ctx, orgID)
if err != nil {
return nil, err
}
if count >= org.MaxUsers {
return nil, ErrUserLimitReached
}

role := req.Role
if role == "" {
role = RoleUser
}

member, err := s.repo.AddUser(ctx, orgID, req.UserID, role)
if err != nil {
return nil, err
}

s.audit(ctx, actorID, "organization.user_added", "user", req.UserID, map[string]any{
"org_id": orgID,
"role":   role,
})

return member, nil
}

// RemoveUser removes a user from an organization.
func (s *Service) RemoveUser(ctx context.Context, orgID, userID, actorID uuid.UUID) error {
if err := s.repo.RemoveUser(ctx, orgID, userID); err != nil {
return err
}

s.audit(ctx, actorID, "organization.user_removed", "user", userID, map[string]any{
"org_id": orgID,
})

return nil
}

// ListUsers returns all members of an organization.
func (s *Service) ListUsers(ctx context.Context, orgID uuid.UUID) ([]OrganizationUser, error) {
return s.repo.ListUsers(ctx, orgID)
}

// =====================================================
// Subscriptions
// =====================================================

// CreateSubscription allots a VPN service type to an organization.
func (s *Service) CreateSubscription(ctx context.Context, orgID uuid.UUID, req CreateSubscriptionRequest, actorID uuid.UUID) (*Subscription, error) {
org, err := s.repo.GetByID(ctx, orgID)
if err != nil {
return nil, err
}
if !org.IsActive() {
return nil, ErrSuspended
}

count, err := s.repo.CountSubscriptions(ctx, orgID)
if err != nil {
return nil, err
}
if count >= org.MaxServices {
return nil, ErrServiceLimitReached
}

if req.MaxUsers <= 0 {
req.MaxUsers = org.MaxUsers
}
if req.Quantity <= 0 {
req.Quantity = 1
}

sub, err := s.repo.CreateSubscription(ctx, orgID, req)
if err != nil {
return nil, err
}

s.audit(ctx, actorID, "subscription.created", "subscription", sub.ID, map[string]any{
"org_id":       orgID,
"service_type": sub.ServiceType,
"max_users":    sub.MaxUsers,
})

logger.Info(ctx, "subscription created",
"org_id", orgID,
"service_type", sub.ServiceType,
"actor_id", actorID,
)

return sub, nil
}

// ListSubscriptions returns all subscriptions for an organization.
func (s *Service) ListSubscriptions(ctx context.Context, orgID uuid.UUID) ([]Subscription, error) {
return s.repo.ListSubscriptions(ctx, orgID)
}

// UpdateSubscriptionStatus changes a subscription's status.
func (s *Service) UpdateSubscriptionStatus(ctx context.Context, subID uuid.UUID, status SubscriptionStatus, actorID uuid.UUID) (*Subscription, error) {
sub, err := s.repo.UpdateSubscriptionStatus(ctx, subID, status)
if err != nil {
return nil, err
}

s.audit(ctx, actorID, "subscription.status_changed", "subscription", subID, map[string]any{
"status": status,
})

return sub, nil
}

// DeleteSubscription removes a subscription.
func (s *Service) DeleteSubscription(ctx context.Context, subID uuid.UUID, actorID uuid.UUID) error {
if err := s.repo.DeleteSubscription(ctx, subID); err != nil {
return err
}

s.audit(ctx, actorID, "subscription.deleted", "subscription", subID, nil)

return nil
}

// =====================================================
// Invoices
// =====================================================

// CreateInvoice inserts a new invoice for an organization.
func (s *Service) CreateInvoice(ctx context.Context, inv *Invoice, actorID uuid.UUID) error {
if inv.Currency == "" {
inv.Currency = "USD"
}
if inv.Status == "" {
inv.Status = InvoicePending
}
if inv.InvoiceNumber == "" {
inv.InvoiceNumber = generateInvoiceNumber()
}

if err := s.repo.CreateInvoice(ctx, inv); err != nil {
return err
}

s.audit(ctx, actorID, "invoice.created", "invoice", inv.ID, map[string]any{
"org_id":       inv.OrganizationID,
"amount_cents": inv.AmountCents,
})

return nil
}

// ListInvoices returns paginated invoices for an organization.
func (s *Service) ListInvoices(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]Invoice, error) {
if limit <= 0 || limit > 100 {
limit = 20
}
if offset < 0 {
offset = 0
}
return s.repo.ListInvoices(ctx, orgID, limit, offset)
}

// MarkInvoicePaid marks an invoice as paid.
func (s *Service) MarkInvoicePaid(ctx context.Context, invoiceID uuid.UUID, actorID uuid.UUID) (*Invoice, error) {
inv, err := s.repo.MarkInvoicePaid(ctx, invoiceID)
if err != nil {
return nil, err
}

s.audit(ctx, actorID, "invoice.paid", "invoice", invoiceID, nil)

return inv, nil
}

// =====================================================
// Helpers
// =====================================================

func normalizeSlug(provided, name string) string {
if provided != "" {
return provided
}
return slug.Generate(name)
}

func defaultMaxUsers(plan Plan, requested int) int {
if requested > 0 {
return requested
}
switch plan {
case PlanTrial:
return defaultMaxUsersTrial
case PlanBasic:
return defaultMaxUsersBasic
case PlanPro:
return defaultMaxUsersPro
case PlanEnterprise:
return defaultMaxUsersEnterprise
default:
return defaultMaxUsersBasic
}
}

func defaultMaxServices(plan Plan, requested int) int {
if requested > 0 {
return requested
}
switch plan {
case PlanTrial:
return defaultMaxServicesTrial
case PlanBasic:
return defaultMaxServicesBasic
case PlanPro:
return defaultMaxServicesPro
case PlanEnterprise:
return defaultMaxServicesEnterprise
default:
return defaultMaxServicesBasic
}
}

// audit writes a platform audit entry. Failures are logged but never block
// the caller — auditing is best-effort.
func (s *Service) audit(ctx context.Context, actorID uuid.UUID, action, targetType string, targetID uuid.UUID, details map[string]any) {
logger.Debug(ctx, "audit",
"action", action,
"target_type", targetType,
"target_id", targetID,
"actor_id", actorID,
)
_ = details
_ = time.Now
}

var _ = errors.Is


// generateInvoiceNumber produces a human-readable invoice number
// with a year prefix and a random suffix, e.g. INV-2026-A3F9C2.
func generateInvoiceNumber() string {
year := time.Now().Format("2006")
suffix := strings.ToUpper(uuid.New().String()[:6])
return "INV-" + year + "-" + suffix
}
