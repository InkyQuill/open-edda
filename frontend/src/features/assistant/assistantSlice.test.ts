import { describe, expect, it } from "vitest";
import {
  assistantActions,
  assistantReducer,
  initialAssistantState,
} from "./assistantSlice";
import type { AgentMessage, AgentSession } from "../../agentTypes";
import { loadAssistantSessions, sendAssistantMessage } from "./assistantThunks";

function session(id: string, updatedAt = "2026-06-14T00:00:00Z"): AgentSession {
  return {
    id,
    projectId: "project-1",
    title: `Session ${id}`,
    actionKind: "chat",
    modelVariantId: "",
    applyMode: "preview",
    createdAt: updatedAt,
    updatedAt,
  };
}

function message(id: string, sessionId: string, role: AgentMessage["role"], bodyMarkdown: string): AgentMessage {
  return {
    id,
    sessionId,
    role,
    bodyMarkdown,
    metadataJson: "{}",
    createdAt: "2026-06-14T00:00:00Z",
  };
}

describe("assistantSlice", () => {
  it("stores the draft assistant message", () => {
    const loaded = assistantReducer(
      initialAssistantState,
      assistantActions.setDraftMessage("Can you continue this scene?"),
    );

    expect(loaded.draftMessage).toBe("Can you continue this scene?");
  });

  it("stores the active assistant session id", () => {
    const loaded = assistantReducer(
      initialAssistantState,
      assistantActions.setActiveSessionId("session-1"),
    );

    expect(loaded.activeSessionId).toBe("session-1");
  });

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

  it("tracks assistant session loading and selects the newest loaded session when current one is absent", () => {
    const pending = assistantReducer(
      initialAssistantState,
      loadAssistantSessions.pending("request-1", { projectId: "project-1" }),
    );

    expect(pending.sessionsStatus).toBe("pending");

    const loaded = assistantReducer(
      {
        ...pending,
        activeSessionId: "missing-session",
      },
      loadAssistantSessions.fulfilled(
        [session("session-2", "2026-06-14T02:00:00Z"), session("session-1", "2026-06-14T01:00:00Z")],
        "request-1",
        { projectId: "project-1" },
      ),
    );

    expect(loaded.sessionsStatus).toBe("succeeded");
    expect(loaded.sessions.map((item) => item.id)).toEqual(["session-2", "session-1"]);
    expect(loaded.activeSessionId).toBe("session-2");
  });

  it("clears the draft message after a message is sent", () => {
    const loaded = assistantReducer(
      {
        ...initialAssistantState,
        draftMessage: "Can you continue this scene?",
      },
      sendAssistantMessage.fulfilled(
        {
          userMessage: message("message-1", "session-1", "user", "Can you continue this scene?"),
          assistantMessage: message("message-2", "session-1", "assistant", "Yes."),
        },
        "request-1",
        {
          projectId: "project-1",
          sessionId: "session-1",
          bodyMarkdown: "Can you continue this scene?",
        },
      ),
    );

    expect(loaded.draftMessage).toBe("");
    expect(loaded.messagesBySessionId["session-1"].map((item) => item.id)).toEqual(["message-1", "message-2"]);
  });

  it("stores readable assistant load errors", () => {
    const loaded = assistantReducer(
      initialAssistantState,
      loadAssistantSessions.rejected(new Error("Could not load sessions"), "request-1", { projectId: "project-1" }),
    );

    expect(loaded.sessionsStatus).toBe("failed");
    expect(loaded.error).toBe("Could not load sessions");
  });
});
