import { createSlice } from "@reduxjs/toolkit";
import type { ScriptAudit, ScriptRun } from "../../scriptRuntimeTypes";
import { loadScriptAudits, loadScriptRuns, setScriptApproval } from "./scriptRuntimeThunks";

export type AsyncStatus = "idle" | "pending" | "succeeded" | "failed";

export type ScriptRuntimeState = {
  projectId: string | null;
  auditsBySkillId: Record<string, ScriptAudit[]>;
  runs: ScriptRun[];
  auditsStatus: AsyncStatus;
  runsStatus: AsyncStatus;
  approvalStatus: AsyncStatus;
  auditsRequestId: string | null;
  runsRequestId: string | null;
  approvalRequestId: string | null;
  loadingAuditSkillId: string | null;
  updatingAuditId: string | null;
  selectedAuditId: string | null;
  error: string | null;
};

export const initialScriptRuntimeState: ScriptRuntimeState = {
  projectId: null,
  auditsBySkillId: {},
  runs: [],
  auditsStatus: "idle",
  runsStatus: "idle",
  approvalStatus: "idle",
  auditsRequestId: null,
  runsRequestId: null,
  approvalRequestId: null,
  loadingAuditSkillId: null,
  updatingAuditId: null,
  selectedAuditId: null,
  error: null,
};

function clearProjectData(state: ScriptRuntimeState, projectId: string): void {
  if (state.projectId === projectId) return;
  state.projectId = projectId;
  state.auditsBySkillId = {};
  state.runs = [];
  state.auditsStatus = "idle";
  state.runsStatus = "idle";
  state.approvalStatus = "idle";
  state.auditsRequestId = null;
  state.runsRequestId = null;
  state.approvalRequestId = null;
  state.loadingAuditSkillId = null;
  state.updatingAuditId = null;
  state.selectedAuditId = null;
  state.error = null;
}

const scriptRuntimeSlice = createSlice({
  name: "scriptRuntime",
  initialState: initialScriptRuntimeState,
  reducers: {
    selectAudit(state, action: { payload: string | null }) {
      state.selectedAuditId = action.payload;
    },
    resetForProject() {
      return initialScriptRuntimeState;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loadScriptAudits.pending, (state, action) => {
        clearProjectData(state, action.meta.arg.projectId);
        state.auditsStatus = "pending";
        state.auditsRequestId = action.meta.requestId;
        state.loadingAuditSkillId = action.meta.arg.skillId;
        state.error = null;
      })
      .addCase(loadScriptAudits.fulfilled, (state, action) => {
        if (
          state.projectId !== action.meta.arg.projectId ||
          state.auditsRequestId !== action.meta.requestId ||
          state.loadingAuditSkillId !== action.meta.arg.skillId
        ) {
          return;
        }
        state.auditsBySkillId[action.meta.arg.skillId] = action.payload;
        state.auditsStatus = "succeeded";
        state.auditsRequestId = null;
        state.loadingAuditSkillId = null;
        state.error = null;
      })
      .addCase(loadScriptAudits.rejected, (state, action) => {
        if (
          state.projectId !== action.meta.arg.projectId ||
          state.auditsRequestId !== action.meta.requestId ||
          state.loadingAuditSkillId !== action.meta.arg.skillId
        ) {
          return;
        }
        state.auditsStatus = "failed";
        state.auditsRequestId = null;
        state.loadingAuditSkillId = null;
        state.error = action.error.message ?? "Could not load script audits";
      })
      .addCase(loadScriptRuns.pending, (state, action) => {
        clearProjectData(state, action.meta.arg.projectId);
        state.runsStatus = "pending";
        state.runsRequestId = action.meta.requestId;
        state.error = null;
      })
      .addCase(loadScriptRuns.fulfilled, (state, action) => {
        if (state.projectId !== action.meta.arg.projectId || state.runsRequestId !== action.meta.requestId) {
          return;
        }
        state.runs = action.payload;
        state.runsStatus = "succeeded";
        state.runsRequestId = null;
        state.error = null;
      })
      .addCase(loadScriptRuns.rejected, (state, action) => {
        if (state.projectId !== action.meta.arg.projectId || state.runsRequestId !== action.meta.requestId) {
          return;
        }
        state.runsStatus = "failed";
        state.runsRequestId = null;
        state.error = action.error.message ?? "Could not load script runs";
      })
      .addCase(setScriptApproval.pending, (state, action) => {
        clearProjectData(state, action.meta.arg.projectId);
        state.approvalStatus = "pending";
        state.approvalRequestId = action.meta.requestId;
        state.updatingAuditId = action.meta.arg.audit.id;
        state.error = null;
      })
      .addCase(setScriptApproval.fulfilled, (state, action) => {
        if (
          state.projectId !== action.meta.arg.projectId ||
          state.approvalRequestId !== action.meta.requestId ||
          state.updatingAuditId !== action.meta.arg.audit.id
        ) {
          return;
        }
        const audits = state.auditsBySkillId[action.payload.skillId] ?? [];
        const audit = audits.find((item) => item.id === action.payload.auditId);
        if (audit) {
          audit.approval = action.payload;
        }
        state.approvalStatus = "succeeded";
        state.approvalRequestId = null;
        state.updatingAuditId = null;
        state.error = null;
      })
      .addCase(setScriptApproval.rejected, (state, action) => {
        if (
          state.projectId !== action.meta.arg.projectId ||
          state.approvalRequestId !== action.meta.requestId ||
          state.updatingAuditId !== action.meta.arg.audit.id
        ) {
          return;
        }
        state.approvalStatus = "failed";
        state.approvalRequestId = null;
        state.updatingAuditId = null;
        state.error = action.error.message ?? "Could not update script approval";
      });
  },
});

export const scriptRuntimeActions = scriptRuntimeSlice.actions;
export const scriptRuntimeReducer = scriptRuntimeSlice.reducer;
