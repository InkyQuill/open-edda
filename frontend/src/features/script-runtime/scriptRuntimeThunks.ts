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
  async ({ projectId, audit, enabled }: { projectId: string; audit: ScriptAudit; enabled: boolean }) =>
    updateScriptApproval(projectId, audit.skillId, audit.skillFileId, {
      enabled,
      runtimeCommand: audit.approval?.runtimeCommand ?? audit.runtime,
      timeoutMs: audit.approval?.timeoutMs ?? 5000,
      maxStdoutBytes: audit.approval?.maxStdoutBytes ?? 65536,
      maxStderrBytes: audit.approval?.maxStderrBytes ?? 16384,
      allowNetwork: audit.approval?.allowNetwork ?? false,
      allowProjectFiles: audit.approval?.allowProjectFiles ?? false,
    }),
);
