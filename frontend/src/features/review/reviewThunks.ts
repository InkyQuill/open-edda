import { createAsyncThunk } from "@reduxjs/toolkit";

import { listActivity, listPromptRecords } from "../../agentApi";

export const loadReviewActivity = createAsyncThunk(
  "review/loadActivity",
  async ({ projectId }: { projectId: string }, { signal }) => listActivity(projectId, 50, signal),
);

export const loadPromptRecords = createAsyncThunk(
  "review/loadPromptRecords",
  async ({ projectId }: { projectId: string }, { signal }) => listPromptRecords(projectId, 50, signal),
);
