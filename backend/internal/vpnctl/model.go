package vpnctl

import (
	"time"

	"github.com/google/uuid"
)

// VpnServer represents a VPN gateway node.
type VpnServer struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	PublicIP  string    `json:"public_ip"`
	Region    string    `json:"region,omitempty"`
	Status    string    `json:"status"` // online, offline, degraded
	Version   string    `json:"version,omitempty"`
	LastSeen  time.Time `json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
}

// Session tracks an active VPN client session.
type Session struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	DeviceID  *uuid.UUID `json:"device_id,omitempty"`
	ServerID  *uuid.UUID `json:"server_id,omitempty"`
	ClientIP  string     `json:"client_ip"`
	AssignedIP string    `json:"assigned_ip"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	BytesIn   int64      `json:"bytes_in"`
	BytesOut  int64      `json:"bytes_out"`
	Status    string     `json:"status"` // active, ended
}

// Registration request from VPN server
type RegisterServerRequest struct {
	Name     string `json:"name" validate:"required"`
	PublicIP string `json:"public_ip" validate:"required"`
	Region   string `json:"region,omitempty"`
	Version  string `json:"version,omitempty"`
}

// Heartbeat
type HeartbeatRequest struct {
	ServerID uuid.UUID `json:"server_id" validate:"required"`
	Status   string    `json:"status" validate:"required,oneof=online offline degraded"`
}

// Session start
type StartSessionRequest struct {
	UserID     uuid.UUID  `json:"user_id" validate:"required"`
	DeviceID   *uuid.UUID `json:"device_id,omitempty"`
	ServerID   uuid.UUID  `json:"server_id" validate:"required"`
	ClientIP   string     `json:"client_ip" validate:"required"`
	AssignedIP string     `json:"assigned_ip" validate:"required"`
}

// Session update
type UpdateSessionRequest struct {
	BytesIn  int64 `json:"bytes_in"`
	BytesOut int64 `json:"bytes_out"`
}

// Session end
type EndSessionRequest struct {
	BytesIn  int64 `json:"bytes_in"`
	BytesOut int64 `json:"bytes_out"`
}
