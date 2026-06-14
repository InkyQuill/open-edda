import { createSlice } from "@reduxjs/toolkit";

export type AsyncStatus = "idle" | "pending" | "succeeded" | "failed";

export type ScriptRuntimeState = {
  auditsStatus: AsyncStatus;
  runsStatus: AsyncStatus;
  selectedAuditId: string | null;
  error: string | null;
};

export const initialScriptRuntimeState: ScriptRuntimeState = {
  auditsStatus: "idle",
  runsStatus: "idle",
  selectedAuditId: null,
  error: null,
};

const scriptRuntimeSlice = createSlice({
  name: "scriptRuntime",
  initialState: initialScriptRuntimeState,
  reducers: {
    resetForProject() {
      return initialScriptRuntimeState;
    },
  },
});

export const scriptRuntimeActions = scriptRuntimeSlice.actions;
export const scriptRuntimeReducer = scriptRuntimeSlice.reducer;
