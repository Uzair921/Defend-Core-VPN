package vpnctl

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
	repo     *Repository
	validate *validator.Validate
	apiKey   string
}

func NewHandler(repo *Repository, apiKey string) *Handler {
	return &Handler{
		repo:     repo,
		validate: validator.New(),
		apiKey:   apiKey,
	}
}

// ===== Server-side endpoints (called by VPN server, auth via API key) =====

func (h *Handler) RegisterServer(w http.ResponseWriter, r *http.Request) {
	if !h.checkAPIKey(r) {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "Invalid API key")
		return
	}

	var req RegisterServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "validation", err.Error())
		return
	}

	srv, err := h.repo.RegisterServer(r.Context(), req.Name, req.PublicIP, req.Region, req.Version)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to register server")
		return
	}
	httputil.JSON(w, http.StatusOK, srv)
}

func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	if !h.checkAPIKey(r) {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "Invalid API key")
		return
	}

	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "validation", err.Error())
		return
	}

	if err := h.repo.Heartbeat(r.Context(), req.ServerID, req.Status); err != nil {
		if errors.Is(err, ErrServerNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Server not registered")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Heartbeat failed")
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) StartSession(w http.ResponseWriter, r *http.Request) {
	if !h.checkAPIKey(r) {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "Invalid API key")
		return
	}

	var req StartSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "validation", err.Error())
		return
	}

	sess, err := h.repo.StartSession(r.Context(), req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to start session")
		return
	}
	httputil.JSON(w, http.StatusCreated, sess)
}

func (h *Handler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	if !h.checkAPIKey(r) {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "Invalid API key")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid session ID")
		return
	}

	var req UpdateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request")
		return
	}

	if err := h.repo.UpdateSession(r.Context(), id, req.BytesIn, req.BytesOut); err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Session not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Update failed")
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) EndSession(w http.ResponseWriter, r *http.Request) {
	if !h.checkAPIKey(r) {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized", "Invalid API key")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid session ID")
		return
	}

	var req EndSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request")
		return
	}

	if err := h.repo.EndSession(r.Context(), id, req.BytesIn, req.BytesOut); err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Session not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "End failed")
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ended"})
}

// ===== User-facing endpoints (auth via JWT) =====

func (h *Handler) ListServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.repo.ListServers(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to list servers")
		return
	}
	if servers == nil {
		servers = []*VpnServer{}
	}
	httputil.JSON(w, http.StatusOK, map[string]interface{}{"servers": servers})
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauth", "Not authenticated")
		return
	}

	sessions, err := h.repo.ListActiveSessions(r.Context(), &userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to list sessions")
		return
	}
	if sessions == nil {
		sessions = []*Session{}
	}
	httputil.JSON(w, http.StatusOK, map[string]interface{}{"sessions": sessions})
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauth", "Not authenticated")
		return
	}
	_ = userID

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_id", "Invalid session ID")
		return
	}

	sess, err := h.repo.GetSession(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			httputil.Error(w, http.StatusNotFound, "not_found", "Session not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to get session")
		return
	}
	httputil.JSON(w, http.StatusOK, sess)
}

func (h *Handler) checkAPIKey(r *http.Request) bool {
	if h.apiKey == "" {
		return true // dev mode
	}
	key := r.Header.Get("X-API-Key")
	return key == h.apiKey
}

// APIKeyAuth is a middleware that validates the X-API-Key header.
func (h *Handler) APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" || key != h.apiKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized","message":"Invalid or missing API key"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// APIKeyAuth is a middleware that validates the X-API-Key header.
