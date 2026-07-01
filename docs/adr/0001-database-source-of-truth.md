# Project Folder as the Story Source of Truth

Open Edda treats an Edda project folder as the authoritative home for story prose, story bible material, writing briefs, project notes, local skills, and other author-owned Markdown files. The service reads and writes one defined project layout, so an author can work in the web app or in normal editors on their computers without turning Open Edda into a git client.

SQLite remains part of the architecture, but it is not the canonical store for author prose. It stores indexes, search rows, project maps, prompt records, activity, session state, assistant caches, script runtime records, and mirrors of checkpoint metadata. If the database is lost, Open Edda should be able to rebuild project content and checkpoint history from the folder plus `.edda/` metadata. Data that only exists in operational tables may be lost unless a later ADR explicitly persists it into `.edda/`.

This supersedes the earlier database-source-of-truth assumption. Existing database-backed project APIs can remain as implementation scaffolding while the file-first migration proceeds, but new product docs should describe the Edda project folder and `.edda/` metadata as the durable project model.
