import { describe, expect, it } from "vitest";
import { initialSettingsState, settingsActions, settingsReducer } from "./settingsSlice";

describe("settingsSlice", () => {
  it("selects a project for project-scoped administration panels", () => {
    const state = settingsReducer(initialSettingsState, settingsActions.setSelectedProjectId("project-1"));

    expect(state.selectedProjectId).toBe("project-1");
  });

  it("clears selected project when set to null", () => {
    const state = settingsReducer(
      { ...initialSettingsState, selectedProjectId: "project-1" },
      settingsActions.setSelectedProjectId(null),
    );

    expect(state.selectedProjectId).toBeNull();
  });
});
