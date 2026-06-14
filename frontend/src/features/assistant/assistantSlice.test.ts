import { describe, expect, it } from "vitest";
import {
  assistantActions,
  assistantReducer,
  initialAssistantState,
} from "./assistantSlice";

describe("assistantSlice", () => {
  it("resets project-scoped assistant state", () => {
    const loaded = assistantReducer(
      {
        ...initialAssistantState,
        activeSessionId: "session-1",
        draftMessage: "hello",
        sessionsStatus: "succeeded",
        messagesStatus: "failed",
        error: "Could not load messages",
      },
      assistantActions.resetForProject(),
    );

    expect(loaded).toEqual(initialAssistantState);
  });
});
