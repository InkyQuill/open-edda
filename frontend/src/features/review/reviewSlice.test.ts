import { describe, expect, it } from "vitest";
import type { ActivityEvent, PromptRecord } from "../../agentTypes";
import { loadPromptRecords, loadReviewActivity } from "./reviewThunks";
import { initialReviewState, reviewActions, reviewReducer, type ReviewState } from "./reviewSlice";

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

  it("ignores stale activity loads from a previous project", () => {
    const loadingFirst = reviewReducer(
      initialReviewState,
      loadReviewActivity.pending("request-1", { projectId: "project-1" }),
    );
    const loadingSecond = reviewReducer(
      loadingFirst,
      loadReviewActivity.pending("request-2", { projectId: "project-2" }),
    );
    const loadedSecond = reviewReducer(
      loadingSecond,
      loadReviewActivity.fulfilled([{ ...activityEvent, id: "activity-2", projectId: "project-2" }], "request-2", {
        projectId: "project-2",
      }),
    );

    const staleLoadedFirst = reviewReducer(
      loadedSecond,
      loadReviewActivity.fulfilled([activityEvent], "request-1", {
        projectId: "project-1",
      }),
    );

    expect(staleLoadedFirst.activityEvents).toEqual([
      { ...activityEvent, id: "activity-2", projectId: "project-2" },
    ]);
    expect(staleLoadedFirst.projectId).toBe("project-2");
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

  it("ignores stale prompt record loads from a previous project", () => {
    const loadingFirst = reviewReducer(
      initialReviewState,
      loadPromptRecords.pending("request-1", { projectId: "project-1" }),
    );
    const loadingSecond = reviewReducer(
      loadingFirst,
      loadPromptRecords.pending("request-2", { projectId: "project-2" }),
    );
    const projectTwoPromptRecord = { ...promptRecord, id: "prompt-record-2", projectId: "project-2" };
    const loadedSecond = reviewReducer(
      loadingSecond,
      loadPromptRecords.fulfilled([projectTwoPromptRecord], "request-2", {
        projectId: "project-2",
      }),
    );

    const staleLoadedFirst = reviewReducer(
      loadedSecond,
      loadPromptRecords.fulfilled([promptRecord], "request-1", {
        projectId: "project-1",
      }),
    );

    expect(staleLoadedFirst.promptRecords).toEqual([projectTwoPromptRecord]);
    expect(staleLoadedFirst.projectId).toBe("project-2");
    expect(staleLoadedFirst.promptRecordsStatus).toBe("succeeded");
  });

  it("clears activity events when activity is requested for a new project", () => {
    const previousState: ReviewState = {
      ...initialReviewState,
      projectId: "project-1",
      activityEvents: [activityEvent],
      promptRecords: [promptRecord],
      selectedPromptRecordId: promptRecord.id,
    };

    const loading = reviewReducer(
      previousState,
      loadReviewActivity.pending("request-2", { projectId: "project-2" }),
    );

    expect(loading.activityEvents).toEqual([]);
    expect(loading.promptRecords).toEqual([promptRecord]);
    expect(loading.selectedPromptRecordId).toBe(promptRecord.id);
    expect(loading.projectId).toBe("project-2");
    expect(loading.activityStatus).toBe("pending");
  });

  it("clears prompt records and selection when prompts are requested for a new project", () => {
    const previousState: ReviewState = {
      ...initialReviewState,
      projectId: "project-1",
      activityEvents: [activityEvent],
      promptRecords: [promptRecord],
      selectedPromptRecordId: promptRecord.id,
    };

    const loading = reviewReducer(
      previousState,
      loadPromptRecords.pending("request-2", { projectId: "project-2" }),
    );

    expect(loading.promptRecords).toEqual([]);
    expect(loading.selectedPromptRecordId).toBeNull();
    expect(loading.activityEvents).toEqual([activityEvent]);
    expect(loading.projectId).toBe("project-2");
    expect(loading.promptRecordsStatus).toBe("pending");
  });

  it("keeps existing data when refetching the same project", () => {
    const previousState: ReviewState = {
      ...initialReviewState,
      projectId: "project-1",
      activityEvents: [activityEvent],
      promptRecords: [promptRecord],
      selectedPromptRecordId: promptRecord.id,
    };

    const loadingActivity = reviewReducer(
      previousState,
      loadReviewActivity.pending("activity-request-2", { projectId: "project-1" }),
    );
    const loadingPrompts = reviewReducer(
      previousState,
      loadPromptRecords.pending("prompt-request-2", { projectId: "project-1" }),
    );

    expect(loadingActivity.activityEvents).toEqual([activityEvent]);
    expect(loadingPrompts.promptRecords).toEqual([promptRecord]);
    expect(loadingPrompts.selectedPromptRecordId).toBe(promptRecord.id);
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
