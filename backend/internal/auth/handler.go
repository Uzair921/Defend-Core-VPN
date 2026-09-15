package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"vpn.local/backend/internal/httputil"
	"vpn.local/backend/internal/middleware"
	"vpn.local/backend/internal/users"
)

type Handler struct {
	svc      *Service
	validate *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req users.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if err := h.validate.Struct(req); err != nil {
		fields := map[string]string{}
		for _, e := range err.(validator.ValidationErrors) {
			fields[strings.ToLower(e.Field())] = e.Tag()
		}
		httputil.ValidationError(w, fields)
		return
	}

	result, err := h.svc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, users.ErrEmailExists) {
			httputil.Error(w, http.StatusConflict, "email_exists", "Email already registered")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Registration failed")
		return
	}

	httputil.JSON(w, http.StatusCreated, toAuthResponse(result))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req users.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if err := h.validate.Struct(req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "validation_error", "Email and password are required")
		return
	}

	result, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httputil.Error(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
			return
		}
		if errors.Is(err, ErrAccountInactive) {
			httputil.Error(w, http.StatusForbidden, "account_inactive", "Account is not active")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Login failed")
		return
	}

	httputil.JSON(w, http.StatusOK, toAuthResponse(result))
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		httputil.Error(w, http.StatusBadRequest, "invalid_request", "refresh_token is required")
		return
	}

	result, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid_refresh_token", "Refresh token invalid or expired")
		return
	}

	httputil.JSON(w, http.StatusOK, toAuthResponse(result))
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}
	if err := h.svc.Logout(r.Context(), userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "internal_error", "Logout failed")
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
		return
	}
	// For simplicity, return claims. In a real app, fetch from DB.
	httputil.JSON(w, http.StatusOK, map[string]interface{}{
		"id": userID,
	})
}

func toAuthResponse(r *AuthResult) users.AuthResponse {
	return users.AuthResponse{
		User:         r.User,
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		ExpiresIn:    r.ExpiresIn,
	}
}
