-- +goose Up
-- Generic project storage is additive. Legacy content/revision tables remain intact.
CREATE TABLE project_objects (
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  sha256 TEXT NOT NULL CHECK(length(sha256) = 64),
  bytes INTEGER NOT NULL CHECK(bytes >= 0),
  PRIMARY KEY(project_id, sha256),
  UNIQUE(project_id, sha256, bytes)
);

CREATE TABLE project_versions (
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  id TEXT NOT NULL,
  parent_id TEXT,
  operation_id TEXT NOT NULL,
  request_sha256 TEXT NOT NULL,
  message TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY(project_id, id),
  UNIQUE(project_id, operation_id),
  FOREIGN KEY(project_id, parent_id) REFERENCES project_versions(project_id, id)
);

CREATE TABLE project_version_heads (
  project_id TEXT PRIMARY KEY REFERENCES story_projects(id) ON DELETE CASCADE,
  version_id TEXT,
  FOREIGN KEY(project_id, version_id) REFERENCES project_versions(project_id, id)
);

CREATE TABLE project_tree_ids (
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  id TEXT NOT NULL,
  kind TEXT NOT NULL CHECK(kind IN ('file', 'directory')),
  PRIMARY KEY(project_id, id),
  UNIQUE(project_id, id, kind)
);

CREATE TABLE project_version_entries (
  project_id TEXT NOT NULL,
  version_id TEXT NOT NULL,
  id TEXT NOT NULL,
  path TEXT NOT NULL,
  kind TEXT NOT NULL CHECK(kind IN ('file', 'directory')),
  sha256 TEXT,
  bytes INTEGER NOT NULL CHECK(bytes >= 0),
  PRIMARY KEY(project_id, version_id, id),
  UNIQUE(project_id, version_id, path),
  CHECK((kind = 'directory' AND sha256 IS NULL AND bytes = 0) OR
        (kind = 'file' AND sha256 IS NOT NULL)),
  FOREIGN KEY(project_id, version_id) REFERENCES project_versions(project_id, id) ON DELETE CASCADE,
  FOREIGN KEY(project_id, id, kind) REFERENCES project_tree_ids(project_id, id, kind),
  FOREIGN KEY(project_id, sha256, bytes) REFERENCES project_objects(project_id, sha256, bytes)
);

-- +goose Down
DROP TABLE project_version_entries;
DROP TABLE project_tree_ids;
DROP TABLE project_version_heads;
DROP TABLE project_versions;
DROP TABLE project_objects;
