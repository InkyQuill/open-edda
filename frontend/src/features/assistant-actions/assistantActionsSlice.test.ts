import { configureStore } from "@reduxjs/toolkit";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type {
  ContinuationRequest,
  ContinuationResult,
  GenerationCandidate,
  ReadAndCheckResult,
  RewriteRequest,
  RewriteResult,
} from "../../agentTypes";
import {
  runContinuation,
  runReadAndCheck,
  runRewrite,
} from "../../agentApi";
import {
  assistantActionsReducer,
  initialAssistantActionsState,
  runCheck,
  runGenerate,
  runRewriteAction,
} from "./assistantActionsSlice";

vi.mock("../../agentApi", () => ({
  runContinuation: vi.fn(),
  runRewrite: vi.fn(),
  runReadAndCheck: vi.fn(),
}));

const runContinuationMock = vi.mocked(runContinuation);
const runRewriteMock = vi.mocked(runRewrite);
const runReadAndCheckMock = vi.mocked(runReadAndCheck);

function makeStore() {
  return configureStore({
    reducer: {
      assistantActions: assistantActionsReducer,
    },
  });
}

const candidate: GenerationCandidate = {
  id: "candidate-1",
  projectId: "project-1",
  sessionId: "session-1",
  contentItemId: "content-1",
  actionKind: "continuation",
  operationKind: "insert",
  expectedRevision: 7,
  insertPosition: 12,
  generatedMarkdown: "Generated text",
  status: "pending",
  createdAt: "2026-07-01T00:00:00Z",
  updatedAt: "2026-07-01T00:00:00Z",
};

const session = {
  id: "session-1",
  projectId: "project-1",
  title: "Action",
  actionKind: "continuation" as const,
  modelVariantId: "model-1",
  applyMode: "preview" as const,
  createdAt: "2026-07-01T00:00:00Z",
  updatedAt: "2026-07-01T00:00:00Z",
};

describe("assistantActionsSlice", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("runs generate in preview mode with paragraph continuation defaults", async () => {
    const store = makeStore();
    const result: ContinuationResult = { session, candidate };
    runContinuationMock.mockResolvedValue(result);

    await store.dispatch(
      runGenerate({
        projectId: "project-1",
        contentId: "content-1",
        modelVariantId: "model-1",
        expectedRevision: 7,
        cursorByte: 12,
        instructions: "continue here",
        skillIds: ["skill-1"],
        requestKey: "route:content-1:7",
        requestToken: "token-1",
      }),
    );

    expect(runContinuationMock).toHaveBeenCalledWith("project-1", {
      contentId: "content-1",
      modelVariantId: "model-1",
      expectedRevision: 7,
      insertPosition: 12,
      guidance: "continue here",
      skillIds: ["skill-1"],
      applyMode: "preview",
      insert: true,
      continuationUnits: "paragraph",
      continuationCount: 1,
    } satisfies ContinuationRequest);
    expect(store.getState().assistantActions.candidate).toEqual(candidate);
    expect(store.getState().assistantActions.status).toBe("succeeded");
  });

  it("runs rewrite with selection bytes and preview mode", async () => {
    const store = makeStore();
    const result: RewriteResult = { session: { ...session, actionKind: "rewrite" }, candidate };
    runRewriteMock.mockResolvedValue(result);

    await store.dispatch(
      runRewriteAction({
        projectId: "project-1",
        contentId: "content-1",
        modelVariantId: "model-1",
        expectedRevision: 7,
        selectionStartByte: 3,
        selectionEndByte: 18,
        instructions: "tighten this",
        requestKey: "route:content-1:7",
        requestToken: "token-1",
      }),
    );

    expect(runRewriteMock).toHaveBeenCalledWith("project-1", {
      contentId: "content-1",
      modelVariantId: "model-1",
      expectedRevision: 7,
      selectionStart: 3,
      selectionEnd: 18,
      guidance: "tighten this",
      applyMode: "preview",
    } satisfies RewriteRequest);
    expect(store.getState().assistantActions.actionKind).toBe("rewrite");
    expect(store.getState().assistantActions.candidate).toEqual(candidate);
  });

  it("runs check with selection bytes and preview mode", async () => {
    const store = makeStore();
    const result: ReadAndCheckResult = {
      session: { ...session, actionKind: "read_check" },
      assistantMessage: {
        id: "message-1",
        sessionId: "session-1",
        role: "assistant",
        bodyMarkdown: "Report",
        metadataJson: "{}",
        createdAt: "2026-07-01T00:00:00Z",
      },
      note: {
        id: "note-1",
        projectId: "project-1",
        contentItemId: "content-1",
        selectionStart: 3,
        selectionEnd: 18,
        title: "Check",
        bodyMarkdown: "Report",
        source: "agent",
        createdAt: "2026-07-01T00:00:00Z",
        updatedAt: "2026-07-01T00:00:00Z",
      },
    };
    runReadAndCheckMock.mockResolvedValue(result);

    await store.dispatch(
      runCheck({
        projectId: "project-1",
        contentId: "content-1",
        modelVariantId: "model-1",
        expectedRevision: 7,
        selectionStartByte: 3,
        selectionEndByte: 18,
        instructions: "check tone",
        skillIds: ["skill-1"],
        requestKey: "route:content-1:7",
        requestToken: "token-1",
      }),
    );

    expect(runReadAndCheckMock).toHaveBeenCalledWith("project-1", {
      contentId: "content-1",
      modelVariantId: "model-1",
      expectedRevision: 7,
      selectionStart: 3,
      selectionEnd: 18,
      guidance: "check tone",
      skillIds: ["skill-1"],
      applyMode: "preview",
    } satisfies RewriteRequest);
    expect(store.getState().assistantActions.actionKind).toBe("check");
    expect(store.getState().assistantActions.checkResult).toEqual(result);
  });

  it("ignores a stale fulfilled result with an old request key", () => {
    const pendingState = assistantActionsReducer(
      initialAssistantActionsState,
      runGenerate.pending("request-1", {
        projectId: "project-1",
        contentId: "content-1",
        modelVariantId: "model-1",
        expectedRevision: 7,
        cursorByte: 12,
        instructions: "",
        requestKey: "new-key",
        requestToken: "token-1",
      }),
    );

    const state = assistantActionsReducer(
      pendingState,
      runGenerate.fulfilled({ session, candidate }, "request-2", {
        projectId: "project-1",
        contentId: "content-1",
        modelVariantId: "model-1",
        expectedRevision: 6,
        cursorByte: 12,
        instructions: "",
        requestKey: "old-key",
        requestToken: "token-1",
      }),
    );

    expect(state.status).toBe("running");
    expect(state.candidate).toBeNull();
  });

  it("ignores a stale fulfilled result with an old request token", () => {
    const pendingState = assistantActionsReducer(
      initialAssistantActionsState,
      runGenerate.pending("request-1", {
        projectId: "project-1",
        contentId: "content-1",
        modelVariantId: "model-1",
        expectedRevision: 7,
        cursorByte: 12,
        instructions: "",
        requestKey: "same-key",
        requestToken: "new-token",
      }),
    );

    const state = assistantActionsReducer(
      pendingState,
      runGenerate.fulfilled({ session, candidate }, "request-2", {
        projectId: "project-1",
        contentId: "content-1",
        modelVariantId: "model-1",
        expectedRevision: 7,
        cursorByte: 12,
        instructions: "",
        requestKey: "same-key",
        requestToken: "old-token",
      }),
    );

    expect(state.status).toBe("running");
    expect(state.candidate).toBeNull();
  });

  it("stores a rejected run error without keeping stale results", () => {
    const pendingState = assistantActionsReducer(
      { ...initialAssistantActionsState, candidate, checkResult: {} as ReadAndCheckResult },
      runGenerate.pending("request-1", {
        projectId: "project-1",
        contentId: "content-1",
        modelVariantId: "model-1",
        expectedRevision: 7,
        cursorByte: 12,
        instructions: "",
        requestKey: "same-key",
        requestToken: "token-1",
      }),
    );

    const state = assistantActionsReducer(
      pendingState,
      runGenerate.rejected(new Error("provider unavailable"), "request-1", {
        projectId: "project-1",
        contentId: "content-1",
        modelVariantId: "model-1",
        expectedRevision: 7,
        cursorByte: 12,
        instructions: "",
        requestKey: "same-key",
        requestToken: "token-1",
      }),
    );

    expect(state.status).toBe("failed");
    expect(state.error).toBe("provider unavailable");
    expect(state.candidate).toBeNull();
    expect(state.checkResult).toBeNull();
  });
});
