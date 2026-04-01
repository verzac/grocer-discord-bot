CREATE TABLE IF NOT EXISTS user_sessions (
    discord_user_id TEXT NOT NULL PRIMARY KEY,
    discord_access_token_enc BLOB NOT NULL,
    discord_refresh_token_enc BLOB NOT NULL,
    discord_token_expires_at DATETIME,
    gro_refresh_token_hash TEXT NOT NULL,
    gro_refresh_expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_gro_refresh_expires ON user_sessions(gro_refresh_expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_gro_refresh_hash ON user_sessions(gro_refresh_token_hash);
