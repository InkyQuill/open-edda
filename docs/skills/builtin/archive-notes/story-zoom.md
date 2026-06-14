# story-zoom

- Source path: `docs/skills/suggested/fiction/structure/story-zoom/`
- Final disposition: `deferred`
- Why it is not included as a 3.5 built-in skill: the source skill is built around Markdown directory watching, change logs, and file-level synchronization across `pitch/`, `structure/`, `scenes/`, `entities/`, and `manuscript/`. Writer Milestone 3.5 is curating a Writer-native skill shelf, not shipping file watchers or a consistency dashboard.
- What would need to change for reconsideration: Writer would need a database-backed consistency workflow that can inspect Story Text, Story Bible Entries, Entry Sections, Project Notes, and revisions without relying on filesystem watchers or terminal commands.
- Scripts/assets involved: yes. The source folder includes `scripts/init.ts`, `scripts/status.ts`, `scripts/watcher.ts`, plus templates under `templates/`.
- Future roadmap mapping: yes. Revisit through the roadmap's Later milestone for the Story Consistency Dashboard.
