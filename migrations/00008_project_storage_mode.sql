-- +goose Up
ALTER TABLE story_projects ADD COLUMN storage_mode TEXT NOT NULL DEFAULT 'legacy' CHECK(storage_mode IN ('legacy','files'));

-- +goose Down
ALTER TABLE story_projects DROP COLUMN storage_mode;
