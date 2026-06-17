# Milestone 4 System Settings And IA Correction Plan

**Goal:** Correct the Milestone 4 information architecture before assistant action wiring continues: provider/model administration and skill management belong in settings, assistant mode's right panel is chat-only, project/content creation controls are first-class, and the projects page has an intentional wireframed layout.

**Phase Position:** Milestone 4 Phase 3.5. This plan is a prerequisite for Phase 4 assistant actions because action UX should attach to the corrected workspace surfaces, not to the temporary right-panel settings layout.

## Product Decisions

- Provider connection settings are system-level, not project-level. API keys, base URLs, provider configs, and available model variants must be managed from a System Settings page.
- Model selection for chat/actions should use the system-level model catalog. A project or session may remember the active model choice, but the provider configuration itself is not owned by a story project.
- Skill administration is not part of the assistant chat drawer. Imported/installed skills, script approvals, and runtime/admin controls move to Settings. Session-level skill selection remains a separate product question and should not reappear in the right panel by default.
- Assistant mode's right drawer is only chat: transcript, message composer, chat state, and minimal model disclosure/status. No provider forms, skill browsers, script runtime panels, or project settings live there.
- The workspace must expose creation controls for chapters, worldbuilding entries, briefs, notes, and other content kinds supported by the backend.
- The `/projects` page needs a real layout/wireframe, not a raw list.

## Scope

### 1. System Settings Route

Create a dedicated settings surface, for example `/settings`, reachable from the projects page and workspace chrome.

It must include:

- Provider connection management: provider name, OpenAI-compatible base URL, API key entry/update, enabled/disabled status, and clear "secret not shown after save" behavior.
- Model catalog management: list provider models/variants available to Open Edda, choose which models are usable in the app, edit display names and generation defaults, and expose pricing/usage fields where the backend already supports them.
- Skill administration: browse installed skills, import skills, inspect skill metadata/files/routing hints, see script-disabled status, manage script approvals, and view script run history.
- System status affordances: show whether at least one provider and one usable model are configured.

### 2. Project Settings Boundary

Create or reserve a project settings surface for project-owned configuration only.

Project settings may include:

- project title, slug, language, export/import options;
- project prompt profile or writing preferences;
- project-level defaults that reference system resources, such as default model variant, without owning provider credentials;
- project skill enablement only if the product keeps skills scoped by project in the data model.

Project settings must not be the only place to configure provider credentials.

### 3. Assistant Drawer Cleanup

Keep the right drawer in assistant mode focused on chat.

Acceptance:

- The right panel contains transcript/session list or active chat state, message composer, and minimal model status/disclosure.
- It does not render provider settings forms.
- It does not render skill administration, installed skill counts, session skill pickers, script runtime controls, or project settings.
- Empty states stay short and action-oriented.

### 4. Content Creation Controls

Add visible creation controls for content kinds supported by the backend.

Acceptance:

- From the workspace content drawer, the author can create a chapter.
- From world/notes/briefs surfaces, the author can create the corresponding item type.
- New items become selected immediately and navigate to their canonical route.
- Creation controls handle empty projects; an empty project should not be a dead end.
- Tests cover creating at least one chapter and one non-chapter content item through the UI/API boundary.

### 5. Projects Page Redesign

Replace the bare project list with an intentional projects dashboard.

Acceptance:

- `/projects` has clear page chrome, settings access, logout, and project creation/import actions.
- Project cards show title, language, updated timestamp or useful metadata, and open affordance.
- Empty state offers creation/import instead of only saying no projects exist.
- Layout is responsive and visually consistent with the workspace design system.
- It is acceptable for this phase to be wireframe-quality, but spacing, hierarchy, and buttons must be deliberate.

## Implementation Notes

- Reuse existing provider/model APIs where possible; do not duplicate provider secrets into project state.
- Reuse existing skill APIs initially, even if backend tables remain project-scoped. If the UX says "system settings" but the schema remains project-scoped, document the mismatch and decide whether to migrate skills later.
- Keep `ModelStatus` as disclosure/status only; move `ModelSettingsPanel` out of workspace drawers.
- Keep `SkillsPanel`, `SkillChipsPanel`, and script runtime panels available for relocation, but do not render them in `AssistantDrawer`.
- Add browser smoke coverage for `/settings`, `/projects`, project open, content creation, and assistant drawer cleanliness.

## Open Questions

- Should skills remain project-scoped in storage, or should built-in/imported skills become instance-scoped with per-project enablement?
- Should the active model be a system default, project default, session choice, or layered fallback?
- Should content creation use modals, inline buttons in each drawer section, or a command menu?
- Should system settings be available before a project exists? Current answer should be yes.

## Verification

- Frontend unit tests for settings reducers/components added or updated.
- Backend tests only if API shape changes.
- Browser smoke proves:
  - `/projects` is styled and has settings/project creation affordances;
  - `/settings` can show provider/model/skill administration surfaces;
  - assistant right drawer contains chat/model disclosure only;
  - chapter and non-chapter content creation controls exist and create navigable items;
  - console has no application errors.
