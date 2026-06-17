# Milestone 4 Workspace Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Replace the single-screen frontend entry with a routed, Redux-backed writing workspace foundation that supports project/content deep links, Draft/Assistant/Review presets, responsive drawers/sheets, and editor-local action shells.

**Architecture:** Build the Milestone 4 foundation as a first vertical slice without rewriting all agent/settings/skill functionality at once. `app/` owns providers, router, and store; `features/*` owns domain UI/state; existing API modules remain usable and can be moved into feature slices in later plans. Route params own durable project/content identity, while Redux owns workspace mode, drawer/sheet state, and editor action state.

**Tech Stack:** React + Vite, React Router, Redux Toolkit, React Redux, Tailwind v4 via `@tailwindcss/vite`, shadcn/ui installed through the shadcn CLI, TypeScript, Vitest.

---

## File Structure

Create the Milestone 4 frontend structure:

```txt
frontend/src/
  app/
    App.tsx
    providers/AppProviders.tsx
    router/RequireAuth.tsx
    router/routes.tsx
    store/store.ts
  features/
    auth/AuthPage.tsx
    projects/ProjectsPage.tsx
    projects/projectHooks.ts
    workspace/WorkspacePage.tsx
    workspace/WorkspaceShell.tsx
    workspace/workspacePersistence.ts
    workspace/workspaceSelectors.ts
    workspace/workspaceSlice.ts
    workspace/workspaceSlice.test.ts
    editor/EditorFrame.tsx
    editor/GenerateComposer.tsx
    editor/SelectionActionDialog.tsx
    editor/editorSlice.ts
    editor/editorSlice.test.ts
    assistant/AssistantDrawer.tsx
    review/ReviewDrawer.tsx
    notes/ContextDrawer.tsx
    model-settings/ModelStatus.tsx
    skills/SkillChipsPanel.tsx
  shared/
    api/authFetch.ts
    lib/byteOffsets.ts
    lib/utils.ts
    ui/button.tsx
    ui/dialog.tsx
    ui/input.tsx
    ui/sheet.tsx
    ui/tabs.tsx
    ui/textarea.tsx
```

Keep these existing files during this plan:

- `frontend/src/agentApi.ts`
- `frontend/src/agentTypes.ts`
- `frontend/src/api.ts`
- `frontend/src/authApi.ts`
- `frontend/src/scriptRuntimeApi.ts`
- `frontend/src/scriptRuntimeTypes.ts`
- `frontend/src/skillApi.ts`
- `frontend/src/skillTypes.ts`
- `frontend/src/types.ts`

`frontend/src/App.tsx` should become a compatibility shim that re-exports `App` from `frontend/src/app/App.tsx` or be updated so existing imports still compile.

---

## Task 1: Add Frontend Foundation Dependencies

**Files:**
- Modify: `frontend/package.json`
- Modify: `frontend/vite.config.ts`
- Modify: `frontend/src/styles.css`

- [x] **Step 1: Update `frontend/package.json` dependencies**

Add runtime dependencies:

```json
{
  "@reduxjs/toolkit": "latest",
  "@tailwindcss/vite": "latest",
  "clsx": "latest",
  "lucide-react": "latest",
  "react-redux": "latest",
  "react-router-dom": "latest",
  "tailwind-merge": "latest",
  "tailwindcss": "latest"
}
```

Add dev dependencies:

```json
{
  "vitest": "latest"
}
```

Add scripts:

```json
{
  "test": "vitest run",
  "test:watch": "vitest"
}
```

- [x] **Step 2: Install dependencies**

Run:

```bash
cd frontend
bun install
```

Expected: `bun.lock` updates and install exits 0. If network sandboxing blocks the install, rerun with approved escalation.

- [x] **Step 3: Add Tailwind v4 Vite plugin**

Update `frontend/vite.config.ts`. Use `vitest/config` so the `test` key type-checks:

```ts
import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
  test: {
    environment: "node",
    globals: true,
  },
});
```

- [x] **Step 4: Convert global CSS entry to Tailwind v4**

At the top of `frontend/src/styles.css`, add the Tailwind v4 CSS-first import:

```css
@import "tailwindcss";
```

Do not add `tailwind.config.js`, `@tailwind` directives, PostCSS-only Tailwind setup, `autoprefixer`, or `postcss-import` for this Vite app.

- [x] **Step 5: Verify baseline build still starts TypeScript**

Run:

```bash
cd frontend
bun run typecheck
```

Expected at this stage: either PASS or only missing-module errors caused by files not created yet. Do not proceed with implementation if package install failed.

---

## Task 2: Add App Providers, Store, And Routes

**Files:**
- Create: `frontend/src/app/store/store.ts`
- Create: `frontend/src/app/providers/AppProviders.tsx`
- Create: `frontend/src/app/router/RequireAuth.tsx`
- Create: `frontend/src/app/router/routes.tsx`
- Create: `frontend/src/app/App.tsx`
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/main.tsx`

- [x] **Step 1: Create Redux store**

Create `frontend/src/app/store/store.ts`:

```ts
import { configureStore } from "@reduxjs/toolkit";
import { editorReducer } from "../../features/editor/editorSlice";
import { workspaceReducer } from "../../features/workspace/workspaceSlice";

export const store = configureStore({
  reducer: {
    editor: editorReducer,
    workspace: workspaceReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

- [x] **Step 2: Create app providers**

Create `frontend/src/app/providers/AppProviders.tsx`:

```tsx
import type { ReactNode } from "react";
import { Provider } from "react-redux";
import { BrowserRouter } from "react-router-dom";
import { store } from "../store/store";

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <Provider store={store}>
      <BrowserRouter>{children}</BrowserRouter>
    </Provider>
  );
}
```

- [x] **Step 3: Create auth guard**

Create `frontend/src/app/router/RequireAuth.tsx`:

```tsx
import { Navigate, Outlet, useLocation } from "react-router-dom";
import { getToken } from "../../authApi";

export function RequireAuth() {
  const location = useLocation();
  if (!getToken()) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }
  return <Outlet />;
}
```

- [x] **Step 4: Create route table**

Create `frontend/src/app/router/routes.tsx`:

```tsx
import { Navigate, Route, Routes } from "react-router-dom";
import { AuthPage } from "../../features/auth/AuthPage";
import { ProjectsPage } from "../../features/projects/ProjectsPage";
import { WorkspacePage } from "../../features/workspace/WorkspacePage";
import { RequireAuth } from "./RequireAuth";

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<AuthPage />} />
      <Route element={<RequireAuth />}>
        <Route path="/projects" element={<ProjectsPage />} />
        <Route path="/projects/:projectId" element={<WorkspacePage />} />
        <Route path="/projects/:projectId/content/:contentKind/:contentId" element={<WorkspacePage />} />
      </Route>
      <Route path="/" element={<Navigate to="/projects" replace />} />
      <Route path="*" element={<Navigate to="/projects" replace />} />
    </Routes>
  );
}
```

- [x] **Step 5: Create app shell entry**

Create `frontend/src/app/App.tsx`:

```tsx
import { AppRoutes } from "./router/routes";

export function App() {
  return <AppRoutes />;
}
```

- [x] **Step 6: Keep root import compatibility**

Replace `frontend/src/App.tsx` with:

```ts
export { App } from "./app/App";
```

- [x] **Step 7: Wrap root render**

Update `frontend/src/main.tsx`:

```tsx
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";
import { AppProviders } from "./app/providers/AppProviders";
import "./styles.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AppProviders>
      <App />
    </AppProviders>
  </StrictMode>,
);
```

---

## Task 3: Add Workspace And Editor Redux State

**Files:**
- Create: `frontend/src/features/workspace/workspaceSlice.ts`
- Create: `frontend/src/features/workspace/workspaceSelectors.ts`
- Create: `frontend/src/features/workspace/workspacePersistence.ts`
- Create: `frontend/src/features/editor/editorSlice.ts`
- Create: `frontend/src/features/workspace/workspaceSlice.test.ts`
- Create: `frontend/src/features/editor/editorSlice.test.ts`

- [x] **Step 1: Create workspace persistence helpers**

Create `frontend/src/features/workspace/workspacePersistence.ts`:

```ts
import type { WorkspaceMode, WorkspaceProjectState } from "./workspaceSlice";

const browserKey = "open_edda_workspace_browser";
const projectPrefix = "open_edda_workspace_project:";

export type BrowserWorkspaceState = {
  defaultMode: WorkspaceMode;
  density: "comfortable" | "compact";
};

export function loadBrowserWorkspaceState(storage: Storage = window.localStorage): BrowserWorkspaceState {
  const raw = storage.getItem(browserKey);
  if (!raw) return { defaultMode: "assistant", density: "comfortable" };
  try {
    const parsed = JSON.parse(raw) as Partial<BrowserWorkspaceState>;
    return {
      defaultMode: parsed.defaultMode === "draft" || parsed.defaultMode === "review" ? parsed.defaultMode : "assistant",
      density: parsed.density === "compact" ? "compact" : "comfortable",
    };
  } catch {
    return { defaultMode: "assistant", density: "comfortable" };
  }
}

export function saveBrowserWorkspaceState(value: BrowserWorkspaceState, storage: Storage = window.localStorage): void {
  storage.setItem(browserKey, JSON.stringify(value));
}

export function loadProjectWorkspaceState(projectId: string, storage: Storage = window.localStorage): Partial<WorkspaceProjectState> {
  const raw = storage.getItem(`${projectPrefix}${projectId}`);
  if (!raw) return {};
  try {
    return JSON.parse(raw) as Partial<WorkspaceProjectState>;
  } catch {
    return {};
  }
}

export function saveProjectWorkspaceState(
  projectId: string,
  value: WorkspaceProjectState,
  storage: Storage = window.localStorage,
): void {
  storage.setItem(`${projectPrefix}${projectId}`, JSON.stringify(value));
}
```

- [x] **Step 2: Create workspace slice**

Create `frontend/src/features/workspace/workspaceSlice.ts`:

```ts
import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export type WorkspaceMode = "draft" | "assistant" | "review";
export type DrawerTab = "contents" | "world" | "notes" | "chat" | "tools" | "revisions" | "model";
export type MobileSheet = null | "contents" | "assistant" | "review" | "world-notes" | "model";

export type WorkspaceProjectState = {
  mode: WorkspaceMode;
  leftDrawerOpen: boolean;
  rightDrawerOpen: boolean;
  leftDrawerWidth: number;
  rightDrawerWidth: number;
  activeLeftTab: DrawerTab;
  activeRightTab: DrawerTab;
  lastContentKind: string;
  fallbackContentId: string | null;
};

export type WorkspaceState = WorkspaceProjectState & {
  mobileSheet: MobileSheet;
};

export const initialWorkspaceState: WorkspaceState = {
  mode: "assistant",
  leftDrawerOpen: true,
  rightDrawerOpen: true,
  leftDrawerWidth: 304,
  rightDrawerWidth: 380,
  activeLeftTab: "contents",
  activeRightTab: "chat",
  lastContentKind: "chapter",
  fallbackContentId: null,
  mobileSheet: null,
};

const modeDefaults: Record<WorkspaceMode, Pick<WorkspaceState, "leftDrawerOpen" | "rightDrawerOpen" | "activeRightTab">> = {
  draft: { leftDrawerOpen: true, rightDrawerOpen: false, activeRightTab: "chat" },
  assistant: { leftDrawerOpen: true, rightDrawerOpen: true, activeRightTab: "chat" },
  review: { leftDrawerOpen: true, rightDrawerOpen: true, activeRightTab: "tools" },
};

const workspaceSlice = createSlice({
  name: "workspace",
  initialState: initialWorkspaceState,
  reducers: {
    setMode(state, action: PayloadAction<WorkspaceMode>) {
      state.mode = action.payload;
      Object.assign(state, modeDefaults[action.payload]);
    },
    setLeftDrawerOpen(state, action: PayloadAction<boolean>) {
      state.leftDrawerOpen = action.payload;
    },
    setRightDrawerOpen(state, action: PayloadAction<boolean>) {
      state.rightDrawerOpen = action.payload;
    },
    setDrawerWidths(state, action: PayloadAction<{ left?: number; right?: number }>) {
      if (action.payload.left) state.leftDrawerWidth = Math.min(420, Math.max(240, action.payload.left));
      if (action.payload.right) state.rightDrawerWidth = Math.min(520, Math.max(320, action.payload.right));
    },
    setActiveLeftTab(state, action: PayloadAction<DrawerTab>) {
      state.activeLeftTab = action.payload;
    },
    setActiveRightTab(state, action: PayloadAction<DrawerTab>) {
      state.activeRightTab = action.payload;
      state.rightDrawerOpen = true;
    },
    setMobileSheet(state, action: PayloadAction<MobileSheet>) {
      state.mobileSheet = action.payload;
    },
    setLastContentKind(state, action: PayloadAction<string>) {
      state.lastContentKind = action.payload;
    },
    setFallbackContentId(state, action: PayloadAction<string | null>) {
      state.fallbackContentId = action.payload;
    },
    hydrateProjectState(state, action: PayloadAction<Partial<WorkspaceProjectState>>) {
      Object.assign(state, { ...action.payload, mobileSheet: state.mobileSheet });
    },
  },
});

export const workspaceActions = workspaceSlice.actions;
export const workspaceReducer = workspaceSlice.reducer;
```

- [x] **Step 3: Create editor slice**

Create `frontend/src/features/editor/editorSlice.ts`:

```ts
import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export type EditorActionKind = "rewrite" | "check" | "note";

export type EditorSelection = {
  startByte: number;
  endByte: number;
  previewText: string;
} | null;

export type EditorActionModal = {
  kind: EditorActionKind;
  instructions: string;
} | null;

export type EditorState = {
  cursorByte: number | null;
  selection: EditorSelection;
  actionModal: EditorActionModal;
  generateInstructions: string;
};

export const initialEditorState: EditorState = {
  cursorByte: null,
  selection: null,
  actionModal: null,
  generateInstructions: "",
};

const editorSlice = createSlice({
  name: "editor",
  initialState: initialEditorState,
  reducers: {
    setCursorByte(state, action: PayloadAction<number | null>) {
      state.cursorByte = action.payload;
    },
    setSelection(state, action: PayloadAction<EditorSelection>) {
      state.selection = action.payload;
    },
    openActionModal(state, action: PayloadAction<{ kind: EditorActionKind; instructions?: string }>) {
      state.actionModal = { kind: action.payload.kind, instructions: action.payload.instructions ?? "" };
    },
    setActionInstructions(state, action: PayloadAction<string>) {
      if (state.actionModal) state.actionModal.instructions = action.payload;
    },
    closeActionModal(state) {
      state.actionModal = null;
    },
    setGenerateInstructions(state, action: PayloadAction<string>) {
      state.generateInstructions = action.payload;
    },
  },
});

export const editorActions = editorSlice.actions;
export const editorReducer = editorSlice.reducer;
```

- [x] **Step 4: Add reducer tests**

Create `frontend/src/features/workspace/workspaceSlice.test.ts`:

```ts
import { describe, expect, it } from "vitest";
import { initialWorkspaceState, workspaceActions, workspaceReducer } from "./workspaceSlice";

describe("workspaceSlice", () => {
  it("applies preset drawer defaults", () => {
    const draft = workspaceReducer(initialWorkspaceState, workspaceActions.setMode("draft"));
    expect(draft.rightDrawerOpen).toBe(false);
    expect(draft.leftDrawerOpen).toBe(true);

    const review = workspaceReducer(draft, workspaceActions.setMode("review"));
    expect(review.rightDrawerOpen).toBe(true);
    expect(review.activeRightTab).toBe("tools");
  });

  it("clamps drawer widths", () => {
    const state = workspaceReducer(initialWorkspaceState, workspaceActions.setDrawerWidths({ left: 120, right: 900 }));
    expect(state.leftDrawerWidth).toBe(240);
    expect(state.rightDrawerWidth).toBe(520);
  });
});
```

Create `frontend/src/features/editor/editorSlice.test.ts`:

```ts
import { describe, expect, it } from "vitest";
import { editorActions, editorReducer, initialEditorState } from "./editorSlice";

describe("editorSlice", () => {
  it("stores selection and opens rewrite modal with draft instructions", () => {
    const selected = editorReducer(
      initialEditorState,
      editorActions.setSelection({ startByte: 4, endByte: 16, previewText: "selected prose" }),
    );
    const modal = editorReducer(selected, editorActions.openActionModal({ kind: "rewrite", instructions: "tighten" }));
    expect(modal.selection?.previewText).toBe("selected prose");
    expect(modal.actionModal).toEqual({ kind: "rewrite", instructions: "tighten" });
  });
});
```

- [x] **Step 5: Run tests**

Run:

```bash
cd frontend
bun run test
```

Expected: the two reducer test files pass.

---

## Task 4: Add shadcn/ui Foundation And Byte Offset Utilities

**Files:**
- Create: `frontend/components.json`
- Create: `frontend/src/shared/lib/utils.ts`
- Create: `frontend/src/shared/lib/byteOffsets.ts`
- Create: `frontend/src/shared/ui/button.tsx`
- Create: `frontend/src/shared/ui/input.tsx`
- Create: `frontend/src/shared/ui/textarea.tsx`
- Create: `frontend/src/shared/ui/dialog.tsx`
- Create: `frontend/src/shared/ui/sheet.tsx`
- Create: `frontend/src/shared/ui/tabs.tsx`
- Modify: `frontend/tsconfig.json`
- Modify: `frontend/vite.config.ts`
- Modify: `frontend/src/styles.css`

- [x] **Step 1: Initialize shadcn/ui for the existing Vite app**

Run from `frontend/`:

```bash
bunx --bun shadcn@latest init
```

Choose settings compatible with this project:

- framework: Vite / React
- Tailwind: v4
- global CSS: `src/styles.css`
- import alias for components: `@/shared/ui`
- import alias for utils: `@/shared/lib/utils`
- icon library: lucide

Expected: `components.json` is created and shadcn theme variables are added to `src/styles.css`.

- [x] **Step 2: Add required shadcn components**

Run from `frontend/`:

```bash
bunx --bun shadcn@latest add button input textarea dialog sheet tabs
```

Expected files are created under `frontend/src/shared/ui/`.

- [x] **Step 3: Verify shadcn project context**

Run from `frontend/`:

```bash
bunx --bun shadcn@latest info
```

Expected: command reports the configured aliases and installed components. If it reports no shadcn project, fix `components.json` before continuing.

- [x] **Step 4: Ensure `cn` utility exists**

If shadcn did not create `frontend/src/shared/lib/utils.ts`, create it:

```ts
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
```

- [x] **Step 5: Add UTF-8 byte offset helper**

Create `frontend/src/shared/lib/byteOffsets.ts`:

```ts
const encoder = new TextEncoder();

export function utf8ByteLength(value: string): number {
  return encoder.encode(value).length;
}

export function textareaSelectionToByteRange(
  value: string,
  selectionStart: number,
  selectionEnd: number,
): { startByte: number; endByte: number; previewText: string } | null {
  if (selectionStart === selectionEnd) return null;
  return {
    startByte: utf8ByteLength(value.slice(0, selectionStart)),
    endByte: utf8ByteLength(value.slice(0, selectionEnd)),
    previewText: value.slice(selectionStart, selectionEnd),
  };
}
```

- [x] **Step 6: Read added shadcn files**

Read the generated files under `frontend/src/shared/ui/` and verify:

- Components import `cn` from `@/shared/lib/utils` or the configured alias.
- Dialog and Sheet exports include Title components for accessibility.
- No generated imports point to `@/components/ui`.
- Styling uses semantic tokens such as `bg-background`, `text-foreground`, `border-input`, `text-muted-foreground`.

- [x] **Step 7: Run typecheck**

Run:

```bash
cd frontend
bun run typecheck
```

Expected: TypeScript passes.

---

## Task 5: Add Auth And Project Routes

**Files:**
- Create: `frontend/src/features/auth/AuthPage.tsx`
- Create: `frontend/src/features/projects/projectHooks.ts`
- Create: `frontend/src/features/projects/ProjectsPage.tsx`

- [x] **Step 1: Create login/register page**

Create `AuthPage` using existing `login`, `register`, `getToken` and `useNavigate`. On successful auth, navigate to `/projects`. If already authenticated, redirect to `/projects`.

Required visible controls:

- Email input.
- Password input.
- Login/Register submit button.
- Mode toggle button.
- Recoverable error region with `role="alert"`.

- [x] **Step 2: Create project hook**

Create `useProjects()` in `projectHooks.ts` using existing `listProjects()`, local loading/error state, and abort-safe effect cleanup.

- [x] **Step 3: Create projects page**

Create `ProjectsPage` that lists projects as buttons/links to `/projects/:projectId`. If there is exactly one project, keep the list visible rather than auto-navigating, because the user needs a stable project landing route.

- [x] **Step 4: Verify routes manually**

Run:

```bash
cd frontend
bun run typecheck
```

Expected: TypeScript passes for route/auth/project files.

---

## Task 6: Build Workspace Page And Shell

**Files:**
- Create: `frontend/src/features/workspace/WorkspacePage.tsx`
- Create: `frontend/src/features/workspace/WorkspaceShell.tsx`
- Create: `frontend/src/features/assistant/AssistantDrawer.tsx`
- Create: `frontend/src/features/review/ReviewDrawer.tsx`
- Create: `frontend/src/features/notes/ContextDrawer.tsx`
- Create: `frontend/src/features/model-settings/ModelStatus.tsx`
- Create: `frontend/src/features/skills/SkillChipsPanel.tsx`

- [x] **Step 1: Load route-owned project/content state**

`WorkspacePage` must use:

```ts
const { projectId, contentKind, contentId } = useParams();
```

Then load:

- `listProjects()` to find the active project title.
- `listContent(projectId, kind)` for the current content kind.
- If `contentId` exists, select that content from the loaded list.
- If `contentId` does not exist, use `workspace.fallbackContentId` only when it belongs to the loaded list.
- If neither exists, use the first item in the loaded list.

- [x] **Step 2: Keep deep links authoritative**

When a user clicks a content row, navigate to:

```ts
navigate(`/projects/${projectId}/content/${item.kind}/${item.id}`);
```

Do not store route-owned `projectId` or `contentId` in Redux.

- [x] **Step 3: Build desktop layout**

`WorkspaceShell` should render:

- Mode rail with Draft, Assistant, Review buttons.
- Left drawer for content/world/notes tabs.
- Center editor stage with max-width prose column of `min(100%, 760px)`.
- Right drawer for assistant or review surfaces.

Draft mode defaults:

```ts
rightDrawerOpen = false
activeRightTab = "chat"
```

Assistant mode defaults:

```ts
rightDrawerOpen = true
activeRightTab = "chat"
```

Review mode defaults:

```ts
rightDrawerOpen = true
activeRightTab = "tools"
```

- [x] **Step 4: Build mobile layout**

At max-width `760px`:

- Hide persistent left/right drawers.
- Keep editor as the primary screen.
- Show top toolbar buttons for Files, Assistant, Review, World/Notes, Model.
- Render `Sheet` based on `workspace.mobileSheet`.
- Keep Generate composer visible below editor.

- [x] **Step 5: Add structural drawer content**

This first slice should not reimplement every legacy form. Add stable placeholders wired to real context:

- `AssistantDrawer`: chat transcript placeholder, minimal selected model disclosure, quick action status area. It must not own provider settings, skill administration, or script runtime controls; those move to settings in Phase 3.5.
- `ReviewDrawer`: Read/Check reports area, revisions area, activity area.
- `ContextDrawer`: content list, story bible/world lookup placeholder, attached notes placeholder.
- `ModelStatus`: selected model unavailable state and settings link placeholder.
- `SkillChipsPanel`: superseded by Phase 3.5 settings IA; do not render it in the assistant right panel.

Text should be concise UI labels only, not tutorial copy.

---

## Task 7: Build Editor Frame And Local Actions

**Files:**
- Create: `frontend/src/features/editor/EditorFrame.tsx`
- Create: `frontend/src/features/editor/GenerateComposer.tsx`
- Create: `frontend/src/features/editor/SelectionActionDialog.tsx`
- Modify: `frontend/src/features/workspace/WorkspaceShell.tsx`

- [x] **Step 1: Create editor frame**

`EditorFrame` props:

```ts
type EditorFrameProps = {
  content: ContentItem | null;
  mode: WorkspaceMode;
};
```

Render:

- Title and revision.
- Read-only textarea for this slice.
- `onSelect` converts selection to UTF-8 byte offsets via `textareaSelectionToByteRange`.
- Desktop selection bubble with Rewrite, Check, Note when selection exists.
- Mobile sticky toolbar with Rewrite, Check, Note when selection exists.

- [x] **Step 2: Create Generate composer**

`GenerateComposer` reads/writes `editor.generateInstructions`. It has:

- Instruction textarea.
- Generate button with a Lucide icon.
- Disabled state when there is no selected content.

For this slice, `onGenerate` may write a route-level recoverable error/status saying backend wiring is next. Do not call the continuation API until action scope, cursor handling, and selected model are passed in safely.

- [x] **Step 3: Create Rewrite/Check dialog**

`SelectionActionDialog` must:

- Open for `rewrite` or `check`.
- Show selected text preview with CSS line clamp/ellipsis.
- Preserve instruction text while open.
- Use `editorActions.setActionInstructions`.
- Close on Cancel.
- Render a disabled Run/Preview button if there is no selection.

- [x] **Step 4: Add Note shell**

Clicking Note should open the same modal shell with `kind: "note"` and a note instruction field, but submit should stay disabled in this slice because attached-note backend wiring is out of scope for this foundation plan.

---

## Task 8: Replace Old Global CSS With Milestone 4 Layout CSS

**Files:**
- Modify: `frontend/src/styles.css`

- [x] **Step 1: Add theme tokens**

Use CSS variables under `:root`:

```css
:root {
  color-scheme: light;
  --background: #f7f7f5;
  --surface: #ffffff;
  --surface-muted: #f1f3f5;
  --border: #d8dee4;
  --text: #1f2933;
  --muted: #667085;
  --accent: #2f6f73;
  --accent-strong: #184e52;
}
```

- [x] **Step 2: Add layout classes**

Required classes:

- `.workspace-page`
- `.workspace-shell-v2`
- `.workspace-rail`
- `.workspace-context-drawer`
- `.workspace-editor-stage`
- `.workspace-prose-column`
- `.workspace-side-drawer`
- `.mobile-workspace-toolbar`
- `.mobile-selection-toolbar`
- `.selection-bubble`
- `.generate-composer`
- `.editor-action-preview`

- [x] **Step 3: Enforce editor max width**

The prose column rule must include:

```css
.workspace-prose-column {
  width: min(100%, 760px);
  margin-inline: auto;
}
```

- [x] **Step 4: Add mobile sheet behavior**

At `@media (max-width: 760px)`, hide desktop drawers and ensure:

```css
.workspace-editor-stage {
  min-width: 0;
}

.workspace-prose-column {
  width: 100%;
}
```

---

## Task 9: Verification And Commit

**Files:**
- All files from Tasks 1-8

- [x] **Step 1: Run frontend tests**

Run:

```bash
cd frontend
bun run test
```

Expected: reducer tests pass.

- [x] **Step 2: Run frontend build**

Run:

```bash
cd frontend
bun run build
```

Expected: TypeScript and Vite build pass.

- [x] **Step 3: Run backend tests**

Run:

```bash
go test -tags sqlite_fts5 ./...
```

Expected: existing Go test suite still passes.

- [x] **Step 4: Browser smoke**

Run frontend dev server:

```bash
cd frontend
bun run dev -- --host 127.0.0.1
```

Open:

```txt
http://127.0.0.1:5173/login
http://127.0.0.1:5173/projects
http://127.0.0.1:5173/projects/<projectId>
http://127.0.0.1:5173/projects/<projectId>/content/<contentKind>/<contentId>
```

Smoke checks:

- `/login` renders auth form.
- `/projects` redirects to `/login` without a token.
- With a token, `/projects` lists projects.
- Workspace mode buttons change drawer state.
- Editor column stays visually constrained around 700-800px on desktop.
- Mobile viewport hides drawers and exposes sheet buttons.
- Selecting text shows desktop bubble or mobile sticky toolbar.
- Rewrite/Check opens a modal/sheet shell with selected text preview and instructions.

- [x] **Step 5: Commit**

Run:

```bash
git status --short
git add frontend docs/superpowers/plans/2026-06-14-milestone-4-workspace-foundation.md
git commit -m "feat: add milestone 4 workspace foundation"
```

Expected: commit contains the plan and the frontend foundation.

---

## Self-Review

Spec coverage:

- Routing: covered by Tasks 2, 5, and 6.
- Redux workspace/editor state: covered by Task 3.
- Hybrid local persistence: covered by Task 3.
- Domain-oriented vertical slices: covered by File Structure and Tasks 2-7.
- Draft/Assistant/Review presets: covered by Tasks 3 and 6.
- Desktop dual drawers and editor max-width: covered by Tasks 6 and 8.
- Mobile editor-first sheets: covered by Tasks 6 and 8.
- Generate composer under editor: covered by Task 7.
- Desktop selection bubble and mobile sticky toolbar: covered by Task 7.
- Rewrite/Check modal with selected preview and instructions: covered by Task 7.
- shadcn/Tailwind foundation: covered by Tasks 1, 4, and 8.

Intentional gaps for later plans:

- Full migration of existing Agent Settings and Skill Browser forms into vertical slices.
- Real Generate/Rewrite/Check execution from the new composer/dialog shell.
- Galley Editor adapter.
- Revision restore UI and attached-note backend workflows.
- Browser screenshot verification through Playwright/DevTools after the dev server exists.
