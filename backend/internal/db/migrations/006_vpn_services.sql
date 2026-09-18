-- Migration 006: Multi-VPN Service Types

CREATE TABLE IF NOT EXISTS vpn_service_types (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code                    TEXT NOT NULL UNIQUE,
    name                    TEXT NOT NULL,
    description             TEXT,
    icon                    TEXT,
    category                TEXT NOT NULL,
    supports_split_tunnel   BOOLEAN NOT NULL DEFAULT FALSE,
    supports_kill_switch    BOOLEAN NOT NULL DEFAULT FALSE,
    supports_mfa            BOOLEAN NOT NULL DEFAULT TRUE,
    supports_policies       BOOLEAN NOT NULL DEFAULT TRUE,
    supports_ztna           BOOLEAN NOT NULL DEFAULT FALSE,
    default_routes          JSONB NOT NULL DEFAULT '["0.0.0.0/0"]'::jsonb,
    default_dns             TEXT[] DEFAULT ARRAY['1.1.1.1', '8.8.8.8'],
    default_mtu             INTEGER NOT NULL DEFAULT 1420,
    default_keepalive       INTEGER NOT NULL DEFAULT 25,
    display_order           INTEGER NOT NULL DEFAULT 100,
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO vpn_service_types 
(code, name, description, icon, category, supports_split_tunnel, supports_kill_switch, default_routes, default_dns, display_order)
VALUES
('full_tunnel', 'Full Tunnel VPN', 'Route all your internet traffic through secure VPN.', '🔒', 'personal', FALSE, TRUE, '["0.0.0.0/0"]'::jsonb, ARRAY['1.1.1.1', '8.8.8.8'], 10),
('split_tunnel', 'Split Tunnel VPN', 'Only corporate resources through VPN.', '⚡', 'corporate', TRUE, FALSE, '["10.0.0.0/8", "192.168.0.0/16"]'::jsonb, ARRAY['10.0.0.1', '1.1.1.1'], 20),
('remote_access', 'Remote Access VPN', 'Access your office network from anywhere.', '👤', 'corporate', TRUE, TRUE, '["10.0.0.0/8"]'::jsonb, ARRAY['10.0.0.1'], 30),
('site_to_site', 'Site-to-Site VPN', 'Connect multiple office locations.', '🏢', 'enterprise', FALSE, FALSE, '["10.0.0.0/8"]'::jsonb, ARRAY['10.0.0.1'], 40),
('zero_trust', 'Zero Trust VPN', 'Per-application access with verification.', '🔐', 'enterprise', TRUE, TRUE, '[]'::jsonb, ARRAY['10.0.0.1'], 50),
('stealth', 'Stealth VPN', 'Obfuscated traffic bypasses firewalls.', '🎭', 'personal', FALSE, TRUE, '["0.0.0.0/0"]'::jsonb, ARRAY['1.1.1.1'], 60)
ON CONFLICT (code) DO NOTHING;

CREATE TABLE IF NOT EXISTS vpn_services (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                TEXT NOT NULL,
    slug                TEXT NOT NULL UNIQUE,
    service_type        TEXT NOT NULL REFERENCES vpn_service_types(code),
    description         TEXT,
    server_id           UUID REFERENCES vpn_servers(id) ON DELETE SET NULL,
    subnet              TEXT NOT NULL,
    server_ip           TEXT NOT NULL,
    dns_servers         TEXT[] DEFAULT ARRAY['10.8.0.1'],
    max_clients         INTEGER NOT NULL DEFAULT 100,
    current_clients     INTEGER NOT NULL DEFAULT 0,
    bandwidth_limit_mbps INTEGER,
    custom_routes       JSONB,
    custom_mtu          INTEGER,
    status              TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled', 'maintenance')),
    owner_user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
    is_public           BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_vpn_services_type ON vpn_services(service_type);
CREATE INDEX IF NOT EXISTS idx_vpn_services_status ON vpn_services(status);

CREATE TABLE IF NOT EXISTS user_vpn_services (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_id          UUID NOT NULL REFERENCES vpn_services(id) ON DELETE CASCADE,
    role                TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    granted_by          UUID REFERENCES users(id) ON DELETE SET NULL,
    granted_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at          TIMESTAMPTZ,
    UNIQUE(user_id, service_id)
);

CREATE TABLE IF NOT EXISTS device_vpn_configs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id           UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    service_id          UUID NOT NULL REFERENCES vpn_services(id) ON DELETE CASCADE,
    assigned_ip         TEXT NOT NULL,
    custom_routes       JSONB,
    custom_dns          TEXT[],
    is_default          BOOLEAN NOT NULL DEFAULT FALSE,
    last_connected_at   TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(device_id, service_id)
);

CREATE TABLE IF NOT EXISTS vpn_service_audit (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id          UUID REFERENCES vpn_services(id) ON DELETE CASCADE,
    user_id             UUID REFERENCES users(id) ON DELETE SET NULL,
    action              TEXT NOT NULL,
    details             JSONB,
    ip_address          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_vpn_service_audit_service ON vpn_service_audit(service_id);
CREATE INDEX IF NOT EXISTS idx_vpn_service_audit_user ON vpn_service_audit(user_id);
CREATE INDEX IF NOT EXISTS idx_device_vpn_configs_device ON device_vpn_configs(device_id);
CREATE INDEX IF NOT EXISTS idx_device_vpn_configs_service ON device_vpn_configs(service_id);
