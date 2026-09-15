package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrRefreshTokenInvalid = errors.New("refresh token invalid")

type RefreshRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshRepository(pool *pgxpool.Pool) *RefreshRepository {
	return &RefreshRepository{pool: pool}
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (r *RefreshRepository) Store(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, hashToken(token), expiresAt)
	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

func (r *RefreshRepository) Consume(ctx context.Context, token string) (uuid.UUID, error) {
	var userID uuid.UUID
	var expiresAt time.Time
	var revoked bool

	err := r.pool.QueryRow(ctx, `
		SELECT user_id, expires_at, revoked
		FROM refresh_tokens WHERE token_hash = $1
	`, hashToken(token)).Scan(&userID, &expiresAt, &revoked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrRefreshTokenInvalid
		}
		return uuid.Nil, fmt.Errorf("lookup refresh token: %w", err)
	}

	if revoked || time.Now().After(expiresAt) {
		return uuid.Nil, ErrRefreshTokenInvalid
	}

	// Revoke (rotation)
	_, err = r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1
	`, hashToken(token))
	if err != nil {
		return uuid.Nil, fmt.Errorf("revoke refresh token: %w", err)
	}

	return userID, nil
}

func (r *RefreshRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1 AND revoked = FALSE
	`, userID)
	return err
}
