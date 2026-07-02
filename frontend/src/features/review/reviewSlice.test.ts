import { describe, expect, it } from "vitest";
import type { ActivityEvent, PromptRecord } from "../../agentTypes";
import type { ContentItem, Revision } from "../../types";
import { loadContentRevisions, loadPromptRecords, loadReviewActivity, restoreContentRevision } from "./reviewThunks";
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

const revision: Revision = {
  id: "revision-2",
  contentItemId: "content-1",
  revisionNumber: 2,
  bodyMarkdown: "The lantern burned green.",
  metadataJson: "{}",
  reason: "manual revision",
  createdBy: "author",
  createdAt: "2026-06-14T10:02:00.000Z",
};

const restoredContent: ContentItem = {
  id: "content-1",
  projectId: "project-1",
  kind: "chapter",
  title: "Opening",
  slug: "opening",
  bodyMarkdown: "The lantern burned blue.",
  metadataJson: "{}",
  sortOrder: 1,
  currentRevision: 3,
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

  it("stores revisions for the active project content", () => {
    const loading = reviewReducer(
      initialReviewState,
      loadContentRevisions.pending("request-1", { projectId: "project-1", contentId: "content-1" }),
    );

    const loaded = reviewReducer(
      loading,
      loadContentRevisions.fulfilled([revision], "request-1", { projectId: "project-1", contentId: "content-1" }),
    );

    expect(loaded.revisions).toEqual([revision]);
    expect(loaded.revisionsStatus).toBe("succeeded");
    expect(loaded.revisionsRequestId).toBeNull();
    expect(loaded.projectId).toBe("project-1");
    expect(loaded.contentId).toBe("content-1");
    expect(loaded.revisionsError).toBeNull();
    expect(loaded.restoreError).toBeNull();
  });

  it("stores revision load failures separately from restore failures", () => {
    const loading = reviewReducer(
      initialReviewState,
      loadContentRevisions.pending("request-1", { projectId: "project-1", contentId: "content-1" }),
    );

    const failed = reviewReducer(
      loading,
      loadContentRevisions.rejected(new Error("list revisions failed"), "request-1", {
        projectId: "project-1",
        contentId: "content-1",
      }),
    );

    expect(failed.revisionsStatus).toBe("failed");
    expect(failed.revisionsError).toBe("list revisions failed");
    expect(failed.restoreError).toBeNull();
  });

  it("clears stale restore errors when revisions reload", () => {
    const loading = reviewReducer(
      {
        ...initialReviewState,
        projectId: "project-1",
        contentId: "content-1",
        restoreStatus: "failed",
        restoreError: "expected revision conflict",
        restoreErrorCode: "HTTP_409",
      },
      loadContentRevisions.pending("request-1", { projectId: "project-1", contentId: "content-1" }),
    );

    expect(loading.restoreError).toBeNull();
    expect(loading.restoreErrorCode).toBeNull();
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

  it("ignores stale revision loads for a previous content item", () => {
    const loadingFirst = reviewReducer(
      initialReviewState,
      loadContentRevisions.pending("request-1", { projectId: "project-1", contentId: "content-1" }),
    );
    const loadingSecond = reviewReducer(
      loadingFirst,
      loadContentRevisions.pending("request-2", { projectId: "project-1", contentId: "content-2" }),
    );
    const contentTwoRevision = { ...revision, id: "revision-content-2", contentItemId: "content-2" };
    const loadedSecond = reviewReducer(
      loadingSecond,
      loadContentRevisions.fulfilled([contentTwoRevision], "request-2", {
        projectId: "project-1",
        contentId: "content-2",
      }),
    );

    const staleLoadedFirst = reviewReducer(
      loadedSecond,
      loadContentRevisions.fulfilled([revision], "request-1", {
        projectId: "project-1",
        contentId: "content-1",
      }),
    );

    expect(staleLoadedFirst.revisions).toEqual([contentTwoRevision]);
    expect(staleLoadedFirst.contentId).toBe("content-2");
    expect(staleLoadedFirst.revisionsStatus).toBe("succeeded");
  });

  it("clears project-scoped review data when activity is requested for a new project", () => {
    const previousState: ReviewState = {
      ...initialReviewState,
      projectId: "project-1",
      contentId: "content-1",
      activityEvents: [activityEvent],
      promptRecords: [promptRecord],
      revisions: [revision],
      selectedPromptRecordId: promptRecord.id,
      selectedRevisionNumber: revision.revisionNumber,
      error: "Could not load review data",
      restoreError: "Could not restore checkpoint",
    };

    const loading = reviewReducer(
      previousState,
      loadReviewActivity.pending("request-2", { projectId: "project-2" }),
    );

    expect(loading.activityEvents).toEqual([]);
    expect(loading.promptRecords).toEqual([]);
    expect(loading.revisions).toEqual([]);
    expect(loading.selectedPromptRecordId).toBeNull();
    expect(loading.selectedRevisionNumber).toBeNull();
    expect(loading.error).toBeNull();
    expect(loading.restoreError).toBeNull();
    expect(loading.projectId).toBe("project-2");
    expect(loading.contentId).toBeNull();
    expect(loading.activityStatus).toBe("pending");
  });

  it("clears project-scoped review data when prompts are requested for a new project", () => {
    const previousState: ReviewState = {
      ...initialReviewState,
      projectId: "project-1",
      contentId: "content-1",
      activityEvents: [activityEvent],
      promptRecords: [promptRecord],
      revisions: [revision],
      selectedPromptRecordId: promptRecord.id,
      selectedRevisionNumber: revision.revisionNumber,
      error: "Could not load review data",
      restoreError: "Could not restore checkpoint",
    };

    const loading = reviewReducer(
      previousState,
      loadPromptRecords.pending("request-2", { projectId: "project-2" }),
    );

    expect(loading.promptRecords).toEqual([]);
    expect(loading.selectedPromptRecordId).toBeNull();
    expect(loading.activityEvents).toEqual([]);
    expect(loading.revisions).toEqual([]);
    expect(loading.selectedRevisionNumber).toBeNull();
    expect(loading.error).toBeNull();
    expect(loading.restoreError).toBeNull();
    expect(loading.projectId).toBe("project-2");
    expect(loading.contentId).toBeNull();
    expect(loading.promptRecordsStatus).toBe("pending");
  });

  it("clears content-scoped review data when revisions are requested for a new content item", () => {
    const previousState: ReviewState = {
      ...initialReviewState,
      projectId: "project-1",
      contentId: "content-1",
      activityEvents: [activityEvent],
      promptRecords: [promptRecord],
      revisions: [revision],
      selectedPromptRecordId: promptRecord.id,
      selectedRevisionNumber: revision.revisionNumber,
      restoreError: "Could not restore checkpoint",
    };

    const loading = reviewReducer(
      previousState,
      loadContentRevisions.pending("request-2", { projectId: "project-1", contentId: "content-2" }),
    );

    expect(loading.activityEvents).toEqual([activityEvent]);
    expect(loading.promptRecords).toEqual([promptRecord]);
    expect(loading.selectedPromptRecordId).toBe(promptRecord.id);
    expect(loading.revisions).toEqual([]);
    expect(loading.selectedRevisionNumber).toBeNull();
    expect(loading.restoreError).toBeNull();
    expect(loading.projectId).toBe("project-1");
    expect(loading.contentId).toBe("content-2");
    expect(loading.revisionsStatus).toBe("pending");
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

  it("selects a revision number", () => {
    const selected = reviewReducer(initialReviewState, reviewActions.setSelectedRevisionNumber(2));

    expect(selected.selectedRevisionNumber).toBe(2);
  });

  it("tracks successful restore for the active content item", () => {
    const loading = reviewReducer(
      {
        ...initialReviewState,
        projectId: "project-1",
        contentId: "content-1",
        revisions: [revision],
      },
      restoreContentRevision.pending("request-1", {
        projectId: "project-1",
        contentId: "content-1",
        revisionNumber: 1,
        expectedRevision: 2,
      }),
    );

    const restored = reviewReducer(
      loading,
      restoreContentRevision.fulfilled(restoredContent, "request-1", {
        projectId: "project-1",
        contentId: "content-1",
        revisionNumber: 1,
        expectedRevision: 2,
      }),
    );

    expect(restored.restoreStatus).toBe("succeeded");
    expect(restored.restoreRequestId).toBeNull();
    expect(restored.restoreError).toBeNull();
    expect(restored.selectedRevisionNumber).toBe(restoredContent.currentRevision);
    expect(restored.revisions).toEqual([revision]);
  });

  it("tracks restore conflicts for the active content item", () => {
    const loading = reviewReducer(
      {
        ...initialReviewState,
        projectId: "project-1",
        contentId: "content-1",
      },
      restoreContentRevision.pending("request-1", {
        projectId: "project-1",
        contentId: "content-1",
        revisionNumber: 1,
        expectedRevision: 2,
      }),
    );

    const failed = reviewReducer(
      loading,
      restoreContentRevision.rejected(new Error("restore revision failed: 409"), "request-1", {
        projectId: "project-1",
        contentId: "content-1",
        revisionNumber: 1,
        expectedRevision: 2,
      }),
    );

    expect(failed.restoreStatus).toBe("failed");
    expect(failed.restoreRequestId).toBeNull();
    expect(failed.restoreError).toBe("restore revision failed: 409");
    expect(failed.restoreErrorCode).toBeNull();
  });

  it("tracks structured restore conflict codes", () => {
    const loading = reviewReducer(
      {
        ...initialReviewState,
        projectId: "project-1",
        contentId: "content-1",
      },
      restoreContentRevision.pending("request-1", {
        projectId: "project-1",
        contentId: "content-1",
        revisionNumber: 1,
        expectedRevision: 2,
      }),
    );

    const failed = reviewReducer(
      loading,
      restoreContentRevision.rejected(Object.assign(new Error("expected revision conflict"), { code: "HTTP_409" }), "request-1", {
        projectId: "project-1",
        contentId: "content-1",
        revisionNumber: 1,
        expectedRevision: 2,
      }),
    );

    expect(failed.restoreError).toBe("expected revision conflict");
    expect(failed.restoreErrorCode).toBe("HTTP_409");
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
