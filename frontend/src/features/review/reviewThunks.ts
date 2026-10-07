import { createAsyncThunk } from "@reduxjs/toolkit";

import { listRevisions, restoreRevision } from "../../api";
import { listActivity, listPromptRecords } from "../../agentApi";

export const loadReviewActivity = createAsyncThunk(
  "review/loadActivity",
  async ({ projectId }: { projectId: string }, { signal }) => listActivity(projectId, 50, signal),
);

export const loadPromptRecords = createAsyncThunk(
  "review/loadPromptRecords",
  async ({ projectId }: { projectId: string }, { signal }) => listPromptRecords(projectId, 50, signal),
);

export const loadContentRevisions = createAsyncThunk(
  "review/loadContentRevisions",
  async ({ projectId, contentId }: { projectId: string; contentId: string }, { signal }) =>
    listRevisions(projectId, contentId, signal),
);

export const restoreContentRevision = createAsyncThunk(
  "review/restoreContentRevision",
  async (
    {
      projectId,
      contentId,
      revisionNumber,
      expectedRevision,
    }: { projectId: string; contentId: string; revisionNumber: number; expectedRevision: number },
    { signal },
  ) =>
    restoreRevision(projectId, contentId, revisionNumber, {
      expectedRevision,
      reason: `restore checkpoint ${revisionNumber}`,
    }, signal),
);
