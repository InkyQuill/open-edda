import { describe, expect, it } from "vitest";
import { initialSkillsState, skillsActions, skillsReducer } from "./skillsSlice";

describe("skillsSlice", () => {
  it("resets project-scoped skills state", () => {
    const loaded = skillsReducer(
      {
        ...initialSkillsState,
        skillsStatus: "succeeded",
        selectedSkillIds: ["skill-1", "skill-2"],
        importStatus: "failed",
        error: "Could not import skills",
      },
      skillsActions.resetForProject(),
    );

    expect(loaded).toEqual(initialSkillsState);
  });
});
