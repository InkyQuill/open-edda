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
  async ({
    projectId,
    audit,
    enabled,
    runtimeCommand,
  }: {
    projectId: string;
    audit: ScriptAudit;
    enabled: boolean;
    runtimeCommand?: string;
  }) => {
    const approval = audit.approval;
    const command = approval?.runtimeCommand ?? runtimeCommand?.trim();
    if (!command) {
      throw new Error("Runtime command is not configured for this script");
    }

    return updateScriptApproval(projectId, audit.skillId, audit.skillFileId, {
      enabled,
      runtimeCommand: command,
      timeoutMs: approval?.timeoutMs ?? 5000,
      maxStdoutBytes: approval?.maxStdoutBytes ?? 65536,
      maxStderrBytes: approval?.maxStderrBytes ?? 16384,
      allowNetwork: approval?.allowNetwork ?? false,
      allowProjectFiles: approval?.allowProjectFiles ?? false,
    });
  },
);
