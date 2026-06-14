import { createAsyncThunk } from "@reduxjs/toolkit";

import { listSessionSkills, listSkills, selectSessionSkills } from "../../skillApi";

export const loadProjectSkills = createAsyncThunk(
  "skills/loadProjectSkills",
  async ({ projectId }: { projectId: string }) => listSkills(projectId),
);

export const loadSessionSkills = createAsyncThunk(
  "skills/loadSessionSkills",
  async ({ projectId, sessionId }: { projectId: string; sessionId: string }) =>
    listSessionSkills(projectId, sessionId),
);

export const saveSessionSkills = createAsyncThunk(
  "skills/saveSessionSkills",
  async ({
    projectId,
    sessionId,
    skillIds,
  }: {
    projectId: string;
    sessionId: string;
    skillIds: string[];
  }) => selectSessionSkills(projectId, sessionId, skillIds),
);
