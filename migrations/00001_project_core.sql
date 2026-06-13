-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE authors (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE story_projects (
  id TEXT PRIMARY KEY,
  author_id TEXT NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  slug TEXT NOT NULL,
  language TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(author_id, slug)
);

CREATE TABLE content_items (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  kind TEXT NOT NULL CHECK(kind IN ('chapter', 'story_bible_entry', 'writing_brief', 'project_note')),
  title TEXT NOT NULL,
  slug TEXT NOT NULL,
  body_markdown TEXT NOT NULL DEFAULT '',
  metadata_json TEXT NOT NULL DEFAULT '{}',
  sort_order INTEGER NOT NULL DEFAULT 0,
  current_revision INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(id, project_id),
  UNIQUE(project_id, kind, slug)
);

CREATE TABLE entry_sections (
  id TEXT PRIMARY KEY,
  content_item_id TEXT NOT NULL REFERENCES content_items(id) ON DELETE CASCADE,
  heading TEXT NOT NULL,
  body_markdown TEXT NOT NULL,
  sort_order INTEGER NOT NULL,
  UNIQUE(content_item_id, heading)
);

CREATE TABLE entry_relations (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  source_item_id TEXT NOT NULL,
  target_item_id TEXT,
  target_title TEXT NOT NULL,
  relation_type TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  FOREIGN KEY (source_item_id, project_id) REFERENCES content_items(id, project_id) ON DELETE CASCADE,
  FOREIGN KEY (target_item_id, project_id) REFERENCES content_items(id, project_id) ON DELETE CASCADE
);

CREATE TABLE revisions (
  id TEXT PRIMARY KEY,
  content_item_id TEXT NOT NULL REFERENCES content_items(id) ON DELETE CASCADE,
  revision_number INTEGER NOT NULL,
  body_markdown TEXT NOT NULL,
  metadata_json TEXT NOT NULL DEFAULT '{}',
  reason TEXT NOT NULL,
  created_by TEXT NOT NULL CHECK(created_by IN ('author', 'import', 'restore', 'agent')),
  created_at TEXT NOT NULL,
  UNIQUE(content_item_id, revision_number)
);

CREATE TABLE attached_notes (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  content_item_id TEXT,
  selection_start INTEGER,
  selection_end INTEGER,
  title TEXT NOT NULL,
  body_markdown TEXT NOT NULL,
  source TEXT NOT NULL CHECK(source IN ('author', 'read_and_check')),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (content_item_id, project_id) REFERENCES content_items(id, project_id) ON DELETE CASCADE
);

CREATE VIRTUAL TABLE content_search USING fts5(
  title,
  body_markdown,
  content='content_items',
  content_rowid='rowid'
);

-- +goose StatementBegin
CREATE TRIGGER content_items_ai AFTER INSERT ON content_items BEGIN
  INSERT INTO content_search(rowid, title, body_markdown)
  VALUES (new.rowid, new.title, new.body_markdown);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER content_items_ad AFTER DELETE ON content_items BEGIN
  INSERT INTO content_search(content_search, rowid, title, body_markdown)
  VALUES ('delete', old.rowid, old.title, old.body_markdown);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER content_items_au AFTER UPDATE ON content_items BEGIN
  INSERT INTO content_search(content_search, rowid, title, body_markdown)
  VALUES ('delete', old.rowid, old.title, old.body_markdown);
  INSERT INTO content_search(rowid, title, body_markdown)
  VALUES (new.rowid, new.title, new.body_markdown);
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS content_items_au;
DROP TRIGGER IF EXISTS content_items_ad;
DROP TRIGGER IF EXISTS content_items_ai;
DROP TABLE IF EXISTS content_search;
DROP TABLE IF EXISTS attached_notes;
DROP TABLE IF EXISTS revisions;
DROP TABLE IF EXISTS entry_relations;
DROP TABLE IF EXISTS entry_sections;
DROP TABLE IF EXISTS content_items;
DROP TABLE IF EXISTS story_projects;
DROP TABLE IF EXISTS authors;
