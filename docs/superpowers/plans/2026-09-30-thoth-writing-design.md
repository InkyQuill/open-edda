# Thoth Writing Design Implementation Plan

**Goal:** Calm flat Edda workspace with Galley palettes, borderless Galley Editor, and server-persisted author theme selection.
**Architecture:** An authenticated preferences endpoint stores a catalog theme ID per author. A React theme provider applies the package's CSS variables to the document root; application tokens alias those roles. Existing editor selection, save and review adapters remain intact.
**Tech Stack:** Go, SQLite, React, Radix primitives, Tailwind, Galley Editor/Themes 0.16.0, Lucide.
**Spec:** ../../../DESIGN.md

## Constraints
- Thoth Light is the initial default; Thoth Dark is the default dark choice.
- Use published @inkyquill packages; no changes to the Galley source repository.
- Preserve Markdown data and selection offsets.
- Preferences are scoped to the authenticated author, never a supplied author ID.

## Tasks
- [x] Test and implement authenticated GET/PUT /api/auth/preferences with default, persistence, validation and author isolation.
- [x] Integrate catalog, root theme provider, settings selector and save/error/loading states; test palette mapping and fallback.
- [x] Replace old editor package, inherit theme, add Lucide toolbar icons and remove framing.
- [x] Apply flat surfaces and navigation to workspace and other screens; document design rules.
- [x] Run frontend tests/build and backend tests; inspect light/dark desktop/mobile and theme reload/save flows.

## Review focus
Account changes and expired tokens must not reuse another author's theme. Failed saves must be visible and preserve the stored choice. All catalog palettes must map both editor and UI roles. Editor theme changes must preserve unsaved text. Small screens must retain access to settings and panels.
