-- +goose Up
PRAGMA foreign_keys = ON;

ALTER TABLE entry_sections ADD COLUMN current_revision INTEGER NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE entry_sections RENAME TO entry_sections_old;

CREATE TABLE entry_sections (
  id TEXT PRIMARY KEY,
  content_item_id TEXT NOT NULL REFERENCES content_items(id) ON DELETE CASCADE,
  heading TEXT NOT NULL,
  body_markdown TEXT NOT NULL,
  sort_order INTEGER NOT NULL,
  UNIQUE(content_item_id, heading)
);

INSERT INTO entry_sections (id, content_item_id, heading, body_markdown, sort_order)
SELECT id, content_item_id, heading, body_markdown, sort_order
FROM entry_sections_old;

DROP TABLE entry_sections_old;
