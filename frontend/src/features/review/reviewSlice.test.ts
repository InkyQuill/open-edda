import { describe, expect, it } from "vitest";
import type { ActivityEvent, PromptRecord } from "../../agentTypes";
import { loadPromptRecords, loadReviewActivity } from "./reviewThunks";
import { initialReviewState, reviewActions, reviewReducer } from "./reviewSlice";

const activityEvent: ActivityEvent = {
  id: "activity-1",
  projectId: "project-1",
  sessionId: "session-1",
  eventType: "review.check",
  summary: "Checked chapter opening",
  metadataJson: "{}",
  createdAt: "2026-06-14T10:00:00.000Z",
};

const promptRecord: PromptRecord = {
  id: "prompt-record-1",
  projectId: "project-1",
  sessionId: "session-1",
  providerName: "Provider",
  modelName: "model-a",
  actionKind: "read_check",
  requestJson: "{}",
  responseJson: "{}",
  usage: {
    inputTokens: 100,
    outputTokens: 50,
    cacheReadTokens: 0,
    cacheWriteTokens: 0,
    totalTokens: 150,
    inputCost: 0.01,
    outputCost: 0.02,
    cacheReadCost: 0,
    cacheWriteCost: 0,
    totalCost: 0.03,
  },
  createdAt: "2026-06-14T10:01:00.000Z",
};

describe("reviewSlice", () => {
  it("stores activity events for the active project", () => {
    const loading = reviewReducer(
      initialReviewState,
      loadReviewActivity.pending("request-1", { projectId: "project-1" }),
    );

    const loaded = reviewReducer(
      loading,
      loadReviewActivity.fulfilled([activityEvent], "request-1", { projectId: "project-1" }),
    );

    expect(loaded.activityEvents).toEqual([activityEvent]);
    expect(loaded.activityStatus).toBe("succeeded");
    expect(loaded.activityRequestId).toBeNull();
    expect(loaded.projectId).toBe("project-1");
    expect(loaded.error).toBeNull();
  });

  it("stores prompt records for the active project", () => {
    const loading = reviewReducer(
      initialReviewState,
      loadPromptRecords.pending("request-1", { projectId: "project-1" }),
    );

    const loaded = reviewReducer(
      loading,
      loadPromptRecords.fulfilled([promptRecord], "request-1", { projectId: "project-1" }),
    );

    expect(loaded.promptRecords).toEqual([promptRecord]);
    expect(loaded.promptRecordsStatus).toBe("succeeded");
    expect(loaded.promptRecordsRequestId).toBeNull();
    expect(loaded.projectId).toBe("project-1");
    expect(loaded.error).toBeNull();
  });

  it("ignores stale activity loads", () => {
    const loadingFirst = reviewReducer(
      initialReviewState,
      loadReviewActivity.pending("request-1", { projectId: "project-1" }),
    );
    const loadingSecond = reviewReducer(
      loadingFirst,
      loadReviewActivity.pending("request-2", { projectId: "project-1" }),
    );
    const loadedSecond = reviewReducer(
      loadingSecond,
      loadReviewActivity.fulfilled([activityEvent], "request-2", { projectId: "project-1" }),
    );

    const staleLoadedFirst = reviewReducer(
      loadedSecond,
      loadReviewActivity.fulfilled([{ ...activityEvent, id: "activity-stale" }], "request-1", {
        projectId: "project-1",
      }),
    );

    expect(staleLoadedFirst.activityEvents).toEqual([activityEvent]);
    expect(staleLoadedFirst.activityStatus).toBe("succeeded");
  });

  it("ignores stale prompt record loads", () => {
    const loadingFirst = reviewReducer(
      initialReviewState,
      loadPromptRecords.pending("request-1", { projectId: "project-1" }),
    );
    const loadingSecond = reviewReducer(
      loadingFirst,
      loadPromptRecords.pending("request-2", { projectId: "project-1" }),
    );
    const loadedSecond = reviewReducer(
      loadingSecond,
      loadPromptRecords.fulfilled([promptRecord], "request-2", { projectId: "project-1" }),
    );

    const staleLoadedFirst = reviewReducer(
      loadedSecond,
      loadPromptRecords.fulfilled([{ ...promptRecord, id: "prompt-record-stale" }], "request-1", {
        projectId: "project-1",
      }),
    );

    expect(staleLoadedFirst.promptRecords).toEqual([promptRecord]);
    expect(staleLoadedFirst.promptRecordsStatus).toBe("succeeded");
  });

  it("selects a prompt record", () => {
    const selected = reviewReducer(initialReviewState, reviewActions.setSelectedPromptRecordId("prompt-record-1"));

    expect(selected.selectedPromptRecordId).toBe("prompt-record-1");
  });

  it("resets project-scoped review state", () => {
    const loaded = reviewReducer(
      {
        ...initialReviewState,
        projectId: "project-1",
        activityEvents: [activityEvent],
        promptRecords: [promptRecord],
        activityStatus: "succeeded",
        promptRecordsStatus: "failed",
        activityRequestId: "activity-request",
        promptRecordsRequestId: "prompt-records-request",
        selectedPromptRecordId: "prompt-record-1",
        error: "Could not load prompt records",
      },
      reviewActions.resetForProject(),
    );

    expect(loaded).toEqual(initialReviewState);
  });
});
