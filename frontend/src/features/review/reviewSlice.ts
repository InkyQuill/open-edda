import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

import type { ActivityEvent, PromptRecord } from "../../agentTypes";
import type { Revision } from "../../types";
import { loadContentRevisions, loadPromptRecords, loadReviewActivity, restoreContentRevision } from "./reviewThunks";

export type AsyncStatus = "idle" | "pending" | "succeeded" | "failed";

export type ReviewState = {
  projectId: string | null;
  contentId: string | null;
  activityEvents: ActivityEvent[];
  promptRecords: PromptRecord[];
  revisions: Revision[];
  activityStatus: AsyncStatus;
  promptRecordsStatus: AsyncStatus;
  revisionsStatus: AsyncStatus;
  restoreStatus: AsyncStatus;
  activityRequestId: string | null;
  promptRecordsRequestId: string | null;
  revisionsRequestId: string | null;
  restoreRequestId: string | null;
  selectedPromptRecordId: string | null;
  selectedRevisionNumber: number | null;
  error: string | null;
  restoreError: string | null;
};

export const initialReviewState: ReviewState = {
  projectId: null,
  contentId: null,
  activityEvents: [],
  promptRecords: [],
  revisions: [],
  activityStatus: "idle",
  promptRecordsStatus: "idle",
  revisionsStatus: "idle",
  restoreStatus: "idle",
  activityRequestId: null,
  promptRecordsRequestId: null,
  revisionsRequestId: null,
  restoreRequestId: null,
  selectedPromptRecordId: null,
  selectedRevisionNumber: null,
  error: null,
  restoreError: null,
};

function readableError(action: { error: { message?: string } }): string {
  return action.error.message ?? "Could not load review data";
}

function clearProjectScopedReviewState(state: ReviewState) {
  state.activityEvents = [];
  state.promptRecords = [];
  state.selectedPromptRecordId = null;
  state.error = null;
  clearContentScopedReviewState(state);
}

function clearContentScopedReviewState(state: ReviewState) {
  state.contentId = null;
  state.revisions = [];
  state.revisionsStatus = "idle";
  state.restoreStatus = "idle";
  state.revisionsRequestId = null;
  state.restoreRequestId = null;
  state.selectedRevisionNumber = null;
  state.restoreError = null;
}

const reviewSlice = createSlice({
  name: "review",
  initialState: initialReviewState,
  reducers: {
    setSelectedPromptRecordId(state, action: PayloadAction<string | null>) {
      state.selectedPromptRecordId = action.payload;
    },
    setSelectedRevisionNumber(state, action: PayloadAction<number | null>) {
      state.selectedRevisionNumber = action.payload;
    },
    resetForProject() {
      return initialReviewState;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loadReviewActivity.pending, (state, action) => {
        if (state.projectId !== action.meta.arg.projectId) {
          clearProjectScopedReviewState(state);
        }
        state.projectId = action.meta.arg.projectId;
        state.activityStatus = "pending";
        state.activityRequestId = action.meta.requestId;
        state.error = null;
      })
      .addCase(loadReviewActivity.fulfilled, (state, action) => {
        if (state.projectId !== action.meta.arg.projectId || state.activityRequestId !== action.meta.requestId) {
          return;
        }
        state.activityEvents = action.payload;
        state.activityStatus = "succeeded";
        state.activityRequestId = null;
      })
      .addCase(loadReviewActivity.rejected, (state, action) => {
        if (state.projectId !== action.meta.arg.projectId || state.activityRequestId !== action.meta.requestId) {
          return;
        }
        state.activityStatus = "failed";
        state.activityRequestId = null;
        state.error = readableError(action);
      })
      .addCase(loadPromptRecords.pending, (state, action) => {
        if (state.projectId !== action.meta.arg.projectId) {
          clearProjectScopedReviewState(state);
        }
        state.projectId = action.meta.arg.projectId;
        state.promptRecordsStatus = "pending";
        state.promptRecordsRequestId = action.meta.requestId;
        state.error = null;
      })
      .addCase(loadPromptRecords.fulfilled, (state, action) => {
        if (state.projectId !== action.meta.arg.projectId || state.promptRecordsRequestId !== action.meta.requestId) {
          return;
        }
        state.promptRecords = action.payload;
        state.promptRecordsStatus = "succeeded";
        state.promptRecordsRequestId = null;
        if (
          state.selectedPromptRecordId &&
          !action.payload.some((promptRecord) => promptRecord.id === state.selectedPromptRecordId)
        ) {
          state.selectedPromptRecordId = null;
        }
      })
      .addCase(loadPromptRecords.rejected, (state, action) => {
        if (state.projectId !== action.meta.arg.projectId || state.promptRecordsRequestId !== action.meta.requestId) {
          return;
        }
        state.promptRecordsStatus = "failed";
        state.promptRecordsRequestId = null;
        state.error = readableError(action);
      })
      .addCase(loadContentRevisions.pending, (state, action) => {
        if (state.projectId !== action.meta.arg.projectId) {
          clearProjectScopedReviewState(state);
        } else if (state.contentId !== action.meta.arg.contentId) {
          clearContentScopedReviewState(state);
        }
        state.projectId = action.meta.arg.projectId;
        state.contentId = action.meta.arg.contentId;
        state.revisionsStatus = "pending";
        state.revisionsRequestId = action.meta.requestId;
        state.restoreError = null;
      })
      .addCase(loadContentRevisions.fulfilled, (state, action) => {
        if (
          state.projectId !== action.meta.arg.projectId ||
          state.contentId !== action.meta.arg.contentId ||
          state.revisionsRequestId !== action.meta.requestId
        ) {
          return;
        }
        state.revisions = action.payload;
        state.revisionsStatus = "succeeded";
        state.revisionsRequestId = null;
        if (
          state.selectedRevisionNumber &&
          !action.payload.some((revision) => revision.revisionNumber === state.selectedRevisionNumber)
        ) {
          state.selectedRevisionNumber = null;
        }
      })
      .addCase(loadContentRevisions.rejected, (state, action) => {
        if (
          state.projectId !== action.meta.arg.projectId ||
          state.contentId !== action.meta.arg.contentId ||
          state.revisionsRequestId !== action.meta.requestId
        ) {
          return;
        }
        state.revisionsStatus = "failed";
        state.revisionsRequestId = null;
        state.restoreError = readableError(action);
      })
      .addCase(restoreContentRevision.pending, (state, action) => {
        if (state.projectId !== action.meta.arg.projectId) {
          clearProjectScopedReviewState(state);
        } else if (state.contentId !== action.meta.arg.contentId) {
          clearContentScopedReviewState(state);
        }
        state.projectId = action.meta.arg.projectId;
        state.contentId = action.meta.arg.contentId;
        state.selectedRevisionNumber = action.meta.arg.revisionNumber;
        state.restoreStatus = "pending";
        state.restoreRequestId = action.meta.requestId;
        state.restoreError = null;
      })
      .addCase(restoreContentRevision.fulfilled, (state, action) => {
        if (
          state.projectId !== action.meta.arg.projectId ||
          state.contentId !== action.meta.arg.contentId ||
          state.restoreRequestId !== action.meta.requestId
        ) {
          return;
        }
        state.restoreStatus = "succeeded";
        state.restoreRequestId = null;
        state.restoreError = null;
        state.selectedRevisionNumber = action.payload.currentRevision;
      })
      .addCase(restoreContentRevision.rejected, (state, action) => {
        if (
          state.projectId !== action.meta.arg.projectId ||
          state.contentId !== action.meta.arg.contentId ||
          state.restoreRequestId !== action.meta.requestId
        ) {
          return;
        }
        state.restoreStatus = "failed";
        state.restoreRequestId = null;
        state.restoreError = readableError(action);
      });
  },
});

export const reviewActions = reviewSlice.actions;
export const reviewReducer = reviewSlice.reducer;
