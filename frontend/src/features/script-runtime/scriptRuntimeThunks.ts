import { createAsyncThunk } from "@reduxjs/toolkit";

import { listProjectScriptRuns, listScriptAudits, updateScriptApproval } from "../../scriptRuntimeApi";
import type { ScriptAudit } from "../../scriptRuntimeTypes";

export const loadScriptAudits = createAsyncThunk(
  "scriptRuntime/loadScriptAudits",
  async ({ projectId, skillId }: { projectId: string; skillId: string }, { signal }) =>
    listScriptAudits(projectId, skillId, signal),
);

export const loadScriptRuns = createAsyncThunk(
  "scriptRuntime/loadScriptRuns",
  async ({ projectId }: { projectId: string }, { signal }) => listProjectScriptRuns(projectId, signal),
);

export const setScriptApproval = createAsyncThunk(
  "scriptRuntime/setScriptApproval",
  async ({ projectId, audit, enabled }: { projectId: string; audit: ScriptAudit; enabled: boolean }) => {
    const approval = audit.approval;
    if (!approval?.runtimeCommand) {
      throw new Error("Runtime command is not configured for this script");
    }

    return updateScriptApproval(projectId, audit.skillId, audit.skillFileId, {
      enabled,
      runtimeCommand: approval.runtimeCommand,
      timeoutMs: approval.timeoutMs,
      maxStdoutBytes: approval.maxStdoutBytes,
      maxStderrBytes: approval.maxStderrBytes,
      allowNetwork: approval.allowNetwork,
      allowProjectFiles: approval.allowProjectFiles,
    });
  },
);
