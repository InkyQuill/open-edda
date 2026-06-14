import { describe, expect, it, vi } from "vitest";
import type { ScriptApproval, ScriptAudit, ScriptRun } from "../../scriptRuntimeTypes";
import { updateScriptApproval } from "../../scriptRuntimeApi";
import { loadScriptAudits, loadScriptRuns, setScriptApproval } from "./scriptRuntimeThunks";
import {
  initialScriptRuntimeState,
  scriptRuntimeActions,
  scriptRuntimeReducer,
} from "./scriptRuntimeSlice";

vi.mock("../../scriptRuntimeApi", () => ({
  listProjectScriptRuns: vi.fn(),
  listScriptAudits: vi.fn(),
  updateScriptApproval: vi.fn(),
}));

function approval(overrides: Partial<ScriptApproval> = {}): ScriptApproval {
  return {
    id: "approval-1",
    projectId: "project-1",
    skillId: "skill-1",
    skillFileId: "file-1",
    auditId: "audit-1",
    enabled: true,
    runtimeCommand: "node",
    timeoutMs: 5000,
    maxStdoutBytes: 65536,
    maxStderrBytes: 16384,
    allowNetwork: false,
    allowProjectFiles: false,
    approvedBy: "Pavel",
    approvedAt: "2026-06-14T00:00:00.000Z",
    updatedAt: "2026-06-14T00:00:00.000Z",
    ...overrides,
  };
}

function audit(overrides: Partial<ScriptAudit> = {}): ScriptAudit {
  return {
    id: "audit-1",
    projectId: "project-1",
    skillId: "skill-1",
    skillFileId: "file-1",
    relativePath: "scripts/check.ts",
    runtime: "bun",
    destructiveOperations: false,
    filesystemAccess: "none",
    networkAccess: false,
    externalDependencies: "none",
    expectedInputsJson: "{}",
    expectedOutputsJson: "{}",
    riskNotes: "Reads prompt context only.",
    recommendation: "approve_with_limits",
    auditedAt: "2026-06-14T00:00:00.000Z",
    ...overrides,
  };
}

function run(overrides: Partial<ScriptRun> = {}): ScriptRun {
  return {
    id: "run-1",
    projectId: "project-1",
    sessionId: "session-1",
    skillId: "skill-1",
    skillFileId: "file-1",
    approvalId: "approval-1",
    toolCallId: "tool-call-1",
    status: "succeeded",
    outputKind: "markdown",
    inputJson: "{}",
    outputJson: "{}",
    stdoutText: "ok",
    stderrText: "",
    exitCode: 0,
    errorMessage: "",
    durationMs: 42,
    createdAt: "2026-06-14T00:00:00.000Z",
    ...overrides,
  };
}

describe("scriptRuntimeSlice", () => {
  it("stores script audits by skill id", () => {
    const pending = scriptRuntimeReducer(
      initialScriptRuntimeState,
      loadScriptAudits.pending("request-1", { projectId: "project-1", skillId: "skill-1" }),
    );

    expect(pending.projectId).toBe("project-1");
    expect(pending.auditsStatus).toBe("pending");
    expect(pending.auditsRequestId).toBe("request-1");
    expect(pending.loadingAuditSkillId).toBe("skill-1");

    const loaded = scriptRuntimeReducer(
      pending,
      loadScriptAudits.fulfilled([audit()], "request-1", { projectId: "project-1", skillId: "skill-1" }),
    );

    expect(loaded.auditsStatus).toBe("succeeded");
    expect(loaded.auditsBySkillId).toEqual({ "skill-1": [audit()] });
    expect(loaded.auditsRequestId).toBeNull();
    expect(loaded.loadingAuditSkillId).toBeNull();
  });

  it("stores project script runs", () => {
    const pending = scriptRuntimeReducer(
      initialScriptRuntimeState,
      loadScriptRuns.pending("request-1", { projectId: "project-1" }),
    );

    expect(pending.projectId).toBe("project-1");
    expect(pending.runsStatus).toBe("pending");
    expect(pending.runsRequestId).toBe("request-1");

    const loaded = scriptRuntimeReducer(
      pending,
      loadScriptRuns.fulfilled([run()], "request-1", { projectId: "project-1" }),
    );

    expect(loaded.runsStatus).toBe("succeeded");
    expect(loaded.runs).toEqual([run()]);
    expect(loaded.runsRequestId).toBeNull();
  });

  it("selects an audit", () => {
    const selected = scriptRuntimeReducer(initialScriptRuntimeState, scriptRuntimeActions.selectAudit("audit-1"));

    expect(selected.selectedAuditId).toBe("audit-1");
  });

  it("updates approval status on the matching audit", () => {
    const updatedApproval = approval({ enabled: true });
    const enabledAudit = audit({ approval: updatedApproval });
    const state = {
      ...initialScriptRuntimeState,
      projectId: "project-1",
      auditsBySkillId: {
        "skill-1": [audit()],
        "skill-2": [audit({ id: "audit-2", skillId: "skill-2", skillFileId: "file-2" })],
      },
    };

    const pending = scriptRuntimeReducer(
      state,
      setScriptApproval.pending("request-1", { projectId: "project-1", audit: audit(), enabled: true }),
    );

    expect(pending.approvalStatus).toBe("pending");
    expect(pending.approvalRequestId).toBe("request-1");
    expect(pending.updatingAuditId).toBe("audit-1");

    const updated = scriptRuntimeReducer(
      pending,
      setScriptApproval.fulfilled(updatedApproval, "request-1", {
        projectId: "project-1",
        audit: audit(),
        enabled: true,
      }),
    );

    expect(updated.approvalStatus).toBe("succeeded");
    expect(updated.auditsBySkillId["skill-1"]).toEqual([enabledAudit]);
    expect(updated.auditsBySkillId["skill-2"]?.[0]?.approval).toBeUndefined();
    expect(updated.approvalRequestId).toBeNull();
    expect(updated.updatingAuditId).toBeNull();
  });

  it("ignores stale script audit responses", () => {
    const pending = scriptRuntimeReducer(
      initialScriptRuntimeState,
      loadScriptAudits.pending("request-2", { projectId: "project-1", skillId: "skill-1" }),
    );

    const staleSameSkill = scriptRuntimeReducer(
      pending,
      loadScriptAudits.fulfilled([audit({ id: "stale-audit" })], "request-1", {
        projectId: "project-1",
        skillId: "skill-1",
      }),
    );

    expect(staleSameSkill).toEqual(pending);

    const staleOtherProject = scriptRuntimeReducer(
      pending,
      loadScriptAudits.fulfilled([audit({ id: "other-project-audit", projectId: "project-2" })], "request-2", {
        projectId: "project-2",
        skillId: "skill-1",
      }),
    );

    expect(staleOtherProject).toEqual(pending);
  });

  it("ignores stale script run responses", () => {
    const pending = scriptRuntimeReducer(
      initialScriptRuntimeState,
      loadScriptRuns.pending("request-2", { projectId: "project-1" }),
    );

    const staleSameProject = scriptRuntimeReducer(
      pending,
      loadScriptRuns.fulfilled([run({ id: "stale-run" })], "request-1", { projectId: "project-1" }),
    );

    expect(staleSameProject).toEqual(pending);

    const staleOtherProject = scriptRuntimeReducer(
      pending,
      loadScriptRuns.fulfilled([run({ id: "other-project-run", projectId: "project-2" })], "request-2", {
        projectId: "project-2",
      }),
    );

    expect(staleOtherProject).toEqual(pending);
  });

  it("ignores stale approval update responses", () => {
    const state = {
      ...initialScriptRuntimeState,
      projectId: "project-1",
      auditsBySkillId: { "skill-1": [audit({ approval: approval({ enabled: false }) })] },
    };
    const pending = scriptRuntimeReducer(
      state,
      setScriptApproval.pending("request-2", {
        projectId: "project-1",
        audit: audit({ approval: approval({ enabled: false }) }),
        enabled: true,
      }),
    );

    const stale = scriptRuntimeReducer(
      pending,
      setScriptApproval.fulfilled(approval({ enabled: true }), "request-1", {
        projectId: "project-1",
        audit: audit({ approval: approval({ enabled: false }) }),
        enabled: true,
      }),
    );

    expect(stale).toEqual(pending);
  });

  it("clears pending approval markers when loading runtime data for a different project", () => {
    const approvalPending = scriptRuntimeReducer(
      {
        ...initialScriptRuntimeState,
        projectId: "project-1",
        auditsBySkillId: { "skill-1": [audit()] },
        runs: [run()],
      },
      setScriptApproval.pending("approval-request", {
        projectId: "project-1",
        audit: audit(),
        enabled: true,
      }),
    );

    const switchedProject = scriptRuntimeReducer(
      approvalPending,
      loadScriptRuns.pending("runs-request", { projectId: "project-2" }),
    );

    expect(switchedProject.projectId).toBe("project-2");
    expect(switchedProject.auditsBySkillId).toEqual({});
    expect(switchedProject.runs).toEqual([]);
    expect(switchedProject.runsStatus).toBe("pending");
    expect(switchedProject.runsRequestId).toBe("runs-request");
    expect(switchedProject.approvalStatus).toBe("idle");
    expect(switchedProject.approvalRequestId).toBeNull();
    expect(switchedProject.updatingAuditId).toBeNull();
  });

  it("rejects approval updates when the audit has no configured runtime command", async () => {
    const dispatch = vi.fn();
    const getState = vi.fn();

    const action = await setScriptApproval({
      projectId: "project-1",
      audit: audit(),
      enabled: true,
    })(dispatch, getState, undefined);

    expect(setScriptApproval.rejected.match(action)).toBe(true);
    if (!setScriptApproval.rejected.match(action)) {
      throw new Error("Expected setScriptApproval to reject");
    }
    expect(action.error.message).toBe("Runtime command is not configured for this script");
    expect(updateScriptApproval).not.toHaveBeenCalled();
  });

  it("resets project-scoped script runtime state", () => {
    const loaded = scriptRuntimeReducer(
      {
        ...initialScriptRuntimeState,
        projectId: "project-1",
        auditsBySkillId: { "skill-1": [audit()] },
        runs: [run()],
        auditsStatus: "succeeded",
        runsStatus: "pending",
        approvalStatus: "failed",
        auditsRequestId: "audits-request",
        runsRequestId: "runs-request",
        approvalRequestId: "approval-request",
        loadingAuditSkillId: "skill-1",
        updatingAuditId: "audit-1",
        selectedAuditId: "audit-1",
        error: "Could not load runs",
      },
      scriptRuntimeActions.resetForProject(),
    );

    expect(loaded).toEqual(initialScriptRuntimeState);
  });
});
