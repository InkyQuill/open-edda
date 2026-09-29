# Open Edda Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

The primary user is an individual author writing long-form fiction. Edda supports sustained, independent work on a novel or series, including drafting, revision, story planning and maintaining the story world.

The author confirmed the primary usage scene: writing at a computer; reading and small edits on a phone. Mobile web remains part of the web product, not a separate native application.

## Product Purpose

Edda is a private writing studio that connects the manuscript, story-world material, notes and AI-assisted work within one story project. Success means the author can write and revise comfortably, find relevant material, and use an assistant with project context while retaining creative control.

## Positioning

The assistant works with the whole story project through discovery, search and scoped reads, rather than relying only on the currently selected text. Chapters, story bible entries, notes and writing briefs provide connected context for generation, revision and analysis.

This is a product mechanism, not a claim that every project is sent to a model in full or that Edda is unique in the market.

## Operating Context

- Self-hosted/local use is the current priority. The application is currently a single-author instance.
- Authors select or create a story project, navigate chapters and reference material, write Markdown-compatible prose, and optionally open assistant or review tools.
- Desktop supports sustained writing and project organization. Mobile supports reading, small edits and access to contextual tools.
- OpenAI-compatible providers and model variants are configured in settings. Relevant project content may be sent to the configured provider for an assistant action; private/self-hosted does not imply exclusively local model processing.
- The target project model is file-first Markdown with SQLite-backed operational indexes and settings. The transition is ongoing; do not present the entire target architecture as already implemented.

## Capabilities and Constraints

Existing product capabilities are documented in README.md and the roadmap:

- Story projects; ordered chapters; story bible entries, sections and relations; writing briefs; project notes and attached notes.
- Markdown-native editing with Galley Editor, selection-aware actions, saving and revision history.
- Assistant sessions for conversation, continuation, rewriting and read/check reports, supported by project-context tools.
- Project skills and reviewed script execution; provider/model and administration settings.
- Markdown import/export and authenticated author workflows.
- Application/editor palette selection stored in the authenticated author's server-side profile preferences.

Constraints confirmed by the author:

- Use `@inkyquill/galley-editor` as the editor and `@inkyquill/galley-themes` as the palette source. The approved integration details belong in DESIGN.md.
- Keep authored text, project context, selection behavior and revision workflows intact during interface changes.
- Multi-user collaboration, broader deployment hardening and future file-first work are roadmap subjects, not assumed current features.

Open decisions: no product-specific accessibility standard, commercial model or wider audience expansion was established in this init. Future work must not invent commitments in those areas.

## Brand Commitments

The formal product name is **Open Edda**. The short in-app name is **Edda**. Use CONTEXT.md for established terminology, including Story Project, Writing Workspace, Story Bible, Writing Brief and Author.

The author explicitly requested a calm, pleasant, unobtrusive workspace for writing. Respect the visual decisions already approved in DESIGN.md; init does not replace that authority.

## Evidence on Hand

- README.md: current product shape, stack, self-hosted setup and verification commands.
- CONTEXT.md: product terminology and subsystem boundaries. Some architecture descriptions are historical; use the current README and roadmap to distinguish shipped behavior from targets.
- docs/roadmap.md: development status and future work.
- frontend/src/app/router/routes.tsx: login, projects, settings and project/content workspace routes.
- DESIGN.md: approved interface rules and reference provenance.
- Existing code and automated tests demonstrate current workflows. No customer testimonials, adoption figures or performance guarantees were supplied for this record.

## Product Principles

1. The author retains creative authority; AI assistance serves the writing project.
2. The manuscript is the main task. Context and tools must remain easy to reach without constantly competing for attention.
3. Treat a story as connected prose, canon, planning and notes; preserve their different meanings.
4. Keep content portable and changes reviewable. Do not trade away Markdown compatibility or revision safety for presentation.
5. Favor dependable daily writing over expanding administration or decorative features.

## Context Confirmation

The author confirmed the audience, purpose and desktop/mobile usage scene on 2026-09-30 during Impeccable init. Implementation capabilities and boundaries above are grounded in the listed repository sources; open decisions remain explicitly undecided.
