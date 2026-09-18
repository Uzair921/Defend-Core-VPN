-- Add last_seen column to vpn_servers
ALTER TABLE vpn_servers
    ADD COLUMN IF NOT EXISTS last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Index for active sessions queries
CREATE INDEX IF NOT EXISTS idx_sessions_active
    ON sessions(user_id, started_at DESC)
    WHERE ended_at IS NULL;
