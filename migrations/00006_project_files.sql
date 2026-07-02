-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE project_files (
  id TEXT NOT NULL,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  relative_path TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('story', 'character', 'worldbuilding', 'storyline', 'draft', 'guidance', 'skill')),
  title TEXT NOT NULL,
  sha256 TEXT NOT NULL,
  bytes INTEGER NOT NULL,
  updated_at TEXT NOT NULL,
  PRIMARY KEY (project_id, id),
  UNIQUE(project_id, relative_path)
);

CREATE INDEX idx_project_files_project_kind_path ON project_files(project_id, kind, relative_path);

-- +goose Down
DROP INDEX IF EXISTS idx_project_files_project_kind_path;
DROP TABLE IF EXISTS project_files;
