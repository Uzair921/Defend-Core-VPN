package devices

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Name      string     `json:"name"`
	PublicKey string     `json:"public_key"`
	Platform  string     `json:"platform"`
	LastSeen  *time.Time `json:"last_seen,omitempty"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=64"`
	PublicKey string `json:"public_key" validate:"required"`
	Platform  string `json:"platform" validate:"required,oneof=linux windows macos android ios"`
}

type UpdateRequest struct {
	Name   *string `json:"name,omitempty" validate:"omitempty,min=1,max=64"`
	Status *string `json:"status,omitempty" validate:"omitempty,oneof=active disabled revoked"`
}

type CreateResponse struct {
	Device     *Device `json:"device"`
	PrivateKey string  `json:"private_key"`
	PublicKey  string  `json:"public_key"`
	Config     string  `json:"config"`
}
