package policy

import (
	"time"

	"github.com/google/uuid"
)

type PolicyType string

const (
	PolicyTypeCIDR   PolicyType = "cidr"
	PolicyTypeDomain PolicyType = "domain"
	PolicyTypeURL    PolicyType = "url"
	PolicyTypeIP     PolicyType = "ip"
	PolicyTypePort   PolicyType = "port"
)

type PolicyAction string

const (
	PolicyActionAllow PolicyAction = "allow"
	PolicyActionDeny  PolicyAction = "deny"
)

type AccessPolicy struct {
	ID          uuid.UUID    `json:"id"`
	UserID      uuid.UUID    `json:"user_id"`
	DeviceID    *uuid.UUID   `json:"device_id,omitempty"`
	Type        PolicyType   `json:"type"`
	Value       string       `json:"value"`
	Action      PolicyAction `json:"action"`
	Priority    int          `json:"priority"`
	Description string       `json:"description,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CreatedBy   *uuid.UUID   `json:"created_by,omitempty"`
}

type CreatePolicyRequest struct {
	UserID      uuid.UUID    `json:"user_id" validate:"required"`
	DeviceID    *uuid.UUID   `json:"device_id,omitempty"`
	Type        PolicyType   `json:"type" validate:"required,oneof=cidr domain url ip port"`
	Value       string       `json:"value" validate:"required"`
	Action      PolicyAction `json:"action" validate:"required,oneof=allow deny"`
	Priority    int          `json:"priority"`
	Description string       `json:"description,omitempty"`
}

type UpdatePolicyRequest struct {
	Value       *string       `json:"value,omitempty"`
	Action      *PolicyAction `json:"action,omitempty" validate:"omitempty,oneof=allow deny"`
	Priority    *int          `json:"priority,omitempty"`
	Description *string       `json:"description,omitempty"`
}

type PolicySet struct {
	UserID   uuid.UUID      `json:"user_id"`
	DeviceID *uuid.UUID     `json:"device_id,omitempty"`
	Rules    []AccessPolicy `json:"rules"`
}
