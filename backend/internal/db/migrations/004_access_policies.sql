-- Migration 004: access_policies for per-user/device access control

CREATE TABLE IF NOT EXISTS access_policies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id   UUID REFERENCES devices(id) ON DELETE CASCADE,

    type        TEXT NOT NULL CHECK (type IN ('cidr', 'domain', 'url', 'ip', 'port')),
    value       TEXT NOT NULL,
    action      TEXT NOT NULL CHECK (action IN ('allow', 'deny')),
    priority    INTEGER NOT NULL DEFAULT 100,

    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  UUID REFERENCES users(id),

    UNIQUE(user_id, device_id, type, value)
);

CREATE INDEX IF NOT EXISTS idx_access_policies_user   ON access_policies(user_id);
CREATE INDEX IF NOT EXISTS idx_access_policies_device ON access_policies(device_id);
CREATE INDEX IF NOT EXISTS idx_access_policies_type   ON access_policies(type);

-- Policy groups for group-based access
CREATE TABLE IF NOT EXISTS policy_groups (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS policy_group_members (
    group_id    UUID NOT NULL REFERENCES policy_groups(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS policy_group_rules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL REFERENCES policy_groups(id) ON DELETE CASCADE,
    type        TEXT NOT NULL CHECK (type IN ('cidr', 'domain', 'url', 'ip', 'port')),
    value       TEXT NOT NULL,
    action      TEXT NOT NULL CHECK (action IN ('allow', 'deny')),
    priority    INTEGER NOT NULL DEFAULT 100,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- DNS policies for domain-based access
CREATE TABLE IF NOT EXISTS dns_policies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain      TEXT NOT NULL,
    resolved_ip TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, domain)
);

CREATE INDEX IF NOT EXISTS idx_dns_policies_user ON dns_policies(user_id);
