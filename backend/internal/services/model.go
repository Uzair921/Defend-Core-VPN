package services

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ServiceCategory string

const (
	CategoryPersonal   ServiceCategory = "personal"
	CategoryCorporate  ServiceCategory = "corporate"
	CategoryEnterprise ServiceCategory = "enterprise"
)

type VpnServiceType struct {
	ID                  uuid.UUID       `json:"id"`
	Code                string          `json:"code"`
	Name                string          `json:"name"`
	Description         string          `json:"description,omitempty"`
	Icon                string          `json:"icon,omitempty"`
	Category            ServiceCategory `json:"category"`
	SupportsSplitTunnel bool            `json:"supports_split_tunnel"`
	SupportsKillSwitch  bool            `json:"supports_kill_switch"`
	SupportsMFA         bool            `json:"supports_mfa"`
	SupportsPolicies    bool            `json:"supports_policies"`
	SupportsZTNA        bool            `json:"supports_ztna"`
	DefaultRoutes       json.RawMessage `json:"default_routes"`
	DefaultDNS          []string        `json:"default_dns"`
	DefaultMTU          int             `json:"default_mtu"`
	DefaultKeepalive    int             `json:"default_keepalive"`
	DisplayOrder        int             `json:"display_order"`
	IsActive            bool            `json:"is_active"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type ServiceStatus string

const (
	StatusActive      ServiceStatus = "active"
	StatusDisabled    ServiceStatus = "disabled"
	StatusMaintenance ServiceStatus = "maintenance"
)

type VpnService struct {
	ID                 uuid.UUID       `json:"id"`
	Name               string          `json:"name"`
	Slug               string          `json:"slug"`
	ServiceType        string          `json:"service_type"`
	Description        string          `json:"description,omitempty"`
	ServerID           *uuid.UUID      `json:"server_id,omitempty"`
	Subnet             string          `json:"subnet"`
	ServerIP           string          `json:"server_ip"`
	DNSServers         []string        `json:"dns_servers"`
	MaxClients         int             `json:"max_clients"`
	CurrentClients     int             `json:"current_clients"`
	BandwidthLimitMbps *int            `json:"bandwidth_limit_mbps,omitempty"`
	CustomRoutes       json.RawMessage `json:"custom_routes,omitempty"`
	CustomMTU          *int            `json:"custom_mtu,omitempty"`
	Status             ServiceStatus   `json:"status"`
	OwnerUserID        *uuid.UUID      `json:"owner_user_id,omitempty"`
	IsPublic           bool            `json:"is_public"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	TypeName           string          `json:"type_name,omitempty"`
	TypeIcon           string          `json:"type_icon,omitempty"`
}

type UserServiceRole string

const (
	UserRoleUser  UserServiceRole = "user"
	UserRoleAdmin UserServiceRole = "admin"
)

type UserVpnService struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	ServiceID uuid.UUID       `json:"service_id"`
	Role      UserServiceRole `json:"role"`
	GrantedBy *uuid.UUID      `json:"granted_by,omitempty"`
	GrantedAt time.Time       `json:"granted_at"`
	ExpiresAt *time.Time      `json:"expires_at,omitempty"`
}

type DeviceVpnConfig struct {
	ID              uuid.UUID       `json:"id"`
	DeviceID        uuid.UUID       `json:"device_id"`
	ServiceID       uuid.UUID       `json:"service_id"`
	AssignedIP      string          `json:"assigned_ip"`
	CustomRoutes    json.RawMessage `json:"custom_routes,omitempty"`
	CustomDNS       []string        `json:"custom_dns,omitempty"`
	IsDefault       bool            `json:"is_default"`
	LastConnectedAt *time.Time      `json:"last_connected_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type ServiceAudit struct {
	ID        uuid.UUID       `json:"id"`
	ServiceID *uuid.UUID      `json:"service_id,omitempty"`
	UserID    *uuid.UUID      `json:"user_id,omitempty"`
	Action    string          `json:"action"`
	Details   json.RawMessage `json:"details,omitempty"`
	IPAddress string          `json:"ip_address,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type CreateServiceRequest struct {
	Name               string          `json:"name" validate:"required,min=2,max=100"`
	Slug               string          `json:"slug" validate:"omitempty,min=2,max=100"`
	ServiceType        string          `json:"service_type" validate:"required"`
	Description        string          `json:"description,omitempty"`
	ServerID           *uuid.UUID      `json:"server_id,omitempty"`
	Subnet             string          `json:"subnet" validate:"required"`
	ServerIP           string          `json:"server_ip" validate:"required"`
	DNSServers         []string        `json:"dns_servers,omitempty"`
	MaxClients         int             `json:"max_clients"`
	BandwidthLimitMbps *int            `json:"bandwidth_limit_mbps,omitempty"`
	CustomRoutes       json.RawMessage `json:"custom_routes,omitempty"`
	IsPublic           bool            `json:"is_public"`
}

type UpdateServiceRequest struct {
	Name         *string          `json:"name,omitempty"`
	Description  *string          `json:"description,omitempty"`
	Subnet       *string          `json:"subnet,omitempty"`
	ServerIP     *string          `json:"server_ip,omitempty"`
	DNSServers   []string         `json:"dns_servers,omitempty"`
	MaxClients   *int             `json:"max_clients,omitempty"`
	Status       *ServiceStatus   `json:"status,omitempty"`
	IsPublic     *bool            `json:"is_public,omitempty"`
	CustomRoutes *json.RawMessage `json:"custom_routes,omitempty"`
}

type AssignUserRequest struct {
	UserID    uuid.UUID       `json:"user_id" validate:"required"`
	Role      UserServiceRole `json:"role" validate:"omitempty,oneof=user admin"`
	ExpiresAt *time.Time      `json:"expires_at,omitempty"`
}

type BulkAssignRequest struct {
	UserIDs   []uuid.UUID     `json:"user_ids" validate:"required,min=1"`
	Role      UserServiceRole `json:"role" validate:"omitempty,oneof=user admin"`
	ExpiresAt *time.Time      `json:"expires_at,omitempty"`
}

type ServiceWithUsers struct {
	Service VpnService       `json:"service"`
	Users   []UserVpnService `json:"users"`
}

type ServiceWithStats struct {
	Service        VpnService `json:"service"`
	ActiveSessions int        `json:"active_sessions"`
	TotalUsers     int        `json:"total_users"`
}

type ClientConfig struct {
	ServiceID       uuid.UUID       `json:"service_id"`
	ServiceName     string          `json:"service_name"`
	ServiceType     string          `json:"service_type"`
	ServiceIcon     string          `json:"service_icon,omitempty"`
	ServerHost      string          `json:"server_host"`
	ServerPort      int             `json:"server_port"`
	ServerPublicKey string          `json:"server_public_key"`
	AssignedIP      string          `json:"assigned_ip"`
	Routes          json.RawMessage `json:"routes"`
	DNSServers      []string        `json:"dns_servers"`
	MTU             int             `json:"mtu"`
}
