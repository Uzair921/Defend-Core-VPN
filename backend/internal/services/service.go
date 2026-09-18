package services

import (
	"context"
	"os"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// =====================================================
// Service Types
// =====================================================

func (s *Service) ListTypes(ctx context.Context) ([]*VpnServiceType, error) {
	return s.repo.ListServiceTypes(ctx)
}

func (s *Service) GetType(ctx context.Context, code string) (*VpnServiceType, error) {
	return s.repo.GetServiceType(ctx, code)
}

// =====================================================
// Services CRUD
// =====================================================

func (s *Service) CreateService(ctx context.Context, req CreateServiceRequest, ownerID uuid.UUID) (*VpnService, error) {
	// Validate service type exists
	svcType, err := s.repo.GetServiceType(ctx, req.ServiceType)
	if err != nil {
		if errors.Is(err, ErrTypeNotFound) {
			return nil, fmt.Errorf("unknown service type: %s", req.ServiceType)
		}
		return nil, err
	}

	// Generate slug if not provided
	if req.Slug == "" {
		req.Slug = generateSlug(req.Name)
	}

	// Apply defaults
	if req.MaxClients == 0 {
		req.MaxClients = 100
	}
	if len(req.DNSServers) == 0 {
		req.DNSServers = svcType.DefaultDNS
	}
	// Use type's default routes if custom not provided
	if req.CustomRoutes == nil {
		req.CustomRoutes = svcType.DefaultRoutes
	}

	svc, err := s.repo.CreateService(ctx, req, ownerID)
	if err != nil {
		return nil, err
	}

	// Audit log
	s.repo.LogAudit(ctx, &svc.ID, &ownerID, "service_created", mustJSON(map[string]string{
		"name": svc.Name,
		"type": svc.ServiceType,
	}), "")

	return svc, nil
}

func (s *Service) GetService(ctx context.Context, id uuid.UUID) (*VpnService, error) {
	return s.repo.GetServiceByID(ctx, id)
}

func (s *Service) ListServices(ctx context.Context) ([]*VpnService, error) {
	return s.repo.ListServices(ctx)
}

func (s *Service) UpdateService(ctx context.Context, id uuid.UUID, req UpdateServiceRequest, actorID uuid.UUID) (*VpnService, error) {
	svc, err := s.repo.UpdateService(ctx, id, req)
	if err != nil {
		return nil, err
	}

	s.repo.LogAudit(ctx, &id, &actorID, "service_updated", mustJSON(map[string]string{
		"name": svc.Name,
	}), "")

	return svc, nil
}

func (s *Service) DeleteService(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error {
	// Get service first for audit
	svc, err := s.repo.GetServiceByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteService(ctx, id); err != nil {
		return err
	}

	s.repo.LogAudit(ctx, nil, &actorID, "service_deleted", mustJSON(map[string]string{
		"name": svc.Name,
		"id":   svc.ID.String(),
	}), "")

	return nil
}

// =====================================================
// User Assignments
// =====================================================

func (s *Service) AssignUser(ctx context.Context, serviceID uuid.UUID, req AssignUserRequest, actorID uuid.UUID) (*UserVpnService, error) {
	// Verify service exists
	svc, err := s.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Check capacity
	users, err := s.repo.ListServiceUsers(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	if len(users) >= svc.MaxClients {
		return nil, fmt.Errorf("service at max capacity (%d)", svc.MaxClients)
	}

	assignment, err := s.repo.AssignUser(ctx, serviceID, req, actorID)
	if err != nil {
		return nil, err
	}

	s.repo.LogAudit(ctx, &serviceID, &actorID, "user_assigned", mustJSON(map[string]string{
		"user_id": req.UserID.String(),
		"role":    string(assignment.Role),
	}), "")

	return assignment, nil
}

func (s *Service) BulkAssignUsers(ctx context.Context, serviceID uuid.UUID, req BulkAssignRequest, actorID uuid.UUID) ([]*UserVpnService, error) {
	var results []*UserVpnService
	for _, userID := range req.UserIDs {
		assignReq := AssignUserRequest{
			UserID:    userID,
			Role:      req.Role,
			ExpiresAt: req.ExpiresAt,
		}
		u, err := s.AssignUser(ctx, serviceID, assignReq, actorID)
		if err != nil {
			// Skip already assigned, continue with others
			if errors.Is(err, ErrUserAssigned) {
				continue
			}
			return results, fmt.Errorf("failed for %s: %w", userID, err)
		}
		results = append(results, u)
	}
	return results, nil
}

func (s *Service) UnassignUser(ctx context.Context, serviceID, userID, actorID uuid.UUID) error {
	if err := s.repo.UnassignUser(ctx, serviceID, userID); err != nil {
		return err
	}

	s.repo.LogAudit(ctx, &serviceID, &actorID, "user_unassigned", mustJSON(map[string]string{
		"user_id": userID.String(),
	}), "")

	return nil
}

func (s *Service) ListServiceUsers(ctx context.Context, serviceID uuid.UUID) ([]*UserVpnService, error) {
	return s.repo.ListServiceUsers(ctx, serviceID)
}

// =====================================================
// Client-facing
// =====================================================

// ListUserServices returns services available to a user
func (s *Service) ListUserServices(ctx context.Context, userID uuid.UUID) ([]*VpnService, error) {
	return s.repo.ListUserServices(ctx, userID)
}

// GetClientConfig returns full client config for a service
func (s *Service) GetClientConfig(ctx context.Context, userID, serviceID uuid.UUID) (*ClientConfig, error) {
	// Verify user has access
	userServices, err := s.repo.ListUserServices(ctx, userID)
	if err != nil {
		return nil, err
	}

	var svc *VpnService
	for _, s := range userServices {
		if s.ID == serviceID {
			svc = s
			break
		}
	}
	if svc == nil {
		return nil, ErrNotFound
	}

	// Get service type for defaults
	svcType, err := s.repo.GetServiceType(ctx, svc.ServiceType)
	if err != nil {
		return nil, err
	}

	// Build client config
	config := &ClientConfig{
		ServiceID:   svc.ID,
		ServiceName: svc.Name,
		ServiceType: svc.ServiceType,
		ServiceIcon: svcType.Icon,
		ServerHost:  svc.ServerIP,
		ServerPort:  51820, // TODO: from server config
		ServerPublicKey: os.Getenv("VPN_SERVER_PUBLIC_KEY"),
		AssignedIP:      "10.8.0.5",
		Routes:      svc.CustomRoutes,
		DNSServers:  svc.DNSServers,
		MTU:         svcType.DefaultMTU,
	}

	if config.Routes == nil {
		config.Routes = svcType.DefaultRoutes
	}

	return config, nil
}

// =====================================================
// Helpers
// =====================================================

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	
	// Remove non-alphanumeric except dash
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	
	// Append short random suffix to ensure uniqueness
	suffix := uuid.New().String()[:6]
	return fmt.Sprintf("%s-%s", result.String(), suffix)
}

func mustJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
