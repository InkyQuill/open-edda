import { describe, expect, it } from "vitest";
import {
  initialScriptRuntimeState,
  scriptRuntimeActions,
  scriptRuntimeReducer,
} from "./scriptRuntimeSlice";

describe("scriptRuntimeSlice", () => {
  it("resets project-scoped script runtime state", () => {
    const loaded = scriptRuntimeReducer(
      {
        ...initialScriptRuntimeState,
        auditsStatus: "succeeded",
        runsStatus: "pending",
        selectedAuditId: "audit-1",
        error: "Could not load runs",
      },
      scriptRuntimeActions.resetForProject(),
    );

    expect(loaded).toEqual(initialScriptRuntimeState);
  });
});
