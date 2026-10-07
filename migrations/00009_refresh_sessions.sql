-- +goose Up
CREATE TABLE refresh_sessions (
  token_hash TEXT PRIMARY KEY,
  author_id TEXT NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
  expires_at INTEGER NOT NULL
);
CREATE INDEX refresh_sessions_expiry ON refresh_sessions(expires_at);

-- +goose Down
DROP TABLE refresh_sessions;
