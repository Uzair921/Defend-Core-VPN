package users

import (
"time"

"github.com/google/uuid"
)

// User represents a platform account. A user may be a super-admin (platform
// operator) or belong to a single organization.
type User struct {
ID             uuid.UUID  `json:"id"`
Email          string     `json:"email"`
PasswordHash   string     `json:"-"`
Role           string     `json:"role"`
Status         string     `json:"status"`
IsSuperAdmin   bool       `json:"is_superadmin"`
OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
CreatedAt      time.Time  `json:"created_at"`
UpdatedAt      time.Time  `json:"updated_at"`
}

// RegisterRequest is the payload for creating a new account.
type RegisterRequest struct {
Email    string `json:"email" validate:"required,email"`
Password string `json:"password" validate:"required,min=8,max=72"`
}

// LoginRequest is the payload for authenticating an existing account.
type LoginRequest struct {
Email    string `json:"email" validate:"required,email"`
Password string `json:"password" validate:"required"`
}

// AuthResponse is returned on successful registration or login.
type AuthResponse struct {
User         *User  `json:"user"`
AccessToken  string `json:"access_token"`
RefreshToken string `json:"refresh_token"`
ExpiresIn    int64  `json:"expires_in"`
}
