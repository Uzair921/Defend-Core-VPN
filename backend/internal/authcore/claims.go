package authcore

import "github.com/google/uuid"

// Claims is the shared claims type used across auth and middleware packages.
type Claims struct {
	UserID uuid.UUID
	Email  string
	Role   string
}

// TokenValidator is implemented by auth.JWTService.
type TokenValidator interface {
	ValidateAccessToken(token string) (*Claims, error)
}
