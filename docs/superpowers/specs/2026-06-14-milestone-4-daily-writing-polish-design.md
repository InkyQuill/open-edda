# Milestone 4 Daily Writing Polish Design

## Purpose

Milestone 4 turns Open Edda from an API-validation interface into a daily writing workspace. The milestone focuses on UI/UX structure: routing, layout, editor framing, mode presets, mobile behavior, state management, and component boundaries. It does not aim to finalize visual styling.

The workspace must support different author workflows. Some authors primarily draft by hand and only consult the AI. Others want an assistant and a Generate action close to the text. Review workflows need chat and analysis tools nearby without making the text feel secondary.

## Product Principles

- The editor is the center of the product in every mode.
- Prose line length should stay comfortable. The editor text column should be centered and constrained to roughly 700-800px on desktop and tablet, including Assistant mode.
- AI actions that act on the editor must be local to the editor. They must not depend on a side panel that may be hidden on mobile.
- Generate is a frequent action and gets its own persistent composer under the editor.
- Chat, worldbuilding/context lookup, notes, revisions, and model/settings surfaces are panels around the editor, not replacements for the editor.
- Mobile is editor-first. Panels become drawers or sheets over the editor.

## Routing

Milestone 4 should add React Router and move away from a single monolithic `App.tsx` control flow.

Routes:

```txt
/login
/projects
/projects/:projectId
/projects/:projectId/content/:contentKind/:contentId
```

Routes identify durable navigation state: authenticated area, selected project, selected content kind, and selected content item. Deep links to project content are required. The route includes `contentKind` so direct links to story bible entries, writing briefs, and project notes do not depend on local Redux or localStorage state.

Workspace mode, panel state, drawer widths, active tabs, selected action modal, and mobile sheets are not URL search params. They live in Redux and persisted local storage. Search params for `mode` or `panel` are intentionally out of scope because they would duplicate workspace state in the URL and create unclear ownership.

## Workspace Presets

The workspace has three author-selectable presets.

### Draft / Focused Writing

Draft mode is for authors who mostly write themselves. The editor gets the calmest surface and the right drawer is closed by default. The left drawer provides quick access to chapters, briefs, worldbuilding, and notes. Chat, worldbuilding lookup, and attached notes open as temporary drawers or sheets and should not permanently resize the prose column.

The Generate composer remains visible because some authors still want occasional AI continuation without switching into Assistant mode.

### Assistant

Assistant mode keeps the editor primary while showing the right assistant drawer. The right drawer contains chat, quick action status/results, skill chips, recent activity, and model disclosure.

The editor composer remains the primary Generate entry point. The chat drawer is for conversation and co-writing context, not the only place generation can start. The left drawer may stay open on wide screens and collapse on narrower screens.

### Review

Review mode keeps the text central but prioritizes chat and analysis tools in the right drawer. This mode is not primarily a diff application. Revision list, diff/restore UI, Read and Check reports, activity, and attached notes are available nearby, but the main workflow is chat/tools-first review and analysis.

## Desktop Layout

Desktop uses a dual-drawer model:

- A narrow mode/navigation rail.
- A left project/context drawer for content navigation, chapter lists, story bible/worldbuilding lookup, briefs, and notes.
- A centered editor canvas with a max-width prose column.
- A right assistant/review drawer whose default content depends on the active preset.

Panels can resize or collapse, but panel changes must not stretch prose lines beyond the editor's comfortable max width. In Draft mode, panels should behave more like temporary overlays. In Assistant and Review modes, the right drawer can remain persistent.

## Mobile Layout

Mobile uses an editor-first layout:

- The editor takes the screen.
- Files, Assistant, Review, World/Notes, and Model surfaces open as sheets or slideovers.
- The Generate composer remains attached below the editor area.
- Selection actions use a sticky toolbar instead of a floating selection bubble.
- Rewrite, Check, and Note actions open bottom sheets.
- Mobile sheets must preserve partially typed instructions if dismissed temporarily.

## Editor And Local Actions

The editor surface should be prepared for Galley Editor integration. Implementation may stage the adapter, but the architecture should not continue to treat the editor as a read-only textarea.

Editor-local actions:

- Desktop selected text shows a contextual bubble with `Rewrite`, `Check`, and `Note`.
- Mobile selected text exposes those actions in a sticky toolbar.
- `Generate` is separate from selection actions and lives in a persistent composer below the editor.
- The Generate composer contains an instruction input and an AI icon/button.
- Generate runs at the current cursor or selected insertion point and uses the current content revision.
- Rewrite and Check open a modal on desktop and bottom sheet on mobile before running.
- Rewrite/Check modal or sheet shows:
  - action title,
  - ellipsized preview of selected text,
  - instruction input,
  - cancel and preview/run action.
- Note opens an Attached Note flow linked to the selected range.

Selection and cursor offsets sent to the backend must be UTF-8 byte offsets, not UTF-16 textarea positions.

## UI Foundation

Milestone 4 should introduce Tailwind v4 and shadcn/ui as structural foundations for Open Edda's own interface. The goal is layout, interaction, accessibility, and component organization, not final brand styling.

Use shadcn primitives early for:

- `Button`, `Input`, `Textarea`
- `Dialog`, `Sheet`, `Tabs`
- `Resizable` where appropriate for desktop drawer widths
- `DropdownMenu`, `Command`
- `Tooltip` for icon-only controls
- `ScrollArea` for drawer and list content

The app should favor dense but calm tool surfaces over marketing-style cards. Cards are appropriate for repeated records, modals, and framed tools, but page sections and workspace regions should be structured as panes, drawers, tabs, and sheets.

## Frontend Architecture Convention

Open Edda frontend code should use Domain-Oriented Vertical Slices.

Code is organized by product domains and workflows, not by technical file type. A feature owns its UI components, Redux slice, hooks/selectors, feature-specific API adapters, and local types.

Baseline structure:

```txt
frontend/src/
  app/
    router/
    store/
    providers/
  features/
    auth/
    projects/
    workspace/
    editor/
    assistant/
    review/
    notes/
    model-settings/
    skills/
  shared/
    ui/
    api/
    lib/
    types/
```

Rules:

- `app/` wires application providers, router, and store.
- `shared/ui` wraps shadcn primitives and generic layout controls.
- `shared/api`, `shared/lib`, and `shared/types` hold cross-feature utilities only.
- `features/*` may depend on `shared/*`.
- Feature internals should not be imported deeply by other features.
- Cross-feature usage should go through a feature public module or index.
- Route components compose features and should not contain business logic.
- Redux slices live near their owning feature unless they are truly app-wide.
- Avoid root-level dumping grounds such as generic `components/`, `hooks/`, or `types/`.

## State Management

Milestone 4 should add Redux Toolkit as the main app-state layer. Redux is familiar for this project and fits the amount of shared workspace state.

Suggested slices:

- `workspaceSlice`: active mode, drawer state, panel widths, active drawer tabs, mobile sheet state.
- `editorSlice`: selected content context, cursor state, selection byte offsets, editor-local action modal state.
- `projectsSlice`: selected project/content IDs can be derived from routes, but project/content entities and loading state can live here if not handled by a server-cache library.
- `assistantSlice`: active session, chat state, action status/result visibility.
- `modelSettingsSlice`: selected provider/model and disclosure state.
- `skillsSlice`: installed skills, selected skill chips, skill browser state.

Form-local React state remains acceptable for ephemeral input drafts inside forms and modals. If React Query or a similar library is added later, it should own server cache only, not workspace UI state.

## Persistence

Persistence is hybrid and local-browser based. There is no per-user persistence model in scope.

Project-scoped persisted state:

- active workspace mode,
- drawer widths,
- last active right-drawer tab,
- last content kind,
- fallback selected content for `/projects/:projectId` when the route does not include `contentId`.

Browser-scoped persisted state:

- general UI density preference,
- mobile drawer preference,
- preferred default mode for new projects,
- non-project-specific panel habits.

Route state takes precedence for `projectId` and `contentId`. Persisted selected content must not override an explicit deep link.

## Error And Empty States

- Auth or API failures should show recoverable route-level states.
- The editor should remain usable when provider/model configuration is missing; only AI actions should be disabled.
- Generate, Rewrite, and Check should validate cursor/selection and revision before submit.
- Revision conflicts should surface near the editor and offer a clear next step such as reload current text or review conflict.
- Missing selection should disable selection-scoped actions and explain why.
- Mobile sheets must preserve typed instructions while temporarily closed.

## Testing Strategy

Frontend tests should cover structure and state before visual polish.

Required coverage:

- Redux reducer and selector tests for workspace modes, drawer persistence, and action modal state.
- Persistence tests for project-scoped and browser-scoped defaults.
- Component tests for:
  - mode switching,
  - desktop drawer behavior,
  - mobile sheet behavior,
  - selection action availability,
  - Generate composer validation,
  - Rewrite/Check modal preview and instruction handling.
- Browser smoke tests once the workspace shell exists:
  - desktop Assistant mode,
  - desktop Draft mode with max-width editor,
  - mobile Draft mode with sheets,
  - deep link to `/projects/:projectId/content/:contentKind/:contentId`.

Backend integration tests remain in Go. Milestone 4 may add backend endpoints only where required for daily writing polish, such as revision restore or attached note workflows.

## Implementation Scope Guidance

Milestone 4 should be implemented in stages:

1. Routed workspace foundation: add routing, Redux store, Tailwind v4, shadcn/ui setup, vertical-slice structure, workspace shell, editor frame, Generate composer, selection action bubble, mobile toolbar, and Rewrite/Check modal shell.
2. Behavior parity and data slices: move existing assistant, model settings, skill, activity, prompt-record, and script-runtime visibility from the old monolithic frontend into routed vertical slices.
3. Editor adapter: integrate Galley Editor or a staged editor adapter behind `EditorFrame`, with mutation-safe cursor and UTF-8 byte selection APIs.
4. Assistant actions: wire Generate, Rewrite, Check, preview, accept/reject, and revision-safe conflict handling from the editor-local controls.
5. Review surfaces: add review-mode drawer surfaces for chat/tools-first analysis, revisions, diffs, restore, attached notes, activity, and prompt records.
6. Mobile and browser smoke hardening: add focused component tests and browser smoke coverage for desktop Assistant/Draft/Review and mobile sheet workflows.

Avoid unrelated redesign work. Visual polish beyond clear layout, spacing, accessibility, and shadcn/Tailwind consistency should wait until the structural workspace is stable.
