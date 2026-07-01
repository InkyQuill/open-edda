import { configureStore } from "@reduxjs/toolkit";
import { renderToStaticMarkup } from "react-dom/server";
import { Provider } from "react-redux";
import { describe, expect, it } from "vitest";

import type { GenerationCandidate, ReadAndCheckResult } from "../../agentTypes";
import {
  assistantActionsReducer,
  initialAssistantActionsState,
  type AssistantActionState,
} from "../assistant-actions/assistantActionsSlice";
import { AssistantActionPreview } from "./AssistantActionPreview";

function renderPreview(state: AssistantActionState): string {
  const store = configureStore({
    reducer: {
      assistantActions: assistantActionsReducer,
    },
    preloadedState: {
      assistantActions: state,
    },
  });

  return renderToStaticMarkup(
    <Provider store={store}>
      <AssistantActionPreview
        onAccept={() => undefined}
        onReject={() => undefined}
        onDismiss={() => undefined}
      />
    </Provider>,
  );
}

const candidate: GenerationCandidate = {
  id: "candidate-1",
  projectId: "project-1",
  sessionId: "session-1",
  contentItemId: "content-1",
  actionKind: "rewrite",
  operationKind: "replace",
  expectedRevision: 7,
  generatedMarkdown: "A sharper passage.",
  status: "pending",
  createdAt: "2026-07-01T00:00:00Z",
  updatedAt: "2026-07-01T00:00:00Z",
};

const checkResult: ReadAndCheckResult = {
  session: {
    id: "session-1",
    projectId: "project-1",
    title: "Check",
    actionKind: "read_check",
    modelVariantId: "model-1",
    applyMode: "preview",
    createdAt: "2026-07-01T00:00:00Z",
    updatedAt: "2026-07-01T00:00:00Z",
  },
  assistantMessage: {
    id: "message-1",
    sessionId: "session-1",
    role: "assistant",
    bodyMarkdown: "Continuity report",
    metadataJson: "{}",
    createdAt: "2026-07-01T00:00:00Z",
  },
  note: {
    id: "note-1",
    projectId: "project-1",
    contentItemId: "content-1",
    selectionStart: 0,
    selectionEnd: 10,
    title: "Check",
    bodyMarkdown: "Continuity report",
    source: "agent",
    createdAt: "2026-07-01T00:00:00Z",
    updatedAt: "2026-07-01T00:00:00Z",
  },
};

describe("AssistantActionPreview", () => {
  it("renders candidate text with accept and reject controls", () => {
    const html = renderPreview({
      ...initialAssistantActionsState,
      actionKind: "rewrite",
      status: "succeeded",
      candidate,
    });

    expect(html).toContain("Rewrite preview");
    expect(html).toContain("A sharper passage.");
    expect(html).toContain("Accept preview");
    expect(html).toContain("Reject preview");
  });

  it("renders check reports with dismiss control and no accept controls", () => {
    const html = renderPreview({
      ...initialAssistantActionsState,
      actionKind: "check",
      status: "succeeded",
      checkResult,
    });

    expect(html).toContain("Check result");
    expect(html).toContain("Continuity report");
    expect(html).toContain("Dismiss check result");
    expect(html).not.toContain("Accept preview");
  });

  it("renders running state as busy", () => {
    const html = renderPreview({
      ...initialAssistantActionsState,
      actionKind: "generate",
      status: "running",
    });

    expect(html).toContain("aria-busy=\"true\"");
    expect(html).toContain("Running generate");
  });
});
