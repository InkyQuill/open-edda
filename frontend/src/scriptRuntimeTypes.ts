export type ScriptRecommendation = "disabled" | "approve_with_limits" | "defer";
export type ScriptRunStatus = "succeeded" | "failed" | "timed_out" | "rejected";

export type ScriptApproval = {
  id: string;
  projectId: string;
  skillId: string;
  skillFileId: string;
  auditId: string;
  enabled: boolean;
  runtimeCommand: string;
  timeoutMs: number;
  maxStdoutBytes: number;
  maxStderrBytes: number;
  allowNetwork: boolean;
  allowProjectFiles: boolean;
  approvedBy: string;
  approvedAt: string;
  updatedAt: string;
};

export type ScriptAudit = {
  id: string;
  projectId: string;
  skillId: string;
  skillFileId: string;
  relativePath: string;
  runtime: string;
  destructiveOperations: boolean;
  filesystemAccess: string;
  networkAccess: boolean;
  externalDependencies: string;
  expectedInputsJson: string;
  expectedOutputsJson: string;
  riskNotes: string;
  recommendation: ScriptRecommendation;
  auditedAt: string;
  approval?: ScriptApproval;
};

export type ScriptOutput = {
  kind: string;
  title?: string;
  markdown?: string;
  data?: unknown;
};

export type ScriptRun = {
  id: string;
  projectId: string;
  sessionId?: string;
  skillId: string;
  skillFileId: string;
  approvalId: string;
  toolCallId: string;
  status: ScriptRunStatus;
  outputKind: string;
  inputJson: string;
  outputJson: string;
  output?: ScriptOutput;
  stdoutText: string;
  stderrText: string;
  exitCode: number;
  errorMessage: string;
  durationMs: number;
  createdAt: string;
};
