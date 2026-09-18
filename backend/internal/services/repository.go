package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound      = errors.New("service not found")
	ErrSlugExists    = errors.New("service slug already exists")
	ErrTypeNotFound  = errors.New("service type not found")
	ErrUserAssigned  = errors.New("user already assigned to this service")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// =====================================================
// Service Types (Catalog)
// =====================================================

func (r *Repository) ListServiceTypes(ctx context.Context) ([]*VpnServiceType, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name, description, icon, category,
		       supports_split_tunnel, supports_kill_switch,
		       supports_mfa, supports_policies, supports_ztna,
		       default_routes, default_dns, default_mtu, default_keepalive,
		       display_order, is_active, created_at, updated_at
		FROM vpn_service_types
		WHERE is_active = TRUE
		ORDER BY display_order ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list service types: %w", err)
	}
	defer rows.Close()

	var types []*VpnServiceType
	for rows.Next() {
		t := &VpnServiceType{}
		if err := rows.Scan(
			&t.ID, &t.Code, &t.Name, &t.Description, &t.Icon, &t.Category,
			&t.SupportsSplitTunnel, &t.SupportsKillSwitch,
			&t.SupportsMFA, &t.SupportsPolicies, &t.SupportsZTNA,
			&t.DefaultRoutes, &t.DefaultDNS, &t.DefaultMTU, &t.DefaultKeepalive,
			&t.DisplayOrder, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, rows.Err()
}

func (r *Repository) GetServiceType(ctx context.Context, code string) (*VpnServiceType, error) {
	t := &VpnServiceType{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, name, description, icon, category,
		       supports_split_tunnel, supports_kill_switch,
		       supports_mfa, supports_policies, supports_ztna,
		       default_routes, default_dns, default_mtu, default_keepalive,
		       display_order, is_active, created_at, updated_at
		FROM vpn_service_types
		WHERE code = $1
	`, code).Scan(
		&t.ID, &t.Code, &t.Name, &t.Description, &t.Icon, &t.Category,
		&t.SupportsSplitTunnel, &t.SupportsKillSwitch,
		&t.SupportsMFA, &t.SupportsPolicies, &t.SupportsZTNA,
		&t.DefaultRoutes, &t.DefaultDNS, &t.DefaultMTU, &t.DefaultKeepalive,
		&t.DisplayOrder, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTypeNotFound
		}
		return nil, fmt.Errorf("get service type: %w", err)
	}
	return t, nil
}

// =====================================================
// Services (Instances)
// =====================================================

func (r *Repository) CreateService(ctx context.Context, req CreateServiceRequest, ownerID uuid.UUID) (*VpnService, error) {
	s := &VpnService{}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO vpn_services (
			name, slug, service_type, description,
			server_id, subnet, server_ip, dns_servers,
			max_clients, bandwidth_limit_mbps,
			custom_routes, is_public, owner_user_id
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10,
			$11, $12, $13
		)
		RETURNING id, name, slug, service_type, description,
		          server_id, subnet, server_ip, dns_servers,
		          max_clients, current_clients, bandwidth_limit_mbps,
		          custom_routes, custom_mtu, status,
		          owner_user_id, is_public, created_at, updated_at
	`,
		req.Name, req.Slug, req.ServiceType, req.Description,
		req.ServerID, req.Subnet, req.ServerIP, req.DNSServers,
		req.MaxClients, req.BandwidthLimitMbps,
		req.CustomRoutes, req.IsPublic, ownerID,
	).Scan(
		&s.ID, &s.Name, &s.Slug, &s.ServiceType, &s.Description,
		&s.ServerID, &s.Subnet, &s.ServerIP, &s.DNSServers,
		&s.MaxClients, &s.CurrentClients, &s.BandwidthLimitMbps,
		&s.CustomRoutes, &s.CustomMTU, &s.Status,
		&s.OwnerUserID, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrSlugExists
		}
		return nil, fmt.Errorf("create service: %w", err)
	}
	return s, nil
}

func (r *Repository) GetServiceByID(ctx context.Context, id uuid.UUID) (*VpnService, error) {
	s := &VpnService{}
	err := r.pool.QueryRow(ctx, `
		SELECT s.id, s.name, s.slug, s.service_type, s.description,
		       s.server_id, s.subnet, s.server_ip, s.dns_servers,
		       s.max_clients, s.current_clients, s.bandwidth_limit_mbps,
		       s.custom_routes, s.custom_mtu, s.status,
		       s.owner_user_id, s.is_public, s.created_at, s.updated_at,
		       COALESCE(t.name, '') as type_name,
		       COALESCE(t.icon, '') as type_icon
		FROM vpn_services s
		LEFT JOIN vpn_service_types t ON s.service_type = t.code
		WHERE s.id = $1
	`, id).Scan(
		&s.ID, &s.Name, &s.Slug, &s.ServiceType, &s.Description,
		&s.ServerID, &s.Subnet, &s.ServerIP, &s.DNSServers,
		&s.MaxClients, &s.CurrentClients, &s.BandwidthLimitMbps,
		&s.CustomRoutes, &s.CustomMTU, &s.Status,
		&s.OwnerUserID, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt,
		&s.TypeName, &s.TypeIcon,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get service: %w", err)
	}
	return s, nil
}

func (r *Repository) ListServices(ctx context.Context) ([]*VpnService, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.name, s.slug, s.service_type, s.description,
		       s.server_id, s.subnet, s.server_ip, s.dns_servers,
		       s.max_clients, s.current_clients, s.bandwidth_limit_mbps,
		       s.custom_routes, s.custom_mtu, s.status,
		       s.owner_user_id, s.is_public, s.created_at, s.updated_at,
		       COALESCE(t.name, '') as type_name,
		       COALESCE(t.icon, '') as type_icon
		FROM vpn_services s
		LEFT JOIN vpn_service_types t ON s.service_type = t.code
		ORDER BY s.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	defer rows.Close()

	var services []*VpnService
	for rows.Next() {
		s := &VpnService{}
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Slug, &s.ServiceType, &s.Description,
			&s.ServerID, &s.Subnet, &s.ServerIP, &s.DNSServers,
			&s.MaxClients, &s.CurrentClients, &s.BandwidthLimitMbps,
			&s.CustomRoutes, &s.CustomMTU, &s.Status,
			&s.OwnerUserID, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt,
			&s.TypeName, &s.TypeIcon,
		); err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, rows.Err()
}

func (r *Repository) UpdateService(ctx context.Context, id uuid.UUID, req UpdateServiceRequest) (*VpnService, error) {
	s := &VpnService{}
	err := r.pool.QueryRow(ctx, `
		UPDATE vpn_services SET
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			subnet = COALESCE($4, subnet),
			server_ip = COALESCE($5, server_ip),
			dns_servers = COALESCE($6, dns_servers),
			max_clients = COALESCE($7, max_clients),
			status = COALESCE($8, status),
			is_public = COALESCE($9, is_public),
			custom_routes = COALESCE($10, custom_routes),
			updated_at = now()
		WHERE id = $1
		RETURNING id, name, slug, service_type, description,
		          server_id, subnet, server_ip, dns_servers,
		          max_clients, current_clients, bandwidth_limit_mbps,
		          custom_routes, custom_mtu, status,
		          owner_user_id, is_public, created_at, updated_at
	`,
		id, req.Name, req.Description, req.Subnet, req.ServerIP,
		req.DNSServers, req.MaxClients, req.Status, req.IsPublic, req.CustomRoutes,
	).Scan(
		&s.ID, &s.Name, &s.Slug, &s.ServiceType, &s.Description,
		&s.ServerID, &s.Subnet, &s.ServerIP, &s.DNSServers,
		&s.MaxClients, &s.CurrentClients, &s.BandwidthLimitMbps,
		&s.CustomRoutes, &s.CustomMTU, &s.Status,
		&s.OwnerUserID, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update service: %w", err)
	}
	return s, nil
}

func (r *Repository) DeleteService(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM vpn_services WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// =====================================================
// User Assignments
// =====================================================

func (r *Repository) AssignUser(ctx context.Context, serviceID uuid.UUID, req AssignUserRequest, grantedBy uuid.UUID) (*UserVpnService, error) {
	u := &UserVpnService{}
	role := req.Role
	if role == "" {
		role = UserRoleUser
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO user_vpn_services (user_id, service_id, role, granted_by, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, service_id, role, granted_by, granted_at, expires_at
	`, req.UserID, serviceID, role, grantedBy, req.ExpiresAt).Scan(
		&u.ID, &u.UserID, &u.ServiceID, &u.Role, &u.GrantedBy, &u.GrantedAt, &u.ExpiresAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUserAssigned
		}
		return nil, fmt.Errorf("assign user: %w", err)
	}
	return u, nil
}

func (r *Repository) UnassignUser(ctx context.Context, serviceID, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM user_vpn_services
		WHERE service_id = $1 AND user_id = $2
	`, serviceID, userID)
	if err != nil {
		return fmt.Errorf("unassign user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListServiceUsers(ctx context.Context, serviceID uuid.UUID) ([]*UserVpnService, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, service_id, role, granted_by, granted_at, expires_at
		FROM user_vpn_services
		WHERE service_id = $1
		ORDER BY granted_at DESC
	`, serviceID)
	if err != nil {
		return nil, fmt.Errorf("list service users: %w", err)
	}
	defer rows.Close()

	var users []*UserVpnService
	for rows.Next() {
		u := &UserVpnService{}
		if err := rows.Scan(&u.ID, &u.UserID, &u.ServiceID, &u.Role, &u.GrantedBy, &u.GrantedAt, &u.ExpiresAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// ListUserServices returns all services a user has access to
func (r *Repository) ListUserServices(ctx context.Context, userID uuid.UUID) ([]*VpnService, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.name, s.slug, s.service_type, s.description,
		       s.server_id, s.subnet, s.server_ip, s.dns_servers,
		       s.max_clients, s.current_clients, s.bandwidth_limit_mbps,
		       s.custom_routes, s.custom_mtu, s.status,
		       s.owner_user_id, s.is_public, s.created_at, s.updated_at,
		       COALESCE(t.name, '') as type_name,
		       COALESCE(t.icon, '') as type_icon
		FROM vpn_services s
		JOIN user_vpn_services us ON s.id = us.service_id
		LEFT JOIN vpn_service_types t ON s.service_type = t.code
		WHERE us.user_id = $1
		  AND s.status = 'active'
		  AND (us.expires_at IS NULL OR us.expires_at > now())
		ORDER BY t.display_order ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user services: %w", err)
	}
	defer rows.Close()

	var services []*VpnService
	for rows.Next() {
		s := &VpnService{}
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Slug, &s.ServiceType, &s.Description,
			&s.ServerID, &s.Subnet, &s.ServerIP, &s.DNSServers,
			&s.MaxClients, &s.CurrentClients, &s.BandwidthLimitMbps,
			&s.CustomRoutes, &s.CustomMTU, &s.Status,
			&s.OwnerUserID, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt,
			&s.TypeName, &s.TypeIcon,
		); err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, rows.Err()
}

// =====================================================
// Audit
// =====================================================

func (r *Repository) LogAudit(ctx context.Context, serviceID *uuid.UUID, userID *uuid.UUID, action string, details []byte, ip string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO vpn_service_audit (service_id, user_id, action, details, ip_address)
		VALUES ($1, $2, $3, $4, $5)
	`, serviceID, userID, action, details, ip)
	return err
}

// =====================================================
// Helpers
// =====================================================

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "duplicate key") || contains(msg, "unique constraint")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// unused import guard
var _ = time.Now
