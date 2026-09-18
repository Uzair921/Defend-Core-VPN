package vpnctl

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
	ErrServerNotFound  = errors.New("vpn server not found")
	ErrSessionNotFound = errors.New("session not found")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ===== Servers =====

func (r *Repository) RegisterServer(ctx context.Context, name, publicIP, region, version string) (*VpnServer, error) {
	s := &VpnServer{}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO vpn_servers (name, public_ip, region, status, version, last_seen)
		VALUES ($1, $2, $3, 'online', $4, NOW())
		ON CONFLICT (name) DO UPDATE SET
			public_ip = EXCLUDED.public_ip,
			region = EXCLUDED.region,
			status = 'online',
			version = EXCLUDED.version,
			last_seen = NOW()
		RETURNING id, name, public_ip, region, status, version, last_seen, created_at
	`, name, publicIP, region, version).Scan(
		&s.ID, &s.Name, &s.PublicIP, &s.Region, &s.Status, &s.Version, &s.LastSeen, &s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("register server: %w", err)
	}
	return s, nil
}

func (r *Repository) Heartbeat(ctx context.Context, serverID uuid.UUID, status string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE vpn_servers
		SET status = $1, last_seen = NOW()
		WHERE id = $2
	`, status, serverID)
	if err != nil {
		return fmt.Errorf("heartbeat: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrServerNotFound
	}
	return nil
}

func (r *Repository) ListServers(ctx context.Context) ([]*VpnServer, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, public_ip, region, status, version, last_seen, created_at
		FROM vpn_servers
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*VpnServer
	for rows.Next() {
		s := &VpnServer{}
		if err := rows.Scan(&s.ID, &s.Name, &s.PublicIP, &s.Region,
			&s.Status, &s.Version, &s.LastSeen, &s.CreatedAt); err != nil {
			return nil, err
		}
		// Mark stale as offline
		if time.Since(s.LastSeen) > 60*time.Second {
			s.Status = "offline"
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ===== Sessions =====

func (r *Repository) StartSession(ctx context.Context, req StartSessionRequest) (*Session, error) {
	s := &Session{}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO sessions (user_id, device_id, server_id, client_ip, started_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, user_id, device_id, server_id, client_ip, started_at, bytes_in, bytes_out
	`, req.UserID, req.DeviceID, req.ServerID, req.ClientIP).Scan(
		&s.ID, &s.UserID, &s.DeviceID, &s.ServerID, &s.ClientIP,
		&s.StartedAt, &s.BytesIn, &s.BytesOut,
	)
	if err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}
	s.AssignedIP = req.AssignedIP
	s.Status = "active"
	return s, nil
}

func (r *Repository) UpdateSession(ctx context.Context, sessionID uuid.UUID, bytesIn, bytesOut int64) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE sessions SET bytes_in = $1, bytes_out = $2 WHERE id = $3 AND ended_at IS NULL
	`, bytesIn, bytesOut, sessionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (r *Repository) EndSession(ctx context.Context, sessionID uuid.UUID, bytesIn, bytesOut int64) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE sessions
		SET ended_at = NOW(), bytes_in = $1, bytes_out = $2
		WHERE id = $3 AND ended_at IS NULL
	`, bytesIn, bytesOut, sessionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (r *Repository) ListActiveSessions(ctx context.Context, userID *uuid.UUID) ([]*Session, error) {
	query := `
		SELECT id, user_id, device_id, server_id, client_ip, started_at, ended_at, bytes_in, bytes_out
		FROM sessions
		WHERE ended_at IS NULL
	`
	args := []interface{}{}
	if userID != nil {
		query += ` AND user_id = $1`
		args = append(args, *userID)
	}
	query += ` ORDER BY started_at DESC LIMIT 100`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Session
	for rows.Next() {
		s := &Session{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.DeviceID, &s.ServerID,
			&s.ClientIP, &s.StartedAt, &s.EndedAt, &s.BytesIn, &s.BytesOut); err != nil {
			return nil, err
		}
		s.Status = "active"
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) GetSession(ctx context.Context, id uuid.UUID) (*Session, error) {
	s := &Session{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, device_id, server_id, client_ip, started_at, ended_at, bytes_in, bytes_out
		FROM sessions WHERE id = $1
	`, id).Scan(&s.ID, &s.UserID, &s.DeviceID, &s.ServerID,
		&s.ClientIP, &s.StartedAt, &s.EndedAt, &s.BytesIn, &s.BytesOut)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	if s.EndedAt != nil {
		s.Status = "ended"
	} else {
		s.Status = "active"
	}
	return s, nil
}
