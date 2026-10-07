# Backlog

Active delivery priorities live in the [roadmap](roadmap.md) and [projects/sync plan](plans/2026-10-06-projects-and-sync.md). The items below must not displace that order.

## Deferred Items

### Add layered skill scopes and settings controls

- Status: deferred to roadmap P5 (skill parity)
- Found while: Milestone 4 information architecture/settings work and follow-up planning for skill loading.
- Why it matters: Open Edda currently treats project skill loading as the main settings concern, but authors will need three distinct skill scopes: built-in skills shipped with the app, global user-written skills available across projects, and project-local skills. System settings should manage global skill sources, while project settings should control which global and local skills are enabled for each project. Without this split, skill availability and enablement state will become ambiguous as soon as users install reusable personal skills.
- Evidence: `frontend/src/features/settings/SettingsPage.tsx` now hosts provider/model, skills, and script runtime administration; `frontend/src/features/skills/skillsThunks.ts` currently exposes project/session loading paths only; `docs/roadmap.md` Milestone 4 Phase 3.5 moved skill administration into settings, and Phase 4 is already scoped to assistant actions rather than skill-scope design.
- Not doing now because: The accepted 2026-10-06 priority is project storage and local synchronization, then Pocket Editor. Skill scopes belong with later skill parity rather than expanding the first delivery.
- Suggested next step: After projects/local sync and Pocket Editor, add a dedicated settings/project-settings plan that defines built-in, global user, and project-local skill sources; global enabled/disabled defaults; per-project enablement overrides; migration behavior for existing project skills; and UI/API changes for managing those scopes.
