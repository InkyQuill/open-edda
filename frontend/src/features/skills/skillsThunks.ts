import { createAsyncThunk } from "@reduxjs/toolkit";

import { listSessionSkills, listSkills, selectSessionSkills } from "../../skillApi";

export const loadProjectSkills = createAsyncThunk(
  "skills/loadProjectSkills",
  async ({ projectId }: { projectId: string }, { signal }) => listSkills(projectId, signal),
);

export const loadSessionSkills = createAsyncThunk(
  "skills/loadSessionSkills",
  async ({ projectId, sessionId }: { projectId: string; sessionId: string }, { signal }) =>
    listSessionSkills(projectId, sessionId, signal),
);

export const saveSessionSkills = createAsyncThunk(
  "skills/saveSessionSkills",
  async (
    {
      projectId,
      sessionId,
      skillIds,
    }: {
      projectId: string;
      sessionId: string;
      skillIds: string[];
    },
    { signal },
  ) => selectSessionSkills(projectId, sessionId, skillIds, signal),
);
