import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

import type { ActivityEvent, PromptRecord } from "../../agentTypes";
import { loadPromptRecords, loadReviewActivity } from "./reviewThunks";

export type AsyncStatus = "idle" | "pending" | "succeeded" | "failed";

export type ReviewState = {
  projectId: string | null;
  activityEvents: ActivityEvent[];
  promptRecords: PromptRecord[];
  activityStatus: AsyncStatus;
  promptRecordsStatus: AsyncStatus;
  activityRequestId: string | null;
  promptRecordsRequestId: string | null;
  selectedPromptRecordId: string | null;
  error: string | null;
};

export const initialReviewState: ReviewState = {
  projectId: null,
  activityEvents: [],
  promptRecords: [],
  activityStatus: "idle",
  promptRecordsStatus: "idle",
  activityRequestId: null,
  promptRecordsRequestId: null,
  selectedPromptRecordId: null,
  error: null,
};

function readableError(action: { error: { message?: string } }): string {
  return action.error.message ?? "Could not load review data";
}

const reviewSlice = createSlice({
  name: "review",
  initialState: initialReviewState,
  reducers: {
    setSelectedPromptRecordId(state, action: PayloadAction<string | null>) {
      state.selectedPromptRecordId = action.payload;
    },
    resetForProject() {
      return initialReviewState;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loadReviewActivity.pending, (state, action) => {
        if (state.projectId !== action.meta.arg.projectId) {
          state.activityEvents = [];
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
          state.promptRecords = [];
          state.selectedPromptRecordId = null;
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
      });
  },
});

export const reviewActions = reviewSlice.actions;
export const reviewReducer = reviewSlice.reducer;
