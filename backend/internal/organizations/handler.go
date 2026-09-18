package organizations

import (
"encoding/json"
"errors"
"net/http"
"strconv"

"github.com/go-chi/chi/v5"
"github.com/go-playground/validator/v10"
"github.com/google/uuid"

"defendcore-vpn/internal/apperr"
"defendcore-vpn/internal/httputil"
"defendcore-vpn/internal/logger"
"defendcore-vpn/internal/middleware"
)

// Handler exposes the organizations domain over HTTP.
type Handler struct {
svc      *Service
validate *validator.Validate
}

// NewHandler constructs a Handler with the given service.
func NewHandler(svc *Service) *Handler {
return &Handler{
svc:      svc,
validate: validator.New(),
}
}

// =====================================================
// Organization endpoints (SuperAdmin)
// =====================================================

// CreateOrganization handles POST /api/v1/superadmin/organizations.
func (h *Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
actorID, ok := middleware.UserIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
return
}

var req CreateOrganizationRequest
if !h.decode(w, r, &req) {
return
}

org, err := h.svc.Create(r.Context(), req, actorID)
if err != nil {
h.writeError(w, r, err)
return
}

httputil.JSON(w, http.StatusCreated, org)
}

// ListOrganizations handles GET /api/v1/superadmin/organizations.
func (h *Handler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
limit := parseIntDefault(r.URL.Query().Get("limit"), 20)
offset := parseIntDefault(r.URL.Query().Get("offset"), 0)
search := r.URL.Query().Get("search")

filter := ListOrganizationsFilter{
Search: search,
Limit:  limit,
Offset: offset,
}

if s := r.URL.Query().Get("status"); s != "" {
st := Status(s)
filter.Status = &st
}
if p := r.URL.Query().Get("plan"); p != "" {
pl := Plan(p)
filter.Plan = &pl
}

page, err := h.svc.List(r.Context(), filter)
if err != nil {
h.writeError(w, r, err)
return
}

httputil.JSON(w, http.StatusOK, page)
}

// GetOrganization handles GET /api/v1/superadmin/organizations/{id}.
func (h *Handler) GetOrganization(w http.ResponseWriter, r *http.Request) {
id, err := h.parseUUIDParam(r, "id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid organization ID")
return
}

org, err := h.svc.Get(r.Context(), id)
if err != nil {
h.writeError(w, r, err)
return
}

httputil.JSON(w, http.StatusOK, org)
}

// UpdateOrganization handles PATCH /api/v1/superadmin/organizations/{id}.
func (h *Handler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
actorID, ok := middleware.UserIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
return
}

id, err := h.parseUUIDParam(r, "id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid organization ID")
return
}

var req UpdateOrganizationRequest
if !h.decode(w, r, &req) {
return
}

org, err := h.svc.Update(r.Context(), id, req, actorID)
if err != nil {
h.writeError(w, r, err)
return
}

httputil.JSON(w, http.StatusOK, org)
}

// DeleteOrganization handles DELETE /api/v1/superadmin/organizations/{id}.
func (h *Handler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
actorID, ok := middleware.UserIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
return
}

id, err := h.parseUUIDParam(r, "id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid organization ID")
return
}

if err := h.svc.Delete(r.Context(), id, actorID); err != nil {
h.writeError(w, r, err)
return
}

w.WriteHeader(http.StatusNoContent)
}

// =====================================================
// Organization users
// =====================================================

// AddOrganizationUser handles POST /api/v1/superadmin/organizations/{id}/users.
func (h *Handler) AddOrganizationUser(w http.ResponseWriter, r *http.Request) {
actorID, ok := middleware.UserIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
return
}

orgID, err := h.parseUUIDParam(r, "id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid organization ID")
return
}

var req AddOrganizationUserRequest
if !h.decode(w, r, &req) {
return
}

member, err := h.svc.AddUser(r.Context(), orgID, req, actorID)
if err != nil {
h.writeError(w, r, err)
return
}

httputil.JSON(w, http.StatusCreated, member)
}

// ListOrganizationUsers handles GET /api/v1/superadmin/organizations/{id}/users.
func (h *Handler) ListOrganizationUsers(w http.ResponseWriter, r *http.Request) {
orgID, err := h.parseUUIDParam(r, "id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid organization ID")
return
}

users, err := h.svc.ListUsers(r.Context(), orgID)
if err != nil {
h.writeError(w, r, err)
return
}
if users == nil {
users = []OrganizationUser{}
}

httputil.JSON(w, http.StatusOK, map[string]any{"users": users})
}

// RemoveOrganizationUser handles DELETE /api/v1/superadmin/organizations/{id}/users/{user_id}.
func (h *Handler) RemoveOrganizationUser(w http.ResponseWriter, r *http.Request) {
actorID, ok := middleware.UserIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
return
}

orgID, err := h.parseUUIDParam(r, "id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid organization ID")
return
}

userID, err := h.parseUUIDParam(r, "user_id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
return
}

if err := h.svc.RemoveUser(r.Context(), orgID, userID, actorID); err != nil {
h.writeError(w, r, err)
return
}

w.WriteHeader(http.StatusNoContent)
}

// =====================================================
// Subscriptions
// =====================================================

// CreateSubscription handles POST /api/v1/superadmin/organizations/{id}/subscriptions.
func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
actorID, ok := middleware.UserIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
return
}

orgID, err := h.parseUUIDParam(r, "id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid organization ID")
return
}

var req CreateSubscriptionRequest
if !h.decode(w, r, &req) {
return
}

sub, err := h.svc.CreateSubscription(r.Context(), orgID, req, actorID)
if err != nil {
h.writeError(w, r, err)
return
}

httputil.JSON(w, http.StatusCreated, sub)
}

// ListSubscriptions handles GET /api/v1/superadmin/organizations/{id}/subscriptions.
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
orgID, err := h.parseUUIDParam(r, "id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid organization ID")
return
}

subs, err := h.svc.ListSubscriptions(r.Context(), orgID)
if err != nil {
h.writeError(w, r, err)
return
}
if subs == nil {
subs = []Subscription{}
}

httputil.JSON(w, http.StatusOK, map[string]any{"subscriptions": subs})
}

// UpdateSubscriptionStatus handles PATCH /api/v1/superadmin/subscriptions/{id}.
func (h *Handler) UpdateSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
actorID, ok := middleware.UserIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
return
}

subID, err := h.parseUUIDParam(r, "id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid subscription ID")
return
}

var payload struct {
Status SubscriptionStatus `json:"status" validate:"required,oneof=active suspended cancelled"`
}
if !h.decode(w, r, &payload) {
return
}

sub, err := h.svc.UpdateSubscriptionStatus(r.Context(), subID, payload.Status, actorID)
if err != nil {
h.writeError(w, r, err)
return
}

httputil.JSON(w, http.StatusOK, sub)
}

// DeleteSubscription handles DELETE /api/v1/superadmin/subscriptions/{id}.
func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
actorID, ok := middleware.UserIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
return
}

subID, err := h.parseUUIDParam(r, "id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid subscription ID")
return
}

if err := h.svc.DeleteSubscription(r.Context(), subID, actorID); err != nil {
h.writeError(w, r, err)
return
}

w.WriteHeader(http.StatusNoContent)
}

// =====================================================
// Invoices
// =====================================================

// CreateInvoice handles POST /api/v1/superadmin/organizations/{id}/invoices.
func (h *Handler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
actorID, ok := middleware.UserIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
return
}

orgID, err := h.parseUUIDParam(r, "id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid organization ID")
return
}

var inv Invoice
if !h.decode(w, r, &inv) {
return
}
inv.OrganizationID = orgID

if err := h.svc.CreateInvoice(r.Context(), &inv, actorID); err != nil {
h.writeError(w, r, err)
return
}

httputil.JSON(w, http.StatusCreated, inv)
}

// ListInvoices handles GET /api/v1/superadmin/organizations/{id}/invoices.
func (h *Handler) ListInvoices(w http.ResponseWriter, r *http.Request) {
orgID, err := h.parseUUIDParam(r, "id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid organization ID")
return
}

limit := parseIntDefault(r.URL.Query().Get("limit"), 20)
offset := parseIntDefault(r.URL.Query().Get("offset"), 0)

invoices, err := h.svc.ListInvoices(r.Context(), orgID, limit, offset)
if err != nil {
h.writeError(w, r, err)
return
}
if invoices == nil {
invoices = []Invoice{}
}

httputil.JSON(w, http.StatusOK, map[string]any{"invoices": invoices})
}

// MarkInvoicePaid handles POST /api/v1/superadmin/invoices/{id}/paid.
func (h *Handler) MarkInvoicePaid(w http.ResponseWriter, r *http.Request) {
actorID, ok := middleware.UserIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
return
}

invoiceID, err := h.parseUUIDParam(r, "id")
if err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid invoice ID")
return
}

inv, err := h.svc.MarkInvoicePaid(r.Context(), invoiceID, actorID)
if err != nil {
h.writeError(w, r, err)
return
}

httputil.JSON(w, http.StatusOK, inv)
}

// =====================================================
// Self-service (Org admin)
// =====================================================

// GetMyOrganization handles GET /api/v1/org/me.
func (h *Handler) GetMyOrganization(w http.ResponseWriter, r *http.Request) {
orgID, ok := middleware.OrgIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusForbidden, "no_organization", "User is not a member of any organization")
return
}

org, err := h.svc.Get(r.Context(), orgID)
if err != nil {
h.writeError(w, r, err)
return
}

httputil.JSON(w, http.StatusOK, org)
}

// ListMyUsers handles GET /api/v1/org/me/users.
func (h *Handler) ListMyUsers(w http.ResponseWriter, r *http.Request) {
orgID, ok := middleware.OrgIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusForbidden, "no_organization", "User is not a member of any organization")
return
}

users, err := h.svc.ListUsers(r.Context(), orgID)
if err != nil {
h.writeError(w, r, err)
return
}

httputil.JSON(w, http.StatusOK, map[string]any{"users": users})
}

// ListMySubscriptions handles GET /api/v1/org/me/subscriptions.
func (h *Handler) ListMySubscriptions(w http.ResponseWriter, r *http.Request) {
orgID, ok := middleware.OrgIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusForbidden, "no_organization", "User is not a member of any organization")
return
}

subs, err := h.svc.ListSubscriptions(r.Context(), orgID)
if err != nil {
h.writeError(w, r, err)
return
}

httputil.JSON(w, http.StatusOK, map[string]any{"subscriptions": subs})
}

// ListMyInvoices handles GET /api/v1/org/me/invoices.
func (h *Handler) ListMyInvoices(w http.ResponseWriter, r *http.Request) {
orgID, ok := middleware.OrgIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusForbidden, "no_organization", "User is not a member of any organization")
return
}

limit := parseIntDefault(r.URL.Query().Get("limit"), 20)
offset := parseIntDefault(r.URL.Query().Get("offset"), 0)

invoices, err := h.svc.ListInvoices(r.Context(), orgID, limit, offset)
if err != nil {
h.writeError(w, r, err)
return
}

httputil.JSON(w, http.StatusOK, map[string]any{"invoices": invoices})
}

// =====================================================
// Helpers
// =====================================================

// decode reads a JSON request body, validates it, and writes an error response
// on failure. It returns true if decoding and validation succeeded.
func (h *Handler) decode(w http.ResponseWriter, r *http.Request, dst any) bool {
if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
httputil.Error(w, http.StatusBadRequest, "invalid_json", "Request body is not valid JSON")
return false
}

if err := h.validate.Struct(dst); err != nil {
var ve validator.ValidationErrors
if errors.As(err, &ve) {
fields := make(map[string]string, len(ve))
for _, fe := range ve {
fields[fe.Field()] = fe.Tag()
}
httputil.ValidationError(w, fields)
return false
}
httputil.Error(w, http.StatusBadRequest, "validation_error", "Request failed validation")
return false
}

return true
}

// writeError translates a domain error into an HTTP response.
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
var appErr *apperr.Error
if errors.As(err, &appErr) {
httputil.Error(w, appErr.HTTPStatus(), appErr.Code, appErr.Message)
return
}

logger.Error(r.Context(), "unhandled error",
"error", err.Error(),
"path", r.URL.Path,
)
httputil.Error(w, http.StatusInternalServerError, "internal_error", "An internal error occurred")
}

// parseUUIDParam extracts a UUID from a chi URL parameter.
func (h *Handler) parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
raw := chi.URLParam(r, name)
if raw == "" {
return uuid.Nil, errors.New("missing parameter")
}
return uuid.Parse(raw)
}

func parseIntDefault(s string, def int) int {
if s == "" {
return def
}
n, err := strconv.Atoi(s)
if err != nil {
return def
}
return n
}
