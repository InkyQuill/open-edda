# Milestone 4 Behavior Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore the old monolithic frontend's working assistant, model settings, skills, activity, prompt-record, and script-runtime visibility inside the new routed Milestone 4 workspace architecture.

**Architecture:** Keep the current domain-oriented vertical-slice structure. Route params continue to own `projectId`, `contentKind`, and `contentId`; Redux owns workspace/editor/assistant/model/skills UI state; existing API modules remain the transport layer until each feature has enough code to justify moving API adapters into that slice.

**Tech Stack:** React + Vite, React Router, Redux Toolkit, React Redux, Tailwind v4, shadcn/ui primitives, TypeScript, Vitest.

---

## Phase Position

This is Milestone 4 Phase 2.

Phase 1 is complete in `docs/superpowers/plans/2026-06-14-milestone-4-workspace-foundation.md`.

This phase does not redesign the UI. It restores product behavior and data ownership after the routed shell split. Visual polish, Galley Editor integration, and full action execution belong to later phases.

## File Structure

Create or modify these files:

```txt
frontend/src/app/store/store.ts
frontend/src/features/assistant/AssistantDrawer.tsx
frontend/src/features/assistant/assistantSlice.ts
frontend/src/features/assistant/assistantSlice.test.ts
frontend/src/features/assistant/assistantThunks.ts
frontend/src/features/model-settings/ModelStatus.tsx
frontend/src/features/model-settings/ModelSettingsPanel.tsx
frontend/src/features/model-settings/modelSettingsSlice.ts
frontend/src/features/model-settings/modelSettingsSlice.test.ts
frontend/src/features/model-settings/modelSettingsThunks.ts
frontend/src/features/review/ReviewDrawer.tsx
frontend/src/features/review/reviewSlice.ts
frontend/src/features/review/reviewSlice.test.ts
frontend/src/features/review/reviewThunks.ts
frontend/src/features/script-runtime/ScriptRuntimePanel.tsx
frontend/src/features/script-runtime/scriptRuntimeSlice.ts
frontend/src/features/script-runtime/scriptRuntimeSlice.test.ts
frontend/src/features/script-runtime/scriptRuntimeThunks.ts
frontend/src/features/skills/SkillChipsPanel.tsx
frontend/src/features/skills/SkillsPanel.tsx
frontend/src/features/skills/skillsSlice.ts
frontend/src/features/skills/skillsSlice.test.ts
frontend/src/features/skills/skillsThunks.ts
frontend/src/features/workspace/WorkspacePage.tsx
frontend/src/features/workspace/WorkspaceShell.tsx
```

Keep existing transport modules during this phase:

```txt
frontend/src/agentApi.ts
frontend/src/agentTypes.ts
frontend/src/scriptRuntimeApi.ts
frontend/src/scriptRuntimeTypes.ts
frontend/src/skillApi.ts
frontend/src/skillTypes.ts
```

## Task 1: Register Feature Slices In The App Store

**Files:**
- Modify: `frontend/src/app/store/store.ts`
- Create: `frontend/src/features/assistant/assistantSlice.ts`
- Create: `frontend/src/features/model-settings/modelSettingsSlice.ts`
- Create: `frontend/src/features/review/reviewSlice.ts`
- Create: `frontend/src/features/script-runtime/scriptRuntimeSlice.ts`
- Create: `frontend/src/features/skills/skillsSlice.ts`
- Test: `frontend/src/features/assistant/assistantSlice.test.ts`
- Test: `frontend/src/features/model-settings/modelSettingsSlice.test.ts`
- Test: `frontend/src/features/review/reviewSlice.test.ts`
- Test: `frontend/src/features/script-runtime/scriptRuntimeSlice.test.ts`
- Test: `frontend/src/features/skills/skillsSlice.test.ts`

- [ ] **Step 1: Add slice tests for reset-on-project-change behavior**

Each feature slice must expose a `resetForProject` reducer. The assistant test should use this shape:

```ts
import { describe, expect, it } from "vitest";
import { assistantActions, assistantReducer, initialAssistantState } from "./assistantSlice";

describe("assistantSlice", () => {
  it("resets project-scoped assistant state", () => {
    const loaded = assistantReducer(
      {
        ...initialAssistantState,
        activeSessionId: "session-1",
        draftMessage: "hello",
        sessionsStatus: "succeeded",
      },
      assistantActions.resetForProject(),
    );

    expect(loaded).toEqual(initialAssistantState);
  });
});
```

Repeat the same pattern for model settings, review, script runtime, and skills using state fields owned by that slice.

- [ ] **Step 2: Run the new tests and verify they fail**

Run:

```bash
cd frontend
bun run test -- assistantSlice modelSettingsSlice reviewSlice scriptRuntimeSlice skillsSlice
```

Expected: tests fail because the slice files do not exist yet.

- [ ] **Step 3: Create minimal slices**

Use explicit async status unions:

```ts
export type AsyncStatus = "idle" | "pending" | "succeeded" | "failed";
```

Assistant state:

```ts
import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export type AssistantState = {
  activeSessionId: string | null;
  draftMessage: string;
  sessionsStatus: AsyncStatus;
  messagesStatus: AsyncStatus;
  error: string | null;
};

export const initialAssistantState: AssistantState = {
  activeSessionId: null,
  draftMessage: "",
  sessionsStatus: "idle",
  messagesStatus: "idle",
  error: null,
};
```

Model settings state:

```ts
export type ModelSettingsState = {
  selectedProviderId: string | null;
  activeModelVariantId: string | null;
  providersStatus: AsyncStatus;
  modelsStatus: AsyncStatus;
  error: string | null;
};
```

Review state:

```ts
export type ReviewState = {
  activityStatus: AsyncStatus;
  promptRecordsStatus: AsyncStatus;
  selectedPromptRecordId: string | null;
  error: string | null;
};
```

Script runtime state:

```ts
export type ScriptRuntimeState = {
  auditsStatus: AsyncStatus;
  runsStatus: AsyncStatus;
  selectedAuditId: string | null;
  error: string | null;
};
```

Skills state:

```ts
export type SkillsState = {
  skillsStatus: AsyncStatus;
  selectedSkillIds: string[];
  importStatus: AsyncStatus;
  error: string | null;
};
```

Each slice must include:

```ts
resetForProject() {
  return initialState;
}
```

- [ ] **Step 4: Register reducers in `store.ts`**

Update the reducer map:

```ts
reducer: {
  workspace: workspaceReducer,
  editor: editorReducer,
  assistant: assistantReducer,
  modelSettings: modelSettingsReducer,
  review: reviewReducer,
  scriptRuntime: scriptRuntimeReducer,
  skills: skillsReducer,
}
```

- [ ] **Step 5: Run tests**

Run:

```bash
cd frontend
bun run test -- assistantSlice modelSettingsSlice reviewSlice scriptRuntimeSlice skillsSlice
```

Expected: PASS.

## Task 2: Load Assistant Sessions And Chat In The Assistant Drawer

**Files:**
- Create: `frontend/src/features/assistant/assistantThunks.ts`
- Modify: `frontend/src/features/assistant/assistantSlice.ts`
- Modify: `frontend/src/features/assistant/AssistantDrawer.tsx`
- Modify: `frontend/src/features/workspace/WorkspaceShell.tsx`
- Test: `frontend/src/features/assistant/assistantSlice.test.ts`

- [ ] **Step 1: Add reducer tests for sessions and draft message**

Add tests that verify:

```ts
assistantActions.setDraftMessage("Can you continue this scene?")
assistantActions.setActiveSessionId("session-1")
assistantActions.resetForProject()
```

Expected state:

```ts
draftMessage === "Can you continue this scene?"
activeSessionId === "session-1"
resetForProject() returns initialAssistantState
```

- [ ] **Step 2: Create async thunks using existing API functions**

`assistantThunks.ts` should import:

```ts
import { createAsyncThunk } from "@reduxjs/toolkit";
import { createSession, createSessionMessage, listSessions } from "../../agentApi";
```

Create these thunks:

```ts
export const loadAssistantSessions = createAsyncThunk(
  "assistant/loadSessions",
  async ({ projectId }: { projectId: string }) => listSessions(projectId, 20),
);

export const startAssistantSession = createAsyncThunk(
  "assistant/startSession",
  async ({ projectId, contentId }: { projectId: string; contentId: string | null }) =>
    createSession(projectId, {
      contentId,
      title: "Workspace chat",
      actionKind: "chat",
    }),
);

export const sendAssistantMessage = createAsyncThunk(
  "assistant/sendMessage",
  async ({ projectId, sessionId, bodyMarkdown }: { projectId: string; sessionId: string; bodyMarkdown: string }) =>
    createSessionMessage(projectId, sessionId, bodyMarkdown),
);
```

- [ ] **Step 3: Wire extraReducers**

The slice must:

- set `sessionsStatus` to `pending` while loading,
- store loaded sessions,
- set `activeSessionId` to the newest loaded session if the current ID is absent,
- clear `draftMessage` after a successful send,
- store readable errors from rejected thunks.

- [ ] **Step 4: Render real assistant data**

Update `AssistantDrawer` props:

```ts
type AssistantDrawerProps = {
  projectId: string;
  contentId: string | null;
};
```

The drawer should:

- load sessions on mount/project change,
- show current session title,
- show an empty chat state when no session exists,
- provide a textarea for `draftMessage`,
- provide `New chat` and `Send` buttons,
- disable `Send` when there is no active session or the draft is blank.

- [ ] **Step 5: Pass route context from `WorkspaceShell`**

`WorkspaceShell` should accept:

```ts
projectId: string;
selectedContent: ContentItem | null;
```

Pass:

```tsx
<AssistantDrawer projectId={projectId} contentId={selectedContent?.id ?? null} />
```

- [ ] **Step 6: Run focused tests and build**

Run:

```bash
cd frontend
bun run test -- assistantSlice
bun run build
```

Expected: both commands exit 0.

## Task 3: Restore Provider And Model Settings Visibility

**Files:**
- Create: `frontend/src/features/model-settings/modelSettingsThunks.ts`
- Modify: `frontend/src/features/model-settings/modelSettingsSlice.ts`
- Create: `frontend/src/features/model-settings/ModelSettingsPanel.tsx`
- Modify: `frontend/src/features/model-settings/ModelStatus.tsx`
- Modify: `frontend/src/features/workspace/WorkspaceShell.tsx`
- Test: `frontend/src/features/model-settings/modelSettingsSlice.test.ts`

- [ ] **Step 1: Add slice tests**

Cover:

- selecting a provider,
- selecting a model variant,
- clearing selected model when provider changes,
- resetting state for a new project.

- [ ] **Step 2: Create thunks for providers and models**

Use existing API functions:

```ts
import { createAsyncThunk } from "@reduxjs/toolkit";
import { listModelVariants, listProviderConfigs } from "../../agentApi";

export const loadProviderConfigs = createAsyncThunk(
  "modelSettings/loadProviderConfigs",
  async () => listProviderConfigs(),
);

export const loadModelVariants = createAsyncThunk(
  "modelSettings/loadModelVariants",
  async ({ providerId }: { providerId: string }) => listModelVariants(providerId),
);
```

- [ ] **Step 3: Render model disclosure**

`ModelStatus` should load providers on mount and show:

- selected provider name or `No provider selected`,
- selected model name/model id or `No model selected`,
- a disabled AI action warning when no model is selected.

- [ ] **Step 4: Add `ModelSettingsPanel`**

Use shadcn `Tabs`, `Button`, and `Input` if editing fields are needed. This phase may expose read/select behavior without full create/update forms. The panel must provide:

- provider list,
- model list for selected provider,
- selected model disclosure,
- load/error states.

- [ ] **Step 5: Run focused tests and build**

Run:

```bash
cd frontend
bun run test -- modelSettingsSlice
bun run build
```

Expected: both commands exit 0.

## Task 4: Restore Skills Visibility And Session Skill Selection

**Files:**
- Create: `frontend/src/features/skills/skillsThunks.ts`
- Modify: `frontend/src/features/skills/skillsSlice.ts`
- Modify: `frontend/src/features/skills/SkillChipsPanel.tsx`
- Create: `frontend/src/features/skills/SkillsPanel.tsx`
- Modify: `frontend/src/features/assistant/AssistantDrawer.tsx`
- Test: `frontend/src/features/skills/skillsSlice.test.ts`

- [ ] **Step 1: Add slice tests**

Cover:

- loading skills,
- toggling a selected skill id,
- replacing selected session skills from server data,
- clearing project-scoped data.

- [ ] **Step 2: Create thunks**

Use existing API functions:

```ts
import { createAsyncThunk } from "@reduxjs/toolkit";
import { listSessionSkills, listSkills, selectSessionSkills } from "../../skillApi";

export const loadProjectSkills = createAsyncThunk(
  "skills/loadProjectSkills",
  async ({ projectId }: { projectId: string }) => listSkills(projectId),
);

export const loadSessionSkills = createAsyncThunk(
  "skills/loadSessionSkills",
  async ({ projectId, sessionId }: { projectId: string; sessionId: string }) =>
    listSessionSkills(projectId, sessionId),
);

export const saveSessionSkills = createAsyncThunk(
  "skills/saveSessionSkills",
  async ({ projectId, sessionId, skillIds }: { projectId: string; sessionId: string; skillIds: string[] }) =>
    selectSessionSkills(projectId, sessionId, skillIds),
);
```

- [ ] **Step 3: Replace placeholder skill chips**

`SkillChipsPanel` should render real skill display names from Redux state. It should show:

- loading state,
- empty state when no skills are installed,
- selected skill chips,
- script-disabled disclosure when `WriterSkill` indicates scripts are disabled.

- [ ] **Step 4: Add `SkillsPanel`**

This panel should list installed skills and let the author select skills for the active assistant session. Use checkboxes or shadcn buttons with `aria-pressed`.

- [ ] **Step 5: Run focused tests and build**

Run:

```bash
cd frontend
bun run test -- skillsSlice
bun run build
```

Expected: both commands exit 0.

## Task 5: Restore Activity And Prompt Record Visibility In Review Mode

**Files:**
- Create: `frontend/src/features/review/reviewThunks.ts`
- Modify: `frontend/src/features/review/reviewSlice.ts`
- Modify: `frontend/src/features/review/ReviewDrawer.tsx`
- Test: `frontend/src/features/review/reviewSlice.test.ts`

- [ ] **Step 1: Add slice tests**

Cover:

- storing activity events,
- storing prompt records,
- selecting a prompt record,
- resetting on project change.

- [ ] **Step 2: Create thunks**

Use existing API functions:

```ts
import { createAsyncThunk } from "@reduxjs/toolkit";
import { listActivity, listPromptRecords } from "../../agentApi";

export const loadReviewActivity = createAsyncThunk(
  "review/loadActivity",
  async ({ projectId }: { projectId: string }) => listActivity(projectId, 50),
);

export const loadPromptRecords = createAsyncThunk(
  "review/loadPromptRecords",
  async ({ projectId }: { projectId: string }) => listPromptRecords(projectId, 50),
);
```

- [ ] **Step 3: Render review tabs from real data**

`ReviewDrawer` should keep tabs for `Reports`, `Revisions`, and `Activity`, but `Activity` must show loaded activity events and prompt records instead of static placeholder text.

- [ ] **Step 4: Run focused tests and build**

Run:

```bash
cd frontend
bun run test -- reviewSlice
bun run build
```

Expected: both commands exit 0.

## Task 6: Restore Script Runtime Visibility

**Files:**
- Create: `frontend/src/features/script-runtime/scriptRuntimeThunks.ts`
- Modify: `frontend/src/features/script-runtime/scriptRuntimeSlice.ts`
- Create: `frontend/src/features/script-runtime/ScriptRuntimePanel.tsx`
- Modify: `frontend/src/features/workspace/WorkspaceShell.tsx`
- Test: `frontend/src/features/script-runtime/scriptRuntimeSlice.test.ts`

- [ ] **Step 1: Add slice tests**

Cover:

- storing script audits,
- storing project script runs,
- selecting an audit,
- updating approval status,
- resetting project-scoped script data.

- [ ] **Step 2: Create thunks**

Use existing API functions:

```ts
import { createAsyncThunk } from "@reduxjs/toolkit";
import { listProjectScriptRuns, listScriptAudits, updateScriptApproval } from "../../scriptRuntimeApi";

export const loadScriptAudits = createAsyncThunk(
  "scriptRuntime/loadAudits",
  async ({ projectId }: { projectId: string }) => listScriptAudits(projectId),
);

export const loadScriptRuns = createAsyncThunk(
  "scriptRuntime/loadRuns",
  async ({ projectId }: { projectId: string }) => listProjectScriptRuns(projectId, 50),
);

export const setScriptApproval = createAsyncThunk(
  "scriptRuntime/setApproval",
  async ({ projectId, auditId, approved }: { projectId: string; auditId: string; approved: boolean }) =>
    updateScriptApproval(projectId, auditId, approved),
);
```

- [ ] **Step 3: Add `ScriptRuntimePanel` to the Model mobile sheet or right drawer tools area**

The panel should show:

- script audit status,
- run history,
- approval controls,
- clear disabled/deferred status for scripts that cannot run.

- [ ] **Step 4: Run focused tests and build**

Run:

```bash
cd frontend
bun run test -- scriptRuntimeSlice
bun run build
```

Expected: both commands exit 0.

## Task 7: Reset Project-Scoped Feature State From Workspace Route Changes

**Files:**
- Modify: `frontend/src/features/workspace/WorkspacePage.tsx`
- Test: existing feature slice tests

- [ ] **Step 1: Dispatch feature resets when `projectId` changes**

In the same effect that hydrates workspace state for a project, dispatch:

```ts
assistantActions.resetForProject();
modelSettingsActions.resetForProject();
reviewActions.resetForProject();
scriptRuntimeActions.resetForProject();
skillsActions.resetForProject();
```

Then hydrate workspace state:

```ts
dispatch(workspaceActions.hydrateProjectState(loadProjectWorkspaceState(projectId)));
```

- [ ] **Step 2: Run all frontend tests**

Run:

```bash
cd frontend
bun run test
```

Expected: all tests pass.

## Task 8: Final Verification And Commit

**Files:**
- All files changed in this plan.

- [ ] **Step 1: Run final frontend verification**

Run:

```bash
cd frontend
bun run test
bun run build
```

Expected: tests pass and production build completes.

- [ ] **Step 2: Run backend regression tests**

Run:

```bash
go test -tags sqlite_fts5 ./...
```

Expected: all Go packages pass.

- [ ] **Step 3: Run diff hygiene check**

Run:

```bash
git diff --check
```

Expected: no output.

- [ ] **Step 4: Request code review**

Use CodeRabbit or the available review workflow on the final diff. Address all correctness findings before committing.

- [ ] **Step 5: Commit**

Run:

```bash
git add frontend docs/superpowers/plans/2026-06-14-milestone-4-behavior-parity.md docs/roadmap.md docs/superpowers/specs/2026-06-14-milestone-4-daily-writing-polish-design.md
git commit -m "feat: restore milestone 4 behavior parity"
```

Expected: one commit containing Phase 2 behavior parity.

## Self-Review Notes

- Spec coverage: this plan covers Milestone 4 Phase 2 only. Galley Editor integration, action execution, diff/restore, attached note editing, and browser smoke tests are explicitly assigned to later Milestone 4 phases in `docs/roadmap.md`.
- Placeholder scan: this plan avoids unresolved placeholder markers and names exact files, commands, and expected outcomes.
- Type consistency: route-owned identity remains `projectId`, `contentKind`, and `contentId`; feature UI state stays in Redux slices.
