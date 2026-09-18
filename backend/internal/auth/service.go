package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"defendcore-vpn/internal/users"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountInactive    = errors.New("account inactive")
)

type Service struct {
	users    *users.Repository
	refresh  *RefreshRepository
	jwt      *JWTService
}

func NewService(users *users.Repository, refresh *RefreshRepository, jwt *JWTService) *Service {
	return &Service{users: users, refresh: refresh, jwt: jwt}
}

type AuthResult struct {
	User         *users.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

func (s *Service) Register(ctx context.Context, email, password string) (*AuthResult, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user, err := s.users.Create(ctx, email, hash, "user")
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Status != "active" {
		return nil, ErrAccountInactive
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	userID, err := s.refresh.Consume(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.Status != "active" {
		return nil, ErrAccountInactive
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.refresh.RevokeAllForUser(ctx, userID)
}

func (s *Service) issueTokens(ctx context.Context, user *users.User) (*AuthResult, error) {
	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(s.jwt.RefreshTTL())
	if err := s.refresh.Store(ctx, user.ID, refreshToken, expiresAt); err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwt.AccessTTL().Seconds()),
	}, nil
}
