-- +goose Up
CREATE TABLE author_preferences (
    author_id TEXT PRIMARY KEY REFERENCES authors(id) ON DELETE CASCADE,
    theme_id TEXT NOT NULL DEFAULT 'thoth-light'
);

-- +goose Down
DROP TABLE author_preferences;
