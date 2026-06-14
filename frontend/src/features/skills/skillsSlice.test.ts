import { describe, expect, it } from "vitest";
import type { WriterSkill } from "../../skillTypes";
import { loadProjectSkills, loadSessionSkills } from "./skillsThunks";
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
  });

  it("resets project-scoped skills state", () => {
    const loaded = skillsReducer(
      {
        ...initialSkillsState,
        projectId: "project-1",
        skills: [skill("continuity", "Continuity")],
        skillsStatus: "succeeded",
        sessionSkillsStatus: "succeeded",
        selectedSkillIds: ["skill-1", "skill-2"],
        importStatus: "failed",
        error: "Could not import skills",
      },
      skillsActions.resetForProject(),
    );

    expect(loaded).toEqual(initialSkillsState);
  });
});
