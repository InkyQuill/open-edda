import { createSlice } from "@reduxjs/toolkit";

export type AsyncStatus = "idle" | "pending" | "succeeded" | "failed";

export type ReviewState = {
  activityStatus: AsyncStatus;
  promptRecordsStatus: AsyncStatus;
  selectedPromptRecordId: string | null;
  error: string | null;
};

export const initialReviewState: ReviewState = {
  activityStatus: "idle",
  promptRecordsStatus: "idle",
  selectedPromptRecordId: null,
  error: null,
};

const reviewSlice = createSlice({
  name: "review",
  initialState: initialReviewState,
  reducers: {
    resetForProject() {
      return initialReviewState;
    },
  },
});

export const reviewActions = reviewSlice.actions;
export const reviewReducer = reviewSlice.reducer;
