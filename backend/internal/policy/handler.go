package policy

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
	repo     *Repository
	validate *validator.Validate
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo, validate: validator.New()}
}

// POST /api/v1/admin/policies
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}

	var req CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}
	req.Value = strings.TrimSpace(req.Value)

	if err := h.validate.Struct(req); err != nil {
		fields := map[string]string{}
		for _, e := range err.(validator.ValidationErrors) {
			fields[strings.ToLower(e.Field())] = e.Tag()
		}
		httputil.ValidationError(w, fields)
		return
	}

	p, err := h.repo.Create(r.Context(), req, adminID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to create policy")
		return
	}
	httputil.JSON(w, http.StatusCreated, p)
}

// GET /api/v1/admin/policies/user/{user_id}
func (h *Handler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid user ID")
		return
	}

	policies, err := h.repo.ListByUser(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to list policies")
		return
	}
	if policies == nil {
		policies = []*AccessPolicy{}
	}
	httputil.JSON(w, http.StatusOK, map[string]interface{}{"policies": policies})
}

// GET /api/v1/admin/policies/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid policy ID")
		return
	}

	p, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Policy not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to get policy")
		return
	}
	httputil.JSON(w, http.StatusOK, p)
}

// PATCH /api/v1/admin/policies/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid policy ID")
		return
	}

	var req UpdatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "validation_error", "Invalid request")
		return
	}

	p, err := h.repo.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Policy not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to update policy")
		return
	}
	httputil.JSON(w, http.StatusOK, p)
}

// DELETE /api/v1/admin/policies/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid policy ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Policy not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to delete policy")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/vpn/policies/{user_id}
// Called by VPN server (API key auth)
func (h *Handler) GetEffectivePolicy(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid user ID")
		return
	}

	var deviceID *uuid.UUID
	if d := r.URL.Query().Get("device_id"); d != "" {
		parsed, err := uuid.Parse(d)
		if err == nil {
			deviceID = &parsed
		}
	}

	set, err := h.repo.GetEffectivePolicy(r.Context(), userID, deviceID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to get policy")
		return
	}
	if set.Rules == nil {
		set.Rules = []AccessPolicy{}
	}
	httputil.JSON(w, http.StatusOK, set)
}

