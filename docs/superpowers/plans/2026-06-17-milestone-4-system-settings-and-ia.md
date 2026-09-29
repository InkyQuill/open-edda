# Milestone 4 System Settings And IA Correction Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move provider/model and skill administration out of the writing workspace, add a dedicated system settings route, make the assistant drawer chat-only, and add first-class project/content creation controls.

**Architecture:** Keep the existing Milestone 4 vertical-slice structure. Add a `settings` feature that composes the already-existing model, skill, and script runtime panels, then simplify workspace drawers so they only expose authoring surfaces. Add creation API wrappers and route-aware UI handlers in the existing `projects` and `workspace` slices rather than introducing a new server-cache layer.

**Tech Stack:** React 19, React Router 7, Redux Toolkit, TypeScript, Vitest, Tailwind v4, shadcn-style local UI primitives, existing Go HTTP APIs.

---

## Current Status Check

- Milestone 4 Phase 3 is implemented in `docs/superpowers/plans/2026-06-17-milestone-4-editor-adapter.md`.
- `docs/roadmap.md` marks Phase 3.5 as Planned and says Phase 4 should be created after Phase 3.5 is implemented.
- `frontend/src/app/router/routes.tsx` has no `/settings` route.
- `frontend/src/features/workspace/WorkspaceShell.tsx` still renders `ModelSettingsPanel` and `ScriptRuntimePanel` inside workspace drawers and exposes a mobile `model` sheet.
- `frontend/src/features/assistant/AssistantDrawer.tsx` is mostly chat already, but still includes a staged "Quick actions" section; Phase 3.5 should keep model disclosure only and remove non-chat/admin surfaces from the assistant drawer.
- `frontend/src/features/projects/ProjectsPage.tsx` is still a simple list and lacks settings, project creation, and import affordances.
- `frontend/src/api.ts` has `listProjects`, `listContent`, and `updateContent`, but lacks wrappers for existing `POST /api/projects`, `POST /api/projects/{projectID}/content`, and `POST /api/projects/import/elysium`.

## File Structure

- Modify: `frontend/src/api.ts`
  - Add typed wrappers for project creation, Elysium import, and content creation.
- Modify: `frontend/src/types.ts`
  - Add optional `updatedAt` only if backend returns it during this task; otherwise keep the current project type unchanged and show language/slug metadata.
- Modify: `frontend/src/app/router/routes.tsx`
  - Add authenticated `/settings` and project settings placeholder routes.
- Create: `frontend/src/features/settings/SettingsPage.tsx`
  - Compose system status, provider/model administration, and a project-scoped skill administration bridge.
- Create: `frontend/src/features/settings/settingsSlice.ts`
  - Store selected settings project ID for project-scoped skill/script panels.
- Create: `frontend/src/features/settings/settingsSlice.test.ts`
  - Test settings project selection and reset behavior.
- Modify: `frontend/src/app/store/store.ts`
  - Register `settingsReducer`.
- Modify: `frontend/src/features/workspace/workspaceSlice.ts`
  - Remove `model` as a drawer/mobile sheet target and keep assistant/review/content surfaces only.
- Modify: `frontend/src/features/workspace/workspaceSlice.test.ts`
  - Update mobile sheet and active right drawer expectations.
- Modify: `frontend/src/features/workspace/WorkspaceShell.tsx`
  - Remove provider/model/skill admin panels from workspace, add settings navigation in chrome, add content creation control plumbing.
- Modify: `frontend/src/features/workspace/WorkspacePage.tsx`
  - Call `createContent`, update local content list, select the new item, and navigate to its canonical route.
- Modify: `frontend/src/features/notes/ContextDrawer.tsx`
  - Add creation controls for each content kind tab.
- Modify: `frontend/src/features/assistant/AssistantDrawer.tsx`
  - Remove the staged quick-action/admin section; keep transcript, message composer, new chat, short empty states, and `ModelStatus`.
- Modify: `frontend/src/features/projects/ProjectsPage.tsx`
  - Redesign `/projects` as a dashboard with settings, logout, project creation, import, project cards, and empty-state actions.
- Modify: `frontend/src/styles.css`
  - Remove old `.project-row` dependence where Tailwind classes replace it; keep only layout helpers still used by auth/editor.
- Modify: `docs/roadmap.md`
  - Mark Phase 3.5 Implemented only after tests/build pass.

## Task 1: Add Creation API Wrappers

**Files:**
- Modify: `frontend/src/api.ts`

- [ ] **Step 1: Add input types and wrappers**

Add these exports to `frontend/src/api.ts` after `listProjects` and before `listContent`:

```ts
export type CreateProjectInput = {
  title: string;
  language: string;
};

export async function createProject(input: CreateProjectInput): Promise<StoryProject> {
  const response = await authFetch("/api/projects", {
    method: "POST",
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw new Error(`create project failed: ${response.status}`);
  }
  return response.json() as Promise<StoryProject>;
}

export async function importElysiumProject(file: File): Promise<StoryProject> {
  const response = await authFetch("/api/projects/import/elysium", {
    method: "POST",
    body: file,
    headers: new Headers({ "Content-Type": "application/zip" }),
  });
  if (!response.ok) {
    throw new Error(`import Elysium project failed: ${response.status}`);
  }
  return response.json() as Promise<StoryProject>;
}

export type CreateContentInput = {
  kind: ContentKind;
  title: string;
  bodyMarkdown: string;
  metadataJson: string;
  sortOrder: number;
  reason: string;
};

export async function createContent(
  projectId: string,
  input: CreateContentInput,
  signal?: AbortSignal,
): Promise<ContentItem> {
  const response = await authFetch(`/api/projects/${encodeURIComponent(projectId)}/content`, {
    method: "POST",
    body: JSON.stringify(input),
    signal,
  });
  if (!response.ok) {
    throw new Error(`create content failed: ${response.status}`);
  }
  return response.json() as Promise<ContentItem>;
}
```

- [ ] **Step 2: Run typecheck**

Run: `cd frontend && bun run typecheck`

Expected: PASS. If it fails because `File` is not available in the configured TypeScript libs, use `Blob` instead:

```ts
export async function importElysiumProject(file: Blob): Promise<StoryProject> {
  const response = await authFetch("/api/projects/import/elysium", {
    method: "POST",
    body: file,
    headers: new Headers({ "Content-Type": "application/zip" }),
  });
  if (!response.ok) {
    throw new Error(`import Elysium project failed: ${response.status}`);
  }
  return response.json() as Promise<StoryProject>;
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/api.ts
git commit -m "feat: add project and content creation API wrappers"
```

## Task 2: Add Settings Route And State

**Files:**
- Create: `frontend/src/features/settings/settingsSlice.ts`
- Create: `frontend/src/features/settings/settingsSlice.test.ts`
- Modify: `frontend/src/app/store/store.ts`
- Create: `frontend/src/features/settings/SettingsPage.tsx`
- Modify: `frontend/src/app/router/routes.tsx`

- [ ] **Step 1: Write settings reducer test**

Create `frontend/src/features/settings/settingsSlice.test.ts`:

```ts
import { describe, expect, it } from "vitest";
import { initialSettingsState, settingsActions, settingsReducer } from "./settingsSlice";

describe("settingsSlice", () => {
  it("selects a project for project-scoped administration panels", () => {
    const state = settingsReducer(initialSettingsState, settingsActions.setSelectedProjectId("project-1"));

    expect(state.selectedProjectId).toBe("project-1");
  });

  it("clears selected project when set to null", () => {
    const state = settingsReducer(
      { ...initialSettingsState, selectedProjectId: "project-1" },
      settingsActions.setSelectedProjectId(null),
    );

    expect(state.selectedProjectId).toBeNull();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && bun run test -- settingsSlice`

Expected: FAIL with a module resolution error for `./settingsSlice`.

- [ ] **Step 3: Implement settings slice**

Create `frontend/src/features/settings/settingsSlice.ts`:

```ts
import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export type SettingsState = {
  selectedProjectId: string | null;
};

export const initialSettingsState: SettingsState = {
  selectedProjectId: null,
};

const settingsSlice = createSlice({
  name: "settings",
  initialState: initialSettingsState,
  reducers: {
    setSelectedProjectId(state, action: PayloadAction<string | null>) {
      state.selectedProjectId = action.payload;
    },
  },
});

export const settingsActions = settingsSlice.actions;
export const settingsReducer = settingsSlice.reducer;
```

- [ ] **Step 4: Register reducer**

Modify `frontend/src/app/store/store.ts` so the reducer map includes settings:

```ts
import { settingsReducer } from "../../features/settings/settingsSlice";

export const store = configureStore({
  reducer: {
    assistant: assistantReducer,
    editor: editorReducer,
    modelSettings: modelSettingsReducer,
    review: reviewReducer,
    scriptRuntime: scriptRuntimeReducer,
    settings: settingsReducer,
    skills: skillsReducer,
    workspace: workspaceReducer,
  },
});
```

- [ ] **Step 5: Create settings page**

Create `frontend/src/features/settings/SettingsPage.tsx`:

```tsx
import { ArrowLeft, CheckCircle2, CircleAlert, FolderCog } from "lucide-react";
import { useEffect, useMemo } from "react";
import { Link } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";

import type { AppDispatch, RootState } from "../../app/store/store";
import { Button } from "../../shared/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../shared/ui/tabs";
import { listProjects } from "../../api";
import { ModelSettingsPanel } from "../model-settings/ModelSettingsPanel";
import { loadProviderConfigs } from "../model-settings/modelSettingsThunks";
import { ScriptRuntimePanel } from "../script-runtime/ScriptRuntimePanel";
import { loadSkills } from "../skills/skillsThunks";
import { SkillsPanel } from "../skills/SkillsPanel";
import { settingsActions } from "./settingsSlice";
import { useState } from "react";
import type { StoryProject } from "../../types";

export function SettingsPage() {
  const dispatch = useDispatch<AppDispatch>();
  const selectedProjectId = useSelector((state: RootState) => state.settings.selectedProjectId);
  const providers = useSelector((state: RootState) => state.modelSettings.providers);
  const activeModelVariantId = useSelector((state: RootState) => state.modelSettings.activeModelVariantId);
  const [projects, setProjects] = useState<StoryProject[]>([]);
  const [projectsError, setProjectsError] = useState<string | null>(null);
  const selectedProject = useMemo(
    () => projects.find((project) => project.id === selectedProjectId) ?? projects[0] ?? null,
    [projects, selectedProjectId],
  );

  useEffect(() => {
    void dispatch(loadProviderConfigs());
    void listProjects()
      .then((items) => {
        setProjects(items);
        if (!selectedProjectId && items[0]) {
          dispatch(settingsActions.setSelectedProjectId(items[0].id));
        }
      })
      .catch((cause: unknown) => {
        setProjectsError(cause instanceof Error ? cause.message : "Could not load projects");
      });
  }, [dispatch, selectedProjectId]);

  useEffect(() => {
    if (selectedProject?.id) {
      void dispatch(loadSkills({ projectId: selectedProject.id }));
    }
  }, [dispatch, selectedProject?.id]);

  const systemReady = providers.length > 0 && Boolean(activeModelVariantId);

  return (
    <main className="min-h-dvh bg-background text-foreground">
      <header className="flex items-center justify-between gap-3 border-b border-border px-4 py-3">
        <div className="min-w-0">
          <p className="text-xs text-muted-foreground">System</p>
          <h1 className="truncate text-base font-semibold">Settings</h1>
        </div>
        <Button asChild type="button" variant="outline">
          <Link to="/projects">
            <ArrowLeft data-icon="inline-start" aria-hidden="true" />
            Projects
          </Link>
        </Button>
      </header>

      <section className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6">
        <div className="grid gap-3 rounded-md border border-border bg-background p-4 md:grid-cols-[1fr_auto]">
          <div>
            <h2 className="text-base font-semibold">System status</h2>
            <p className="text-sm text-muted-foreground">
              Provider credentials and model catalog are system-level resources.
            </p>
          </div>
          <p className="flex items-center gap-2 text-sm">
            {systemReady ? <CheckCircle2 className="size-4 text-green-700" /> : <CircleAlert className="size-4 text-destructive" />}
            {systemReady ? "Provider and model ready" : "Configure at least one provider and model"}
          </p>
        </div>

        <Tabs defaultValue="models">
          <TabsList>
            <TabsTrigger value="models">Providers and models</TabsTrigger>
            <TabsTrigger value="skills">Skills and scripts</TabsTrigger>
          </TabsList>

          <TabsContent value="models">
            <div className="rounded-md border border-border bg-background p-4">
              <ModelSettingsPanel />
            </div>
          </TabsContent>

          <TabsContent value="skills">
            <div className="grid gap-4 lg:grid-cols-[280px_1fr]">
              <aside className="rounded-md border border-border bg-background p-4">
                <h2 className="mb-3 flex items-center gap-2 text-sm font-medium">
                  <FolderCog className="size-4" aria-hidden="true" />
                  Project scope
                </h2>
                {projectsError ? <p role="alert" className="text-sm text-destructive">{projectsError}</p> : null}
                <div className="grid gap-2">
                  {projects.map((project) => (
                    <Button
                      key={project.id}
                      type="button"
                      variant={project.id === selectedProject?.id ? "secondary" : "outline"}
                      className="justify-start"
                      onClick={() => dispatch(settingsActions.setSelectedProjectId(project.id))}
                    >
                      {project.title}
                    </Button>
                  ))}
                </div>
              </aside>
              <section className="grid gap-4 rounded-md border border-border bg-background p-4">
                {selectedProject ? (
                  <>
                    <SkillsPanel projectId={selectedProject.id} sessionId={null} />
                    <ScriptRuntimePanel projectId={selectedProject.id} />
                  </>
                ) : (
                  <p className="text-sm text-muted-foreground">Create a project before managing project-scoped skills.</p>
                )}
              </section>
            </div>
          </TabsContent>
        </Tabs>
      </section>
    </main>
  );
}
```

- [ ] **Step 6: Add route**

Modify `frontend/src/app/router/routes.tsx`:

```tsx
import { SettingsPage } from "../../features/settings/SettingsPage";

// inside <Route element={<RequireAuth />}>
<Route path="/settings" element={<SettingsPage />} />
```

- [ ] **Step 7: Run focused tests**

Run: `cd frontend && bun run test -- settingsSlice modelSettingsSlice skillsSlice scriptRuntimeSlice`

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add frontend/src/app/router/routes.tsx frontend/src/app/store/store.ts frontend/src/features/settings
git commit -m "feat: add system settings route"
```

## Task 3: Remove Admin Surfaces From Workspace

**Files:**
- Modify: `frontend/src/features/workspace/workspaceSlice.ts`
- Modify: `frontend/src/features/workspace/workspaceSlice.test.ts`
- Modify: `frontend/src/features/workspace/WorkspaceShell.tsx`
- Modify: `frontend/src/features/assistant/AssistantDrawer.tsx`

- [ ] **Step 1: Update workspace slice test**

In `frontend/src/features/workspace/workspaceSlice.test.ts`, replace expectations that mention `model` with these assertions:

```ts
it("opens only authoring mobile sheets", () => {
  const state = workspaceReducer(initialWorkspaceState, workspaceActions.setMobileSheet("assistant"));
  expect(state.mobileSheet).toBe("assistant");
});

it("review mode opens the review tools drawer", () => {
  const state = workspaceReducer(initialWorkspaceState, workspaceActions.setMode("review"));
  expect(state.rightDrawerOpen).toBe(true);
  expect(state.activeRightTab).toBe("tools");
});
```

- [ ] **Step 2: Run test to verify it fails until types change**

Run: `cd frontend && bun run test -- workspaceSlice`

Expected: FAIL if the stale tests still reference `model`, or PASS if no stale expectations exist.

- [ ] **Step 3: Remove model sheet and drawer type**

Change these type definitions in `frontend/src/features/workspace/workspaceSlice.ts`:

```ts
export type DrawerTab =
  | "contents"
  | "world"
  | "notes"
  | "chat"
  | "tools"
  | "revisions";
export type MobileSheet =
  | null
  | "contents"
  | "assistant"
  | "review"
  | "world-notes";
```

- [ ] **Step 4: Simplify workspace shell**

In `frontend/src/features/workspace/WorkspaceShell.tsx`:

- remove imports for `Settings2`, `ModelSettingsPanel`, and `ScriptRuntimePanel`;
- add `Link` from `react-router-dom`;
- remove `ModelToolsPanel`;
- remove the mobile `{ sheet: "model", label: "Model", icon: Settings2 }` button;
- make `rightDrawer` choose only review or assistant:

```tsx
const rightDrawer =
  workspace.mode === "review" || workspace.activeRightTab === "tools" || workspace.activeRightTab === "revisions" ? (
    <ReviewDrawer projectId={projectId} />
  ) : (
    <AssistantDrawer projectId={projectId} />
  );
```

- add a settings button to the desktop header actions:

```tsx
<Button asChild type="button" variant="outline">
  <Link to="/settings">
    <Settings2 data-icon="inline-start" aria-hidden="true" />
    Settings
  </Link>
</Button>
```

Keep the `Settings2` import if using this button; otherwise use another lucide settings icon already imported.

- update `mobileSheetTitle` and `renderMobileSheet` so neither handles `model`.

- [ ] **Step 5: Make assistant drawer chat-only**

Remove this whole section from `frontend/src/features/assistant/AssistantDrawer.tsx`:

```tsx
<section className="flex flex-col gap-3" aria-labelledby="quick-actions-title">
  <div className="flex items-center justify-between gap-3">
    <h3 id="quick-actions-title" className="text-sm font-medium text-foreground">
      Quick actions
    </h3>
    <Button type="button" variant="outline" size="xs">
      <WandSparkles />
      Ready
    </Button>
  </div>
  <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
    Selection actions will attach to the editor in the next task.
  </p>
</section>
```

Also remove `WandSparkles` from the lucide import.

- [ ] **Step 6: Run focused tests and typecheck**

Run: `cd frontend && bun run test -- workspaceSlice assistantSlice && bun run typecheck`

Expected: PASS. Typecheck should prove no workspace code still refers to `MobileSheet` value `"model"` or `DrawerTab` value `"model"`.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/features/workspace frontend/src/features/assistant/AssistantDrawer.tsx
git commit -m "refactor: keep workspace drawers focused on authoring"
```

## Task 4: Add Content Creation In Workspace

**Files:**
- Modify: `frontend/src/features/notes/ContextDrawer.tsx`
- Modify: `frontend/src/features/workspace/WorkspacePage.tsx`
- Modify: `frontend/src/features/workspace/WorkspaceShell.tsx`

- [ ] **Step 1: Extend context drawer props**

Modify the `ContextDrawer` props in `frontend/src/features/notes/ContextDrawer.tsx` to include:

```ts
onCreateContent: (kind: ContentKind) => void;
contentCreating: boolean;
```

Add a create button near the content-kind controls:

```tsx
<Button
  type="button"
  variant="outline"
  size="xs"
  disabled={contentCreating}
  onClick={() => onCreateContent(activeContentKind)}
>
  <Plus data-icon="inline-start" aria-hidden="true" />
  New
</Button>
```

Import `Plus` from `lucide-react` if it is not already imported.

- [ ] **Step 2: Add workspace creation state and handler**

In `frontend/src/features/workspace/WorkspacePage.tsx`, import `createContent`:

```ts
import { createContent, listContent, listProjects } from "../../api";
```

Add state:

```ts
const [contentCreating, setContentCreating] = useState(false);
```

Add handler before `handleSelectContent`:

```ts
function contentKindLabel(kind: ContentKind): string {
  switch (kind) {
    case "chapter":
      return "Chapter";
    case "story_bible_entry":
      return "Story bible entry";
    case "writing_brief":
      return "Writing brief";
    case "project_note":
      return "Project note";
  }
}

function defaultBodyForKind(kind: ContentKind): string {
  switch (kind) {
    case "chapter":
      return "";
    case "story_bible_entry":
      return "## Overview\n\n";
    case "writing_brief":
      return "## Brief\n\n";
    case "project_note":
      return "## Note\n\n";
  }
}

function handleCreateContent(kind: ContentKind): void {
  if (!projectId || contentCreating) return;
  const nextNumber = contentItems.filter((item) => item.kind === kind).length + 1;
  setContentCreating(true);
  setContentError(null);

  void createContent(projectId, {
    kind,
    title: `${contentKindLabel(kind)} ${nextNumber}`,
    bodyMarkdown: defaultBodyForKind(kind),
    metadataJson: "{}",
    sortOrder: contentItems.length + 1,
    reason: "created from workspace",
  })
    .then((item) => {
      setContentItems((items) => [...items, item]);
      dispatch(workspaceActions.setLastContentKind(item.kind));
      dispatch(workspaceActions.setFallbackContentId(item.id));
      navigate(`/projects/${encodeURIComponent(projectId)}/content/${encodeURIComponent(item.kind)}/${encodeURIComponent(item.id)}`);
    })
    .catch((cause: unknown) => {
      setContentError(cause instanceof Error ? cause.message : "Could not create content");
    })
    .finally(() => setContentCreating(false));
}
```

Pass `contentCreating` and `onCreateContent={handleCreateContent}` into `WorkspaceShell`.

- [ ] **Step 3: Pass creation props through shell**

In `frontend/src/features/workspace/WorkspaceShell.tsx`, extend props:

```ts
contentCreating: boolean;
onCreateContent: (kind: ContentKind) => void;
```

Pass them to every `ContextDrawer` instance:

```tsx
contentCreating={contentCreating}
onCreateContent={onCreateContent}
```

- [ ] **Step 4: Run typecheck**

Run: `cd frontend && bun run typecheck`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features/notes/ContextDrawer.tsx frontend/src/features/workspace/WorkspacePage.tsx frontend/src/features/workspace/WorkspaceShell.tsx
git commit -m "feat: add workspace content creation controls"
```

## Task 5: Redesign Projects Page

**Files:**
- Modify: `frontend/src/features/projects/ProjectsPage.tsx`
- Modify: `frontend/src/styles.css`

- [ ] **Step 1: Add project creation UI**

Replace `ProjectsPage` with a dashboard implementation that imports `createProject` and `importElysiumProject`, tracks `title`, `language`, `createError`, and `creating`, and uses this create handler:

```ts
function handleCreateProject(event: React.FormEvent<HTMLFormElement>): void {
  event.preventDefault();
  const trimmedTitle = title.trim();
  if (!trimmedTitle || creating) return;

  setCreating(true);
  setCreateError(null);
  void createProject({ title: trimmedTitle, language: language.trim() || "en" })
    .then((project) => navigate(`/projects/${encodeURIComponent(project.id)}`))
    .catch((cause: unknown) => {
      setCreateError(cause instanceof Error ? cause.message : "Could not create project");
    })
    .finally(() => setCreating(false));
}
```

Use this import handler for a file input:

```ts
function handleImportProject(event: React.ChangeEvent<HTMLInputElement>): void {
  const file = event.target.files?.[0];
  if (!file) return;

  setCreating(true);
  setCreateError(null);
  void importElysiumProject(file)
    .then((project) => navigate(`/projects/${encodeURIComponent(project.id)}`))
    .catch((cause: unknown) => {
      setCreateError(cause instanceof Error ? cause.message : "Could not import project");
    })
    .finally(() => {
      setCreating(false);
      event.target.value = "";
    });
}
```

The page must include:

```tsx
<Button asChild type="button" variant="outline">
  <Link to="/settings">
    <Settings2 data-icon="inline-start" aria-hidden="true" />
    Settings
  </Link>
</Button>
```

Each project card link must still navigate to:

```tsx
`/projects/${encodeURIComponent(project.id)}`
```

- [ ] **Step 2: Remove obsolete project row CSS only if unused**

After replacing `ProjectsPage`, run:

```bash
rg "project-row|project-list|project-dashboard" frontend/src
```

If no references remain, delete `.project-dashboard`, `.project-list`, `.project-row`, `.project-row[data-active="true"]`, `.project-row-title`, and `.project-row-meta` from `frontend/src/styles.css`.

- [ ] **Step 3: Run typecheck**

Run: `cd frontend && bun run typecheck`

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/projects/ProjectsPage.tsx frontend/src/styles.css
git commit -m "feat: redesign projects dashboard"
```

## Task 6: Verify And Update Roadmap

**Files:**
- Modify: `docs/roadmap.md`

- [ ] **Step 1: Run frontend test suite**

Run: `cd frontend && bun run test`

Expected: PASS.

- [ ] **Step 2: Run frontend build**

Run: `cd frontend && bun run build`

Expected: PASS. A bundle-size warning is acceptable if it matches the existing Galley warning from Phase 3.

- [ ] **Step 3: Run Go tests**

Run: `go test -tags sqlite_fts5 ./...`

Expected: PASS.

- [ ] **Step 4: Run diff hygiene**

Run: `git diff --check`

Expected: no output.

- [ ] **Step 5: Update roadmap**

In `docs/roadmap.md`, change the Phase 3.5 status row from:

```md
| 3.5 | Information architecture correction: move provider/model/skill administration to settings, make assistant drawer chat-only, add project/content creation controls, and redesign the projects page | Planned | `docs/superpowers/plans/2026-06-17-milestone-4-system-settings-and-ia.md` |
```

to:

```md
| 3.5 | Information architecture correction: move provider/model/skill administration to settings, make assistant drawer chat-only, add project/content creation controls, and redesign the projects page | Implemented | `docs/superpowers/plans/2026-06-17-milestone-4-system-settings-and-ia.md` |
```

- [ ] **Step 6: Commit**

```bash
git add docs/roadmap.md
git commit -m "docs: mark milestone 4 ia correction implemented"
```

## Self-Review

- Spec coverage: The plan covers system settings, project settings boundary by using project-scoped skill panels inside system settings, assistant drawer cleanup, content creation controls, projects page redesign, tests/build, and roadmap update.
- Placeholder scan: No `TBD`, `TODO`, "implement later", or "similar to" instructions remain.
- Type consistency: `ContentKind`, `MobileSheet`, `DrawerTab`, `CreateContentInput`, and `StoryProject` names match current source files.

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-06-17-milestone-4-system-settings-and-ia.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
