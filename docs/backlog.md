# Backlog

## Open Items

### Add layered skill scopes and settings controls

- Status: open
- Found while: Milestone 4 information architecture/settings work and follow-up planning for skill loading.
- Why it matters: Open Edda currently treats project skill loading as the main settings concern, but authors will need three distinct skill scopes: built-in skills shipped with the app, global user-written skills available across projects, and project-local skills. System settings should manage global skill sources, while project settings should control which global and local skills are enabled for each project. Without this split, skill availability and enablement state will become ambiguous as soon as users install reusable personal skills.
- Evidence: `frontend/src/features/settings/SettingsPage.tsx` now hosts provider/model, skills, and script runtime administration; `frontend/src/features/skills/skillsThunks.ts` currently exposes project/session loading paths only; `docs/roadmap.md` Milestone 4 Phase 3.5 moved skill administration into settings, and Phase 4 is already scoped to assistant actions rather than skill-scope design.
- Not doing now because: The current PR is scoped to Milestone 4 IA correction plus the Phase 4 assistant-actions plan; adding new skill-scope data models and settings screens would expand backend, frontend, import, and permission semantics.
- Suggested next step: Add a dedicated settings/project-settings phase plan after Phase 3.5 that defines built-in, global user, and project-local skill sources; global enabled/disabled defaults; per-project enablement overrides; migration behavior for existing project skills; and UI/API changes for managing those scopes.
