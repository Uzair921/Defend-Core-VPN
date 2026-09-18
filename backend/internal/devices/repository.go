package devices

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound   = errors.New("device not found")
	ErrKeyExists  = errors.New("public key already registered")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, name, publicKey, platform string) (*Device, error) {
	d := &Device{}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO devices (user_id, name, public_key, platform)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, public_key, platform, last_seen, status, created_at
	`, userID, name, publicKey, platform).Scan(
		&d.ID, &d.UserID, &d.Name, &d.PublicKey, &d.Platform, &d.LastSeen, &d.Status, &d.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrKeyExists
		}
		return nil, fmt.Errorf("create device: %w", err)
	}
	return d, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*Device, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, public_key, platform, last_seen, status, created_at
		FROM devices WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var out []*Device
	for rows.Next() {
		d := &Device{}
		if err := rows.Scan(&d.ID, &d.UserID, &d.Name, &d.PublicKey, &d.Platform, &d.LastSeen, &d.Status, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id, userID uuid.UUID) (*Device, error) {
	d := &Device{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, public_key, platform, last_seen, status, created_at
		FROM devices WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(
		&d.ID, &d.UserID, &d.Name, &d.PublicKey, &d.Platform, &d.LastSeen, &d.Status, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get device: %w", err)
	}
	return d, nil
}

func (r *Repository) Update(ctx context.Context, id, userID uuid.UUID, name, status *string) (*Device, error) {
	d := &Device{}
	err := r.pool.QueryRow(ctx, `
		UPDATE devices SET
			name   = COALESCE($3, name),
			status = COALESCE($4, status)
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, public_key, platform, last_seen, status, created_at
	`, id, userID, name, status).Scan(
		&d.ID, &d.UserID, &d.Name, &d.PublicKey, &d.Platform, &d.LastSeen, &d.Status, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update device: %w", err)
	}
	return d, nil
}

func (r *Repository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM devices WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete device: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) TouchLastSeen(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE devices SET last_seen = now() WHERE id = $1`, id)
	return err
}

func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate key") || contains(err.Error(), "unique constraint"))
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
