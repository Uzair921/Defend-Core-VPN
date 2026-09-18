package devices

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"defendcore-vpn/internal/httputil"
	"defendcore-vpn/internal/middleware"
)

type Handler struct {
	svc        *Service
	validate   *validator.Validate
	serverHost string
	serverPort int
}

func NewHandler(svc *Service, serverHost string, serverPort int) *Handler {
	return &Handler{
		svc:        svc,
		validate:   validator.New(),
		serverHost: serverHost,
		serverPort: serverPort,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		fields := map[string]string{}
		for _, e := range err.(validator.ValidationErrors) {
			fields[e.Field()] = e.Tag()
		}
		httputil.ValidationError(w, fields)
		return
	}

	resp, err := h.svc.Create(r.Context(), userID, req, h.serverHost, h.serverPort)
	if err != nil {
		if errors.Is(err, ErrLimitReached) {
			httputil.Error(w, http.StatusForbidden, "limit_reached",
				"Maximum number of devices reached")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to create device")
		return
	}
	httputil.JSON(w, http.StatusCreated, resp)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}

	devices, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to list devices")
		return
	}

	if devices == nil {
		devices = []*Device{}
	}
	httputil.JSON(w, http.StatusOK, map[string]interface{}{"devices": devices})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid device ID")
		return
	}

	device, err := h.svc.Get(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Device not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to get device")
		return
	}
	httputil.JSON(w, http.StatusOK, device)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid device ID")
		return
	}

	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Device not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to delete device")
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
