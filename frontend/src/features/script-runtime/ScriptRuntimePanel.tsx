import { AlertCircle, CheckCircle2, Clock3, Loader2, PlayCircle, RefreshCw, ShieldAlert, ShieldCheck, Slash } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useDispatch, useSelector } from "react-redux";

import type { AppDispatch, RootState } from "../../app/store/store";
import type { ScriptAudit, ScriptRecommendation, ScriptRun, ScriptRunStatus } from "../../scriptRuntimeTypes";
import { cn } from "../../shared/lib/utils";
import { Button } from "../../shared/ui/button";
import { Input } from "../../shared/ui/input";
import { loadProjectSkills } from "../skills/skillsThunks";
import { scriptRuntimeActions } from "./scriptRuntimeSlice";
import { loadScriptAudits, loadScriptRuns, setScriptApproval } from "./scriptRuntimeThunks";

type ScriptRuntimePanelProps = {
  projectId: string;
};

const EMPTY_AUDITS_BY_SKILL_ID: Record<string, ScriptAudit[]> = {};
const EMPTY_RUNS: ScriptRun[] = [];

const recommendationLabels: Record<ScriptRecommendation, string> = {
  disabled: "Disabled",
  approve_with_limits: "Approve with limits",
  defer: "Deferred",
};

const runStatusLabels: Record<ScriptRunStatus, string> = {
  succeeded: "Succeeded",
  failed: "Failed",
  timed_out: "Timed out",
  rejected: "Rejected",
};

function recommendationClassName(recommendation: ScriptRecommendation): string {
  if (recommendation === "approve_with_limits") {
    return "border-amber-300 bg-amber-50 text-amber-800 dark:border-amber-400/40 dark:bg-amber-400/10 dark:text-amber-200";
  }
  if (recommendation === "defer") {
    return "border-blue-300 bg-blue-50 text-blue-800 dark:border-blue-400/40 dark:bg-blue-400/10 dark:text-blue-200";
  }
  return "border-muted bg-muted text-muted-foreground";
}

function runStatusClassName(status: ScriptRunStatus): string {
  if (status === "succeeded") return "text-emerald-700 dark:text-emerald-300";
  if (status === "rejected") return "text-muted-foreground";
  return "text-destructive";
}

function auditCanBeEnabled(audit: ScriptAudit): boolean {
  return audit.recommendation === "approve_with_limits";
}

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function findSelectedAudit(audits: ScriptAudit[], selectedAuditId: string | null): ScriptAudit | null {
  if (selectedAuditId) {
    return audits.find((audit) => audit.id === selectedAuditId) ?? audits[0] ?? null;
  }
  return audits[0] ?? null;
}

export function ScriptRuntimePanel({ projectId }: ScriptRuntimePanelProps) {
  const dispatch = useDispatch<AppDispatch>();
  const skillsState = useSelector((state: RootState) => state.skills);
  const runtime = useSelector((state: RootState) => state.scriptRuntime);
  const currentSkills = skillsState.projectId === projectId ? skillsState.skills : [];
  const runtimeIsCurrentProject = runtime.projectId === projectId;
  const visibleAuditsBySkillId = runtimeIsCurrentProject ? runtime.auditsBySkillId : EMPTY_AUDITS_BY_SKILL_ID;
  const visibleRuns = runtimeIsCurrentProject ? runtime.runs : EMPTY_RUNS;
  const visibleRuntimeError = runtimeIsCurrentProject ? runtime.error : null;
  const scriptedSkills = useMemo(
    () => currentSkills.filter((skill) => skill.scriptCount > 0 || skill.scriptsDisabled),
    [currentSkills],
  );
  const [selectedSkillId, setSelectedSkillId] = useState<string | null>(null);
  const [approvalCommandByAuditId, setApprovalCommandByAuditId] = useState<Record<string, string>>({});
  const selectedSkill = scriptedSkills.find((skill) => skill.id === selectedSkillId) ?? scriptedSkills[0] ?? null;
  const audits = selectedSkill ? (visibleAuditsBySkillId[selectedSkill.id] ?? []) : [];
  const selectedAudit = findSelectedAudit(audits, runtime.selectedAuditId);
  const selectedRuns = selectedAudit
    ? visibleRuns.filter((run) => run.skillId === selectedAudit.skillId && run.skillFileId === selectedAudit.skillFileId)
    : visibleRuns;

  useEffect(() => {
    if (skillsState.skillsStatus === "idle" || skillsState.projectId !== projectId) {
      void dispatch(loadProjectSkills({ projectId }));
    }
  }, [dispatch, projectId, skillsState.projectId, skillsState.skillsStatus]);

  useEffect(() => {
    if (scriptedSkills.length === 0) {
      if (selectedSkillId) {
        setSelectedSkillId(null);
      }
      return;
    }
    if (!selectedSkillId || !scriptedSkills.some((skill) => skill.id === selectedSkillId)) {
      setSelectedSkillId(scriptedSkills[0].id);
    }
  }, [scriptedSkills, selectedSkillId]);

  useEffect(() => {
    if (runtime.runsStatus === "idle" || runtime.projectId !== projectId) {
      void dispatch(loadScriptRuns({ projectId }));
    }
  }, [dispatch, projectId, runtime.projectId, runtime.runsStatus]);

  useEffect(() => {
    if (!selectedSkill) return;
    if (visibleAuditsBySkillId[selectedSkill.id] || (runtimeIsCurrentProject && runtime.loadingAuditSkillId === selectedSkill.id)) return;
    void dispatch(loadScriptAudits({ projectId, skillId: selectedSkill.id }));
  }, [dispatch, projectId, runtime.loadingAuditSkillId, runtimeIsCurrentProject, selectedSkill, visibleAuditsBySkillId]);

  useEffect(() => {
    if (selectedAudit && runtime.selectedAuditId !== selectedAudit.id) {
      dispatch(scriptRuntimeActions.selectAudit(selectedAudit.id));
    }
  }, [dispatch, runtime.selectedAuditId, selectedAudit]);

  function handleSkillSelect(skillId: string): void {
    setSelectedSkillId(skillId);
    dispatch(scriptRuntimeActions.selectAudit(null));
  }

  function handleRefresh(): void {
    if (selectedSkill) {
      void dispatch(loadScriptAudits({ projectId, skillId: selectedSkill.id }));
    }
    void dispatch(loadScriptRuns({ projectId }));
  }

  function handleApproval(audit: ScriptAudit, enabled: boolean): void {
    void dispatch(setScriptApproval({ projectId, audit, enabled, runtimeCommand: approvalCommandByAuditId[audit.id] }));
  }

  const loadingAudits = runtimeIsCurrentProject && runtime.auditsStatus === "pending" && runtime.loadingAuditSkillId === selectedSkill?.id;
  const approvalUpdating = runtime.approvalStatus === "pending";
  const selectedAuditCommandDraft = selectedAudit ? (approvalCommandByAuditId[selectedAudit.id] ?? "") : "";
  const selectedAuditHasRuntimeCommand = Boolean(selectedAudit?.approval?.runtimeCommand);
  const selectedAuditCanSubmitRuntimeCommand = selectedAuditHasRuntimeCommand || selectedAuditCommandDraft.trim().length > 0;

  return (
    <section className="flex min-h-0 flex-col gap-4" aria-labelledby="script-runtime-title">
      <header className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 id="script-runtime-title" className="text-base font-semibold text-foreground">
            Script runtime
          </h2>
          <p className="text-sm text-muted-foreground">
            {scriptedSkills.length ? `${scriptedSkills.length} scripted skill${scriptedSkills.length === 1 ? "" : "s"}` : "No scripted skills"}
          </p>
        </div>
        <Button type="button" variant="outline" size="icon-sm" aria-label="Refresh script runtime" onClick={handleRefresh}>
          <RefreshCw />
        </Button>
      </header>

      {visibleRuntimeError ? (
        <p role="alert" className="flex items-start gap-2 rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
          <AlertCircle className="mt-0.5 size-4" aria-hidden="true" />
          {visibleRuntimeError}
        </p>
      ) : null}

      {skillsState.skillsStatus === "pending" ? (
        <p className="flex items-center gap-2 rounded-md border border-border p-3 text-sm text-muted-foreground">
          <Loader2 className="size-4 animate-spin" aria-hidden="true" />
          Loading skills...
        </p>
      ) : null}

      {skillsState.skillsStatus !== "pending" && scriptedSkills.length === 0 ? (
        <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
          No installed skills expose runtime scripts.
        </p>
      ) : null}

      {scriptedSkills.length ? (
        <div className="flex flex-col gap-2">
          {scriptedSkills.map((skill) => (
            <Button
              key={skill.id}
              type="button"
              variant={skill.id === selectedSkill?.id ? "secondary" : "outline"}
              className="h-auto justify-start px-3 py-2 text-left"
              aria-pressed={skill.id === selectedSkill?.id}
              onClick={() => handleSkillSelect(skill.id)}
            >
              <PlayCircle data-icon="inline-start" aria-hidden="true" />
              <span className="min-w-0 flex-1">
                <span className="block truncate font-medium">{skill.displayName}</span>
                <span className="block truncate text-xs text-muted-foreground">
                  {skill.scriptsDisabled ? "Scripts disabled by policy" : `${skill.scriptCount} script${skill.scriptCount === 1 ? "" : "s"}`}
                </span>
              </span>
            </Button>
          ))}
        </div>
      ) : null}

      {selectedSkill ? (
        <div className="flex min-h-0 flex-col gap-3">
          <div className="flex items-center justify-between gap-3">
            <h3 className="text-sm font-medium text-foreground">Audits</h3>
            {loadingAudits ? <Loader2 className="size-4 animate-spin text-muted-foreground" aria-hidden="true" /> : null}
          </div>

          {!loadingAudits && audits.length === 0 ? (
            <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
              No script audits are available for this skill.
            </p>
          ) : null}

          <div className="flex max-h-64 flex-col gap-2 overflow-auto pr-1">
            {audits.map((audit) => {
              const enabled = audit.approval?.enabled ?? false;
              return (
                <Button
                  key={audit.id}
                  type="button"
                  variant={audit.id === selectedAudit?.id ? "secondary" : "outline"}
                  className="h-auto justify-start px-3 py-2 text-left"
                  aria-pressed={audit.id === selectedAudit?.id}
                  onClick={() => dispatch(scriptRuntimeActions.selectAudit(audit.id))}
                >
                  {enabled ? <ShieldCheck data-icon="inline-start" aria-hidden="true" /> : <ShieldAlert data-icon="inline-start" aria-hidden="true" />}
                  <span className="min-w-0 flex-1">
                    <span className="block truncate font-medium">{audit.relativePath}</span>
                    <span className="block truncate text-xs text-muted-foreground">
                      {enabled ? "Approved" : "Not approved"} · {audit.runtime}
                    </span>
                  </span>
                </Button>
              );
            })}
          </div>
        </div>
      ) : null}

      {selectedAudit ? (
        <div className="grid gap-3 rounded-md border border-border bg-muted/30 p-3">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h3 className="truncate text-sm font-medium text-foreground">{selectedAudit.relativePath}</h3>
              <p className="text-xs text-muted-foreground">Audited {formatDate(selectedAudit.auditedAt)}</p>
            </div>
            <span className={cn("shrink-0 rounded-md border px-2 py-1 text-xs font-medium", recommendationClassName(selectedAudit.recommendation))}>
              {recommendationLabels[selectedAudit.recommendation]}
            </span>
          </div>

          <div className="grid gap-2 text-xs text-muted-foreground">
            <div className="flex justify-between gap-3">
              <span>Filesystem</span>
              <span className="truncate text-foreground">{selectedAudit.filesystemAccess || "none"}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span>Network</span>
              <span className="text-foreground">{selectedAudit.networkAccess ? "Requested" : "Not requested"}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span>Destructive</span>
              <span className="text-foreground">{selectedAudit.destructiveOperations ? "Yes" : "No"}</span>
            </div>
          </div>

          {selectedAudit.riskNotes ? <p className="text-sm text-muted-foreground">{selectedAudit.riskNotes}</p> : null}

          {selectedAudit.recommendation !== "approve_with_limits" ? (
            <p className="flex items-start gap-2 rounded-md border border-border bg-background p-2 text-xs text-muted-foreground">
              <Slash className="mt-0.5 size-3.5" aria-hidden="true" />
              {selectedAudit.recommendation === "disabled"
                ? "This script is disabled and cannot be enabled from this panel."
                : "This script is deferred for manual review before runtime approval."}
            </p>
          ) : null}

          {selectedAudit.recommendation === "approve_with_limits" && !selectedAuditHasRuntimeCommand ? (
            <div className="grid gap-1.5">
              <label className="text-xs font-medium text-foreground" htmlFor={`runtime-command-${selectedAudit.id}`}>
                Runtime command
              </label>
              <Input
                id={`runtime-command-${selectedAudit.id}`}
                value={selectedAuditCommandDraft}
                placeholder={selectedAudit.runtime}
                autoComplete="off"
                onChange={(event) =>
                  setApprovalCommandByAuditId((commands) => ({
                    ...commands,
                    [selectedAudit.id]: event.target.value,
                  }))
                }
              />
            </div>
          ) : null}

          <div className="flex gap-2">
            <Button
              type="button"
              size="sm"
              disabled={
                !auditCanBeEnabled(selectedAudit) ||
                !selectedAuditCanSubmitRuntimeCommand ||
                approvalUpdating ||
                selectedAudit.approval?.enabled === true
              }
              onClick={() => handleApproval(selectedAudit, true)}
            >
              <CheckCircle2 data-icon="inline-start" aria-hidden="true" />
              Enable
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={!selectedAudit.approval || selectedAudit.approval.enabled !== true || approvalUpdating}
              onClick={() => handleApproval(selectedAudit, false)}
            >
              Disable
            </Button>
          </div>
        </div>
      ) : null}

      <div className="flex min-h-0 flex-col gap-3">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-sm font-medium text-foreground">Run history</h3>
          {runtime.runsStatus === "pending" ? <Loader2 className="size-4 animate-spin text-muted-foreground" aria-hidden="true" /> : null}
        </div>

        {runtime.runsStatus !== "pending" && selectedRuns.length === 0 ? (
          <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
            No script runs recorded for this selection.
          </p>
        ) : null}

        <div className="flex max-h-56 flex-col gap-2 overflow-auto pr-1">
          {selectedRuns.slice(0, 8).map((run) => (
            <div key={run.id} className="grid gap-1 rounded-md border border-border bg-background p-3 text-sm">
              <div className="flex items-center justify-between gap-3">
                <span className={cn("font-medium", runStatusClassName(run.status))}>{runStatusLabels[run.status]}</span>
                <span className="flex items-center gap-1 text-xs text-muted-foreground">
                  <Clock3 className="size-3" aria-hidden="true" />
                  {formatDate(run.createdAt)}
                </span>
              </div>
              <p className="truncate text-xs text-muted-foreground">
                {run.outputKind || "output"} · {run.durationMs}ms · exit {run.exitCode}
              </p>
              {run.errorMessage ? <p className="text-xs text-destructive">{run.errorMessage}</p> : null}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
