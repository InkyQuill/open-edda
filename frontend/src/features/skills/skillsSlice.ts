import { createSlice } from "@reduxjs/toolkit";
import type { WriterSkill } from "../../skillTypes";
import { loadProjectSkills, loadSessionSkills, saveSessionSkills } from "./skillsThunks";

export type AsyncStatus = "idle" | "pending" | "succeeded" | "failed";

export type SkillsState = {
  projectId: string | null;
  skills: WriterSkill[];
  skillsStatus: AsyncStatus;
  skillsRequestId: string | null;
  sessionSkillsStatus: AsyncStatus;
  sessionSkillsRequestId: string | null;
  loadingSessionId: string | null;
  activeSessionId: string | null;
  selectedSkillIds: string[];
  confirmedSelectedSkillIds: string[];
  saveStatus: AsyncStatus;
  saveRequestId: string | null;
  savingSessionId: string | null;
  importStatus: AsyncStatus;
  error: string | null;
};

export const initialSkillsState: SkillsState = {
  projectId: null,
  skills: [],
  skillsStatus: "idle",
  skillsRequestId: null,
  sessionSkillsStatus: "idle",
  sessionSkillsRequestId: null,
  loadingSessionId: null,
  activeSessionId: null,
  selectedSkillIds: [],
  confirmedSelectedSkillIds: [],
  saveStatus: "idle",
  saveRequestId: null,
  savingSessionId: null,
  importStatus: "idle",
  error: null,
};

function clearProjectScopedState(state: SkillsState, projectId: string): void {
  if (state.projectId === projectId) return;
  state.projectId = projectId;
  state.skills = [];
  state.sessionSkillsStatus = "idle";
  state.sessionSkillsRequestId = null;
  state.loadingSessionId = null;
  state.activeSessionId = null;
  state.selectedSkillIds = [];
  state.confirmedSelectedSkillIds = [];
  state.saveStatus = "idle";
  state.saveRequestId = null;
  state.savingSessionId = null;
  state.importStatus = "idle";
}

const skillsSlice = createSlice({
  name: "skills",
  initialState: initialSkillsState,
  reducers: {
    toggleSelectedSkillId(state, action: { payload: string }) {
      if (state.selectedSkillIds.includes(action.payload)) {
        state.selectedSkillIds = state.selectedSkillIds.filter((skillId) => skillId !== action.payload);
      } else {
        state.selectedSkillIds.push(action.payload);
      }
    },
    replaceSelectedSkillIds(state, action: { payload: string[] }) {
      state.selectedSkillIds = action.payload;
    },
    clearSessionSelection(state) {
      state.activeSessionId = null;
      state.sessionSkillsStatus = "idle";
      state.sessionSkillsRequestId = null;
      state.loadingSessionId = null;
      state.selectedSkillIds = [];
      state.confirmedSelectedSkillIds = [];
      state.saveStatus = "idle";
      state.saveRequestId = null;
      state.savingSessionId = null;
      state.error = null;
    },
    resetForProject() {
      return initialSkillsState;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loadProjectSkills.pending, (state, action) => {
        clearProjectScopedState(state, action.meta.arg.projectId);
        state.skillsStatus = "pending";
        state.skillsRequestId = action.meta.requestId;
        state.error = null;
      })
      .addCase(loadProjectSkills.fulfilled, (state, action) => {
        if (state.projectId !== action.meta.arg.projectId || state.skillsRequestId !== action.meta.requestId) {
          return;
        }
        state.skills = action.payload;
        state.skillsStatus = "succeeded";
        state.skillsRequestId = null;
        state.error = null;
      })
      .addCase(loadProjectSkills.rejected, (state, action) => {
        if (state.projectId !== action.meta.arg.projectId || state.skillsRequestId !== action.meta.requestId) {
          return;
        }
        state.skillsStatus = "failed";
        state.skillsRequestId = null;
        state.error = action.error.message ?? "Could not load skills";
      })
      .addCase(loadSessionSkills.pending, (state, action) => {
        clearProjectScopedState(state, action.meta.arg.projectId);
        state.sessionSkillsStatus = "pending";
        state.sessionSkillsRequestId = action.meta.requestId;
        state.loadingSessionId = action.meta.arg.sessionId;
        state.activeSessionId = action.meta.arg.sessionId;
        if (state.savingSessionId !== action.meta.arg.sessionId) {
          state.saveStatus = "idle";
          state.saveRequestId = null;
          state.savingSessionId = null;
        }
        state.error = null;
      })
      .addCase(loadSessionSkills.fulfilled, (state, action) => {
        if (
          state.projectId !== action.meta.arg.projectId ||
          state.sessionSkillsRequestId !== action.meta.requestId ||
          state.loadingSessionId !== action.meta.arg.sessionId
        ) {
          return;
        }
        const selectedSkillIds = action.payload.map((skill) => skill.id);
        state.activeSessionId = action.meta.arg.sessionId;
        state.selectedSkillIds = selectedSkillIds;
        state.confirmedSelectedSkillIds = selectedSkillIds;
        state.sessionSkillsStatus = "succeeded";
        state.sessionSkillsRequestId = null;
        state.loadingSessionId = null;
        state.error = null;
      })
      .addCase(loadSessionSkills.rejected, (state, action) => {
        if (
          state.projectId !== action.meta.arg.projectId ||
          state.sessionSkillsRequestId !== action.meta.requestId ||
          state.loadingSessionId !== action.meta.arg.sessionId
        ) {
          return;
        }
        state.sessionSkillsStatus = "failed";
        state.sessionSkillsRequestId = null;
        state.loadingSessionId = null;
        state.selectedSkillIds = [];
        state.confirmedSelectedSkillIds = [];
        state.error = action.error.message ?? "Could not load session skills";
      })
      .addCase(saveSessionSkills.pending, (state, action) => {
        clearProjectScopedState(state, action.meta.arg.projectId);
        state.saveStatus = "pending";
        state.saveRequestId = action.meta.requestId;
        state.savingSessionId = action.meta.arg.sessionId;
        state.error = null;
      })
      .addCase(saveSessionSkills.fulfilled, (state, action) => {
        if (
          state.projectId !== action.meta.arg.projectId ||
          state.saveRequestId !== action.meta.requestId ||
          state.savingSessionId !== action.meta.arg.sessionId ||
          state.activeSessionId !== action.meta.arg.sessionId
        ) {
          return;
        }
        const selectedSkillIds = action.payload.map((skill) => skill.id);
        state.selectedSkillIds = selectedSkillIds;
        state.confirmedSelectedSkillIds = selectedSkillIds;
        state.saveStatus = "succeeded";
        state.saveRequestId = null;
        state.savingSessionId = null;
        state.error = null;
      })
      .addCase(saveSessionSkills.rejected, (state, action) => {
        if (
          state.projectId !== action.meta.arg.projectId ||
          state.saveRequestId !== action.meta.requestId ||
          state.savingSessionId !== action.meta.arg.sessionId ||
          state.activeSessionId !== action.meta.arg.sessionId
        ) {
          return;
        }
        state.saveStatus = "failed";
        state.saveRequestId = null;
        state.savingSessionId = null;
        state.selectedSkillIds = state.confirmedSelectedSkillIds;
        state.error = action.error.message ?? "Could not save session skills";
      });
  },
});

export const skillsActions = skillsSlice.actions;
export const skillsReducer = skillsSlice.reducer;
