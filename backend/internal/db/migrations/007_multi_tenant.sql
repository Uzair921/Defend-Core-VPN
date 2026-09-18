-- Migration 007: Multi-Tenant SaaS Model

-- =====================================================
-- 1. Organizations (Customers)
-- =====================================================
CREATE TABLE IF NOT EXISTS organizations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                TEXT NOT NULL,
    slug                TEXT NOT NULL UNIQUE,
    email               TEXT NOT NULL,
    phone               TEXT,
    website             TEXT,
    status              TEXT NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'suspended', 'trial', 'cancelled')),
    plan                TEXT NOT NULL DEFAULT 'basic'
                        CHECK (plan IN ('trial', 'basic', 'pro', 'enterprise')),
    max_users           INTEGER NOT NULL DEFAULT 10,
    max_services        INTEGER NOT NULL DEFAULT 2,
    bandwidth_limit_gb  INTEGER,
    billing_email       TEXT,
    subscription_start  TIMESTAMPTZ,
    subscription_end    TIMESTAMPTZ,
    monthly_price_cents INTEGER,
    metadata            JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_organizations_status ON organizations(status);
CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug);

-- =====================================================
-- 2. Organization Users
-- =====================================================
CREATE TABLE IF NOT EXISTS organization_users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role                TEXT NOT NULL DEFAULT 'user'
                        CHECK (role IN ('admin', 'user')),
    status              TEXT NOT NULL DEFAULT 'active',
    joined_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(organization_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_org_users_org ON organization_users(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_users_user ON organization_users(user_id);

-- =====================================================
-- 3. Organization Subscriptions
-- =====================================================
CREATE TABLE IF NOT EXISTS organization_subscriptions (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id       UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    service_type          TEXT NOT NULL REFERENCES vpn_service_types(code),
    quantity              INTEGER NOT NULL DEFAULT 1,
    max_users             INTEGER NOT NULL DEFAULT 10,
    status                TEXT NOT NULL DEFAULT 'active'
                          CHECK (status IN ('active', 'suspended', 'cancelled')),
    price_cents_per_month INTEGER,
    started_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at            TIMESTAMPTZ,
    UNIQUE(organization_id, service_type)
);

CREATE INDEX IF NOT EXISTS idx_org_subs_org ON organization_subscriptions(organization_id);

-- =====================================================
-- 4. Update users table
-- =====================================================
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_superadmin BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS organization_id UUID REFERENCES organizations(id);
ALTER TABLE users ADD COLUMN IF NOT EXISTS invited_by UUID REFERENCES users(id);
ALTER TABLE users ADD COLUMN IF NOT EXISTS invitation_accepted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_users_superadmin ON users(is_superadmin);
CREATE INDEX IF NOT EXISTS idx_users_org ON users(organization_id);

-- =====================================================
-- 5. Update vpn_services — org ownership
-- =====================================================
ALTER TABLE vpn_services ADD COLUMN IF NOT EXISTS organization_id UUID REFERENCES organizations(id);
CREATE INDEX IF NOT EXISTS idx_vpn_services_org ON vpn_services(organization_id);

-- =====================================================
-- 6. Invoices
-- =====================================================
CREATE TABLE IF NOT EXISTS invoices (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    invoice_number      TEXT NOT NULL UNIQUE,
    amount_cents        INTEGER NOT NULL,
    currency            TEXT NOT NULL DEFAULT 'USD',
    status              TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'paid', 'overdue', 'cancelled')),
    period_start        TIMESTAMPTZ NOT NULL,
    period_end          TIMESTAMPTZ NOT NULL,
    paid_at             TIMESTAMPTZ,
    due_at              TIMESTAMPTZ,
    items               JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_invoices_org ON invoices(organization_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);

-- =====================================================
-- 7. Platform Audit
-- =====================================================
CREATE TABLE IF NOT EXISTS platform_audit (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id            UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_role          TEXT,
    organization_id     UUID REFERENCES organizations(id) ON DELETE SET NULL,
    action              TEXT NOT NULL,
    target_type         TEXT,
    target_id           UUID,
    details             JSONB,
    ip_address          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_platform_audit_org ON platform_audit(organization_id);
CREATE INDEX IF NOT EXISTS idx_platform_audit_actor ON platform_audit(actor_id);

-- =====================================================
-- 8. Bootstrap: existing admin → SuperAdmin
-- =====================================================
UPDATE users
SET is_superadmin = TRUE
WHERE email = 'admin@defendcore.local';
