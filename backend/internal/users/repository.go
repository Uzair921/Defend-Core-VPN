// Package users provides account storage and lookups for the platform.
package users

import (
"context"
"errors"
"fmt"
"strings"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors for the users domain.
var (
ErrNotFound    = errors.New("user not found")
ErrEmailExists = errors.New("email already exists")
)

// Repository provides persistence for user accounts.
type Repository struct {
pool *pgxpool.Pool
}

// NewRepository constructs a Repository backed by the given pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
return &Repository{pool: pool}
}

const userColumns = `id, email, password_hash, role, status,
is_superadmin, organization_id, created_at, updated_at`

// Create inserts a new user account.
func (r *Repository) Create(ctx context.Context, email, passwordHash, role string) (*User, error) {
const q = `INSERT INTO users (email, password_hash, role)
VALUES ($1, $2, $3)
RETURNING ` + userColumns

u := &User{}
err := r.pool.QueryRow(ctx, q, email, passwordHash, role).Scan(
&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Status,
&u.IsSuperAdmin, &u.OrganizationID, &u.CreatedAt, &u.UpdatedAt,
)
if err != nil {
if isUniqueViolation(err) {
return nil, ErrEmailExists
}
return nil, fmt.Errorf("create user: %w", err)
}
return u, nil
}

// GetByEmail looks up a user by email address.
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
const q = `SELECT ` + userColumns + ` FROM users WHERE email = $1`
return scanUser(r.pool.QueryRow(ctx, q, email))
}

// GetByID looks up a user by ID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
const q = `SELECT ` + userColumns + ` FROM users WHERE id = $1`
return scanUser(r.pool.QueryRow(ctx, q, id))
}

// IsSuperAdmin reports whether the given user has platform-operator privileges.
// It is used by the SuperAdmin middleware to gate sensitive routes.
func (r *Repository) IsSuperAdmin(ctx context.Context, userID any) (bool, error) {
id, err := toUUID(userID)
if err != nil {
return false, fmt.Errorf("invalid user id: %w", err)
}

const q = `SELECT is_superadmin FROM users WHERE id = $1`
var isAdmin bool
if err := r.pool.QueryRow(ctx, q, id).Scan(&isAdmin); err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return false, ErrNotFound
}
return false, fmt.Errorf("check superadmin: %w", err)
}
return isAdmin, nil
}

func scanUser(row pgx.Row) (*User, error) {
u := &User{}
err := row.Scan(
&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Status,
&u.IsSuperAdmin, &u.OrganizationID, &u.CreatedAt, &u.UpdatedAt,
)
if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, ErrNotFound
}
return nil, fmt.Errorf("scan user: %w", err)
}
return u, nil
}

// toUUID accepts either a uuid.UUID or a string and returns a UUID.
// This keeps the middleware interface decoupled from the UUID type.
func toUUID(v any) (uuid.UUID, error) {
switch x := v.(type) {
case uuid.UUID:
return x, nil
case *uuid.UUID:
if x == nil {
return uuid.Nil, errors.New("nil uuid")
}
return *x, nil
case string:
return uuid.Parse(x)
default:
return uuid.Nil, fmt.Errorf("unsupported type %T", v)
}
}

func isUniqueViolation(err error) bool {
if err == nil {
return false
}
msg := err.Error()
return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
