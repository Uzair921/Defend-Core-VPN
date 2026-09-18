package policy

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("policy not found")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, req CreatePolicyRequest, createdBy uuid.UUID) (*AccessPolicy, error) {
	p := &AccessPolicy{}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO access_policies (user_id, device_id, type, value, action, priority, description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, device_id, type, value, action, priority, description, created_at, updated_at, created_by
	`, req.UserID, req.DeviceID, req.Type, req.Value, req.Action, req.Priority, req.Description, createdBy).Scan(
		&p.ID, &p.UserID, &p.DeviceID, &p.Type, &p.Value, &p.Action, &p.Priority, &p.Description, &p.CreatedAt, &p.UpdatedAt, &p.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("create policy: %w", err)
	}
	return p, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*AccessPolicy, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, device_id, type, value, action, priority, description, created_at, updated_at, created_by
		FROM access_policies
		WHERE user_id = $1
		ORDER BY priority ASC, created_at ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	defer rows.Close()

	var policies []*AccessPolicy
	for rows.Next() {
		p := &AccessPolicy{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.DeviceID, &p.Type, &p.Value, &p.Action, &p.Priority, &p.Description, &p.CreatedAt, &p.UpdatedAt, &p.CreatedBy); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*AccessPolicy, error) {
	p := &AccessPolicy{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, device_id, type, value, action, priority, description, created_at, updated_at, created_by
		FROM access_policies WHERE id = $1
	`, id).Scan(&p.ID, &p.UserID, &p.DeviceID, &p.Type, &p.Value, &p.Action, &p.Priority, &p.Description, &p.CreatedAt, &p.UpdatedAt, &p.CreatedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get policy: %w", err)
	}
	return p, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, req UpdatePolicyRequest) (*AccessPolicy, error) {
	p := &AccessPolicy{}
	err := r.pool.QueryRow(ctx, `
		UPDATE access_policies SET
			value       = COALESCE($2, value),
			action      = COALESCE($3, action),
			priority    = COALESCE($4, priority),
			description = COALESCE($5, description),
			updated_at  = now()
		WHERE id = $1
		RETURNING id, user_id, device_id, type, value, action, priority, description, created_at, updated_at, created_by
	`, id, req.Value, req.Action, req.Priority, req.Description).Scan(
		&p.ID, &p.UserID, &p.DeviceID, &p.Type, &p.Value, &p.Action, &p.Priority, &p.Description, &p.CreatedAt, &p.UpdatedAt, &p.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update policy: %w", err)
	}
	return p, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM access_policies WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete policy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) GetEffectivePolicy(ctx context.Context, userID uuid.UUID, deviceID *uuid.UUID) (*PolicySet, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, device_id, type, value, action, priority, description, created_at, updated_at, created_by
		FROM access_policies
		WHERE user_id = $1 AND (device_id IS NULL OR device_id = $2)
		ORDER BY priority ASC
	`, userID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("effective policy: %w", err)
	}
	defer rows.Close()

	set := &PolicySet{UserID: userID, DeviceID: deviceID}
	for rows.Next() {
		p := &AccessPolicy{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.DeviceID, &p.Type, &p.Value, &p.Action, &p.Priority, &p.Description, &p.CreatedAt, &p.UpdatedAt, &p.CreatedBy); err != nil {
			return nil, err
		}
		set.Rules = append(set.Rules, *p)
	}
	return set, rows.Err()
}
