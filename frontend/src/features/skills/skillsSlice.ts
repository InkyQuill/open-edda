import { createSlice } from "@reduxjs/toolkit";

export type AsyncStatus = "idle" | "pending" | "succeeded" | "failed";

export type SkillsState = {
  skillsStatus: AsyncStatus;
  selectedSkillIds: string[];
  importStatus: AsyncStatus;
  error: string | null;
};

export const initialSkillsState: SkillsState = {
  skillsStatus: "idle",
  selectedSkillIds: [],
  importStatus: "idle",
  error: null,
};

const skillsSlice = createSlice({
  name: "skills",
  initialState: initialSkillsState,
  reducers: {
    resetForProject() {
      return initialSkillsState;
    },
  },
});

export const skillsActions = skillsSlice.actions;
export const skillsReducer = skillsSlice.reducer;
