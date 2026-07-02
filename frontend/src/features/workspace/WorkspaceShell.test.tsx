import { configureStore } from "@reduxjs/toolkit";
import { renderToStaticMarkup } from "react-dom/server";
import { Provider } from "react-redux";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { assistantReducer, initialAssistantState } from "../assistant/assistantSlice";
import {
  assistantActionsReducer,
  initialAssistantActionsState,
} from "../assistant-actions/assistantActionsSlice";
import { editorReducer, initialEditorState } from "../editor/editorSlice";
import { initialModelSettingsState, modelSettingsReducer } from "../model-settings/modelSettingsSlice";
import { initialReviewState, reviewReducer } from "../review/reviewSlice";
import { initialScriptRuntimeState, scriptRuntimeReducer } from "../script-runtime/scriptRuntimeSlice";
import { initialSettingsState, settingsReducer } from "../settings/settingsSlice";
import { initialSkillsState, skillsReducer } from "../skills/skillsSlice";
import type { ContentItem } from "../../types";
import { initialWorkspaceState, workspaceReducer, type WorkspaceState } from "./workspaceSlice";
import { WorkspaceShell } from "./WorkspaceShell";

const content: ContentItem = {
  id: "content-1",
  projectId: "project-1",
  kind: "chapter",
  title: "Opening",
  slug: "opening",
  bodyMarkdown: "The lantern burned blue.",
  metadataJson: "{}",
  sortOrder: 1,
  currentRevision: 1,
};

function renderShell({
  workspace = initialWorkspaceState,
  selectedContent = content,
}: {
  workspace?: WorkspaceState;
  selectedContent?: ContentItem | null;
} = {}): string {
  const store = configureStore({
    reducer: {
      assistant: assistantReducer,
      assistantActions: assistantActionsReducer,
      editor: editorReducer,
      modelSettings: modelSettingsReducer,
      review: reviewReducer,
      scriptRuntime: scriptRuntimeReducer,
      settings: settingsReducer,
      skills: skillsReducer,
      workspace: workspaceReducer,
    },
    preloadedState: {
      assistant: initialAssistantState,
      assistantActions: initialAssistantActionsState,
      editor: initialEditorState,
      modelSettings: initialModelSettingsState,
      review: initialReviewState,
      scriptRuntime: initialScriptRuntimeState,
      settings: initialSettingsState,
      skills: initialSkillsState,
      workspace,
    },
  });

  return renderToStaticMarkup(
    <Provider store={store}>
      <MemoryRouter>
        <WorkspaceShell
          projectId="project-1"
          projectTitle="Alchemy Draft"
          contentItems={selectedContent ? [selectedContent] : []}
          contentLoading={false}
          contentError={null}
          contentCreateError={null}
          contentCreating={false}
          activeContentKind="chapter"
          selectedContent={selectedContent}
          onCreateContent={() => undefined}
          onSelectContent={() => undefined}
          onContentKindChange={() => undefined}
          onContentSaved={() => undefined}
        />
      </MemoryRouter>
    </Provider>,
  );
}

describe("WorkspaceShell", () => {
  it("renders desktop Assistant mode with assistant drawer and editor content", () => {
    const html = renderShell({
      workspace: {
        ...initialWorkspaceState,
        mode: "assistant",
        rightDrawerOpen: true,
        activeRightTab: "chat",
      },
    });

    expect(html).toContain("Alchemy Draft");
    expect(html).toContain("Assistant");
    expect(html).toContain("Select a model before starting assistant chat.");
    expect(html).toContain("Opening");
    expect(html).toContain("Saved version 1");
  });

  it("renders desktop Review mode with selected content review surfaces", () => {
    const html = renderShell({
      workspace: {
        ...initialWorkspaceState,
        mode: "review",
        rightDrawerOpen: true,
        activeRightTab: "tools",
      },
    });

    expect(html).toContain("Review");
    expect(html).toContain("Current checkpoint 1");
    expect(html).toContain("Opening");
  });

  it("renders mobile sheet controls for authoring panels", () => {
    const html = renderShell();

    expect(html).toContain("Files");
    expect(html).toContain("Assistant");
    expect(html).toContain("Review");
    expect(html).toContain("World/Notes");
  });

  it("renders editor and review empty states when no content is selected", () => {
    const html = renderShell({
      workspace: {
        ...initialWorkspaceState,
        mode: "review",
        rightDrawerOpen: true,
        activeRightTab: "tools",
      },
      selectedContent: null,
    });

    expect(html).toContain("Select content to start drafting.");
    expect(html).toContain("Select content to review checkpoints.");
  });
});
