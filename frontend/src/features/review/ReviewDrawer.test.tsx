import { configureStore } from "@reduxjs/toolkit";
import { renderToStaticMarkup } from "react-dom/server";
import { Provider } from "react-redux";
import { describe, expect, it } from "vitest";

import type { PromptRecord, ReadAndCheckResult } from "../../agentTypes";
import type { ContentItem, Revision } from "../../types";
import {
  assistantActionsReducer,
  initialAssistantActionsState,
  type AssistantActionState,
} from "../assistant-actions/assistantActionsSlice";
import { initialReviewState, reviewReducer, type ReviewState } from "./reviewSlice";
import { ReviewDrawer } from "./ReviewDrawer";

const content: ContentItem = {
  id: "content-1",
  projectId: "project-1",
  kind: "chapter",
  title: "Opening",
  slug: "opening",
  bodyMarkdown: "The lantern burned green.",
  metadataJson: "{}",
  sortOrder: 1,
  currentRevision: 2,
};

const revision: Revision = {
  id: "revision-1",
  contentItemId: "content-1",
  revisionNumber: 1,
  bodyMarkdown: "The lantern burned blue.",
  metadataJson: "{}",
  reason: "initial content",
  createdBy: "author",
  createdAt: "2026-07-01T00:00:00.000Z",
};

const promptRecord: PromptRecord = {
  id: "prompt-record-1",
  projectId: "project-1",
  sessionId: "session-1",
  providerName: "Provider",
  modelName: "model-a",
  actionKind: "read_check",
  requestJson: "{\"contentId\":\"content-1\"}",
  responseJson: "{\"status\":\"ok\"}",
  usage: {
    inputTokens: 10,
    outputTokens: 5,
    cacheReadTokens: 0,
    cacheWriteTokens: 0,
    totalTokens: 15,
    inputCost: 0.001,
    outputCost: 0.002,
    cacheReadCost: 0,
    cacheWriteCost: 0,
    totalCost: 0.003,
  },
  createdAt: "2026-07-01T00:01:00.000Z",
};

const checkResult: ReadAndCheckResult = {
  session: {
    id: "session-1",
    projectId: "project-1",
    title: "Check",
    actionKind: "read_check",
    modelVariantId: "model-1",
    applyMode: "preview",
    createdAt: "2026-07-01T00:00:00.000Z",
    updatedAt: "2026-07-01T00:00:00.000Z",
  },
  assistantMessage: {
    id: "message-1",
    sessionId: "session-1",
    role: "assistant",
    bodyMarkdown: "Continuity report",
    metadataJson: "{}",
    createdAt: "2026-07-01T00:00:00.000Z",
  },
  note: {
    id: "note-1",
    projectId: "project-1",
    contentItemId: "content-1",
    selectionStart: 4,
    selectionEnd: 11,
    title: "Check",
    bodyMarkdown: "Continuity report",
    source: "read_and_check",
    createdAt: "2026-07-01T00:00:00.000Z",
    updatedAt: "2026-07-01T00:00:00.000Z",
  },
};

function renderDrawer({
  review = initialReviewState,
  assistantActions = initialAssistantActionsState,
  selectedContent = content,
}: {
  review?: ReviewState;
  assistantActions?: AssistantActionState;
  selectedContent?: ContentItem | null;
}): string {
  const store = configureStore({
    reducer: {
      review: reviewReducer,
      assistantActions: assistantActionsReducer,
    },
    preloadedState: {
      review,
      assistantActions,
    },
  });

  return renderToStaticMarkup(
    <Provider store={store}>
      <ReviewDrawer projectId="project-1" content={selectedContent} onContentSaved={() => undefined} />
    </Provider>,
  );
}

describe("ReviewDrawer", () => {
  it("renders a content-scoped empty state when no content is selected", () => {
    const html = renderDrawer({ selectedContent: null });

    expect(html).toContain("Select content to review checkpoints.");
  });

  it("renders checkpoint items and a compact diff for the selected content", () => {
    const html = renderDrawer({
      review: {
        ...initialReviewState,
        projectId: "project-1",
        contentId: "content-1",
        revisions: [revision],
        revisionsStatus: "succeeded",
        selectedRevisionNumber: 1,
      },
    });

    expect(html).toContain("Current checkpoint 2");
    expect(html).toContain("Checkpoint 1");
    expect(html).toContain("Checkpoint 1 diff");
    expect(html).toContain("The lantern burned green.");
    expect(html).toContain("The lantern burned blue.");
    expect(html).toContain("Restore");
  });

  it("renders selected prompt record details", () => {
    const html = renderDrawer({
      review: {
        ...initialReviewState,
        projectId: "project-1",
        promptRecords: [promptRecord],
        promptRecordsStatus: "succeeded",
        selectedPromptRecordId: promptRecord.id,
      },
    });

    expect(html).toContain("Prompt record details");
    expect(html).toContain("&quot;contentId&quot;: &quot;content-1&quot;");
    expect(html).toContain("&quot;status&quot;: &quot;ok&quot;");
  });

  it("renders the current check report for selected content", () => {
    const html = renderDrawer({
      assistantActions: {
        ...initialAssistantActionsState,
        actionKind: "check",
        status: "succeeded",
        checkResult,
      },
    });

    expect(html).toContain("Continuity report");
    expect(html).toContain("Attached to bytes 4-11");
  });

  it("renders restore conflict guidance", () => {
    const html = renderDrawer({
      review: {
        ...initialReviewState,
        projectId: "project-1",
        contentId: "content-1",
        revisions: [revision],
        revisionsStatus: "succeeded",
        selectedRevisionNumber: 1,
        restoreStatus: "failed",
        restoreError: "restore revision failed: 409",
      },
    });

    expect(html).toContain("Content changed before this checkpoint was restored.");
  });
});
