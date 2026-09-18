package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"defendcore-vpn/internal/httputil"
	"defendcore-vpn/internal/middleware"
)

type Handler struct {
	svc      *Service
	validate *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc:      svc,
		validate: validator.New(),
	}
}

// =====================================================
// Service Types (Public)
// =====================================================

// GET /api/v1/vpn/service-types
func (h *Handler) ListServiceTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.svc.ListTypes(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to list service types")
		return
	}
	if types == nil {
		types = []*VpnServiceType{}
	}
	httputil.JSON(w, http.StatusOK, map[string]interface{}{"types": types})
}

// GET /api/v1/vpn/service-types/{code}
func (h *Handler) GetServiceType(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	t, err := h.svc.GetType(r.Context(), code)
	if err != nil {
		if errors.Is(err, ErrTypeNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Service type not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to get service type")
		return
	}
	httputil.JSON(w, http.StatusOK, t)
}

// =====================================================
// Services (Admin)
// =====================================================

// POST /api/v1/admin/vpn/services
func (h *Handler) CreateService(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}

	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)

	if err := h.validate.Struct(req); err != nil {
		fields := map[string]string{}
		for _, e := range err.(validator.ValidationErrors) {
			fields[strings.ToLower(e.Field())] = e.Tag()
		}
		httputil.ValidationError(w, fields)
		return
	}

	svc, err := h.svc.CreateService(r.Context(), req, actorID)
	if err != nil {
		if errors.Is(err, ErrSlugExists) {
			httputil.Error(w, http.StatusConflict, "slug_exists", "Service slug already exists")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, svc)
}

// GET /api/v1/admin/vpn/services
func (h *Handler) ListServices(w http.ResponseWriter, r *http.Request) {
	services, err := h.svc.ListServices(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to list services")
		return
	}
	if services == nil {
		services = []*VpnService{}
	}
	httputil.JSON(w, http.StatusOK, map[string]interface{}{"services": services})
}

// GET /api/v1/admin/vpn/services/{id}
func (h *Handler) GetService(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid service ID")
		return
	}

	svc, err := h.svc.GetService(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Service not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to get service")
		return
	}
	httputil.JSON(w, http.StatusOK, svc)
}

// PATCH /api/v1/admin/vpn/services/{id}
func (h *Handler) UpdateService(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid service ID")
		return
	}

	var req UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	svc, err := h.svc.UpdateService(r.Context(), id, req, actorID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Service not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to update service")
		return
	}
	httputil.JSON(w, http.StatusOK, svc)
}

// DELETE /api/v1/admin/vpn/services/{id}
func (h *Handler) DeleteService(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid service ID")
		return
	}

	if err := h.svc.DeleteService(r.Context(), id, actorID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Service not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to delete service")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// =====================================================
// User Assignments (Admin)
// =====================================================

// POST /api/v1/admin/vpn/services/{id}/users
func (h *Handler) AssignUser(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}

	serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid service ID")
		return
	}

	var req AssignUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "validation_error", "Invalid request")
		return
	}

	assignment, err := h.svc.AssignUser(r.Context(), serviceID, req, actorID)
	if err != nil {
		if errors.Is(err, ErrUserAssigned) {
			httputil.Error(w, http.StatusConflict, "already_assigned", "User already assigned")
			return
		}
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Service not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, assignment)
}

// POST /api/v1/admin/vpn/services/{id}/users/bulk
func (h *Handler) BulkAssignUsers(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}

	serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid service ID")
		return
	}

	var req BulkAssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "validation_error", "Invalid request")
		return
	}

	assignments, err := h.svc.BulkAssignUsers(r.Context(), serviceID, req, actorID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, map[string]interface{}{
		"assigned": len(assignments),
		"results":  assignments,
	})
}

// GET /api/v1/admin/vpn/services/{id}/users
func (h *Handler) ListServiceUsers(w http.ResponseWriter, r *http.Request) {
	serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid service ID")
		return
	}

	users, err := h.svc.ListServiceUsers(r.Context(), serviceID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to list users")
		return
	}
	if users == nil {
		users = []*UserVpnService{}
	}
	httputil.JSON(w, http.StatusOK, map[string]interface{}{"users": users})
}

// DELETE /api/v1/admin/vpn/services/{id}/users/{user_id}
func (h *Handler) UnassignUser(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}

	serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid service ID")
		return
	}
	userID, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	if err := h.svc.UnassignUser(r.Context(), serviceID, userID, actorID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Assignment not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to unassign user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// =====================================================
// Client Endpoints
// =====================================================

// GET /api/v1/vpn/my-services
func (h *Handler) ListMyServices(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}

	services, err := h.svc.ListUserServices(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to list services")
		return
	}
	if services == nil {
		services = []*VpnService{}
	}
	httputil.JSON(w, http.StatusOK, map[string]interface{}{"services": services})
}

// GET /api/v1/vpn/services/{id}/config
func (h *Handler) GetClientConfig(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}

	serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid service ID")
		return
	}

	config, err := h.svc.GetClientConfig(r.Context(), userID, serviceID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Service not found or no access")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to get config")
		return
	}
	httputil.JSON(w, http.StatusOK, config)
}
