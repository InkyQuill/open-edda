import { describe, expect, it } from "vitest";
import type { WriterSkill } from "../../skillTypes";
import { loadProjectSkills, loadSessionSkills, saveSessionSkills } from "./skillsThunks";
import { initialSkillsState, skillsActions, skillsReducer } from "./skillsSlice";

function skill(id: string, displayName = id, scriptsDisabled = false): WriterSkill {
  return {
    id,
    projectId: "project-1",
    name: id,
    displayName,
    description: `${displayName} description`,
    sourceType: "upload",
    sourceLabel: `${displayName}.zip`,
    scriptCount: scriptsDisabled ? 1 : 0,
    scriptsDisabled,
    metadataJson: "{}",
    installedAt: "2026-06-14T00:00:00.000Z",
    updatedAt: "2026-06-14T00:00:00.000Z",
  };
}

describe("skillsSlice", () => {
  it("loads project skills", () => {
    const pending = skillsReducer(
      initialSkillsState,
      loadProjectSkills.pending("request-1", { projectId: "project-1" }),
    );

    expect(pending.skillsStatus).toBe("pending");
    expect(pending.projectId).toBe("project-1");
    expect(pending.skillsRequestId).toBe("request-1");

    const loaded = skillsReducer(
      pending,
      loadProjectSkills.fulfilled(
        [skill("continuity", "Continuity"), skill("style", "Style")],
        "request-1",
        { projectId: "project-1" },
      ),
    );

    expect(loaded.skillsStatus).toBe("succeeded");
    expect(loaded.skills.map((item) => item.displayName)).toEqual(["Continuity", "Style"]);
    expect(loaded.skillsRequestId).toBeNull();
  });

  it("clears project-scoped skills state when loading a different project", () => {
    const currentProjectState = {
      ...initialSkillsState,
      projectId: "project-1",
      skills: [skill("continuity", "Continuity")],
      skillsStatus: "succeeded" as const,
      sessionSkillsStatus: "succeeded" as const,
      activeSessionId: "session-1",
      selectedSkillIds: ["continuity"],
      confirmedSelectedSkillIds: ["continuity"],
      saveStatus: "pending" as const,
      saveRequestId: "save-1",
      savingSessionId: "session-1",
    };

    const pending = skillsReducer(
      currentProjectState,
      loadProjectSkills.pending("request-2", { projectId: "project-2" }),
    );

    expect(pending.projectId).toBe("project-2");
    expect(pending.skills).toEqual([]);
    expect(pending.skillsStatus).toBe("pending");
    expect(pending.skillsRequestId).toBe("request-2");
    expect(pending.sessionSkillsStatus).toBe("idle");
    expect(pending.activeSessionId).toBeNull();
    expect(pending.selectedSkillIds).toEqual([]);
    expect(pending.confirmedSelectedSkillIds).toEqual([]);
    expect(pending.saveStatus).toBe("idle");
    expect(pending.saveRequestId).toBeNull();
    expect(pending.savingSessionId).toBeNull();
  });

  it("toggles a selected skill id", () => {
    const selected = skillsReducer(initialSkillsState, skillsActions.toggleSelectedSkillId("continuity"));
    const unselected = skillsReducer(selected, skillsActions.toggleSelectedSkillId("continuity"));

    expect(selected.selectedSkillIds).toEqual(["continuity"]);
    expect(unselected.selectedSkillIds).toEqual([]);
  });

  it("replaces selected session skill ids from server data", () => {
    const pending = skillsReducer(
      initialSkillsState,
      loadSessionSkills.pending("request-1", { projectId: "project-1", sessionId: "session-1" }),
    );

    expect(pending.sessionSkillsStatus).toBe("pending");
    expect(pending.loadingSessionId).toBe("session-1");
    expect(pending.activeSessionId).toBe("session-1");

    const loaded = skillsReducer(
      pending,
      loadSessionSkills.fulfilled(
        [skill("continuity", "Continuity"), skill("revision", "Revision")],
        "request-1",
        { projectId: "project-1", sessionId: "session-1" },
      ),
    );

    expect(loaded.sessionSkillsStatus).toBe("succeeded");
    expect(loaded.selectedSkillIds).toEqual(["continuity", "revision"]);
    expect(loaded.confirmedSelectedSkillIds).toEqual(["continuity", "revision"]);
    expect(loaded.activeSessionId).toBe("session-1");
    expect(loaded.loadingSessionId).toBeNull();
  });

  it("ignores stale selected session skill responses", () => {
    const pending = skillsReducer(
      initialSkillsState,
      loadSessionSkills.pending("request-1", { projectId: "project-1", sessionId: "session-1" }),
    );
    const loadingNewSession = skillsReducer(
      pending,
      loadSessionSkills.pending("request-2", { projectId: "project-1", sessionId: "session-2" }),
    );

    const stale = skillsReducer(
      loadingNewSession,
      loadSessionSkills.fulfilled([skill("stale", "Stale")], "request-1", {
        projectId: "project-1",
        sessionId: "session-1",
      }),
    );

    expect(stale.selectedSkillIds).toEqual([]);
    expect(stale.sessionSkillsStatus).toBe("pending");
    expect(stale.loadingSessionId).toBe("session-2");
    expect(stale.activeSessionId).toBe("session-2");
  });

  it("ignores stale save responses after a different session becomes active", () => {
    const sessionALoaded = skillsReducer(
      skillsReducer(
        initialSkillsState,
        loadSessionSkills.pending("load-a", { projectId: "project-1", sessionId: "session-a" }),
      ),
      loadSessionSkills.fulfilled([skill("continuity", "Continuity")], "load-a", {
        projectId: "project-1",
        sessionId: "session-a",
      }),
    );
    const savePending = skillsReducer(
      {
        ...sessionALoaded,
        selectedSkillIds: ["continuity", "style"],
      },
      saveSessionSkills.pending("save-a", {
        projectId: "project-1",
        sessionId: "session-a",
        skillIds: ["continuity", "style"],
      }),
    );
    const sessionBLoaded = skillsReducer(
      skillsReducer(
        savePending,
        loadSessionSkills.pending("load-b", { projectId: "project-1", sessionId: "session-b" }),
      ),
      loadSessionSkills.fulfilled([skill("revision", "Revision")], "load-b", {
        projectId: "project-1",
        sessionId: "session-b",
      }),
    );

    const staleSave = skillsReducer(
      sessionBLoaded,
      saveSessionSkills.fulfilled([skill("stale", "Stale")], "save-a", {
        projectId: "project-1",
        sessionId: "session-a",
        skillIds: ["continuity", "style"],
      }),
    );

    expect(staleSave.activeSessionId).toBe("session-b");
    expect(staleSave.selectedSkillIds).toEqual(["revision"]);
    expect(staleSave.confirmedSelectedSkillIds).toEqual(["revision"]);
    expect(staleSave.saveStatus).toBe("idle");
  });

  it("rolls back optimistic selection when the current session save fails", () => {
    const loaded = skillsReducer(
      skillsReducer(
        initialSkillsState,
        loadSessionSkills.pending("load-1", { projectId: "project-1", sessionId: "session-1" }),
      ),
      loadSessionSkills.fulfilled([skill("continuity", "Continuity")], "load-1", {
        projectId: "project-1",
        sessionId: "session-1",
      }),
    );
    const optimistic = skillsReducer(
      {
        ...loaded,
        selectedSkillIds: ["continuity", "style"],
      },
      saveSessionSkills.pending("save-1", {
        projectId: "project-1",
        sessionId: "session-1",
        skillIds: ["continuity", "style"],
      }),
    );

    const rejected = skillsReducer(
      optimistic,
      saveSessionSkills.rejected(new Error("Save failed"), "save-1", {
        projectId: "project-1",
        sessionId: "session-1",
        skillIds: ["continuity", "style"],
      }),
    );

    expect(rejected.saveStatus).toBe("failed");
    expect(rejected.selectedSkillIds).toEqual(["continuity"]);
    expect(rejected.confirmedSelectedSkillIds).toEqual(["continuity"]);
    expect(rejected.error).toBe("Save failed");
  });

  it("resets project-scoped skills state", () => {
    const loaded = skillsReducer(
      {
        ...initialSkillsState,
        projectId: "project-1",
        skills: [skill("continuity", "Continuity")],
        skillsStatus: "succeeded",
        sessionSkillsStatus: "succeeded",
        activeSessionId: "session-1",
        selectedSkillIds: ["skill-1", "skill-2"],
        confirmedSelectedSkillIds: ["skill-1"],
        importStatus: "failed",
        error: "Could not import skills",
      },
      skillsActions.resetForProject(),
    );

    expect(loaded).toEqual(initialSkillsState);
  });
});
