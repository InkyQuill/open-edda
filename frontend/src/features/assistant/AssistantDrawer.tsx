import { Plus, Send, WandSparkles } from "lucide-react";
import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";

import type { AppDispatch, RootState } from "../../app/store/store";
import { Button } from "../../shared/ui/button";
import { Textarea } from "../../shared/ui/textarea";
import { ModelStatus } from "../model-settings/ModelStatus";
import { SkillChipsPanel } from "../skills/SkillChipsPanel";
import { assistantActions } from "./assistantSlice";
import {
  loadAssistantSessions,
  sendAssistantMessage,
  startAssistantSession,
} from "./assistantThunks";

type AssistantDrawerProps = {
  projectId: string;
  contentId: string | null;
};

export function AssistantDrawer({ projectId, contentId }: AssistantDrawerProps) {
  const dispatch = useDispatch<AppDispatch>();
  const {
    activeSessionId,
    draftMessage,
    error,
    messagesBySessionId,
    messagesStatus,
    sessions,
    sessionsStatus,
  } = useSelector((state: RootState) => state.assistant);
  const activeSession = sessions.find((session) => session.id === activeSessionId) ?? null;
  const activeMessages = activeSessionId ? (messagesBySessionId[activeSessionId] ?? []) : [];
  const trimmedDraft = draftMessage.trim();
  const isSending = messagesStatus === "pending";
  const isStarting = sessionsStatus === "pending";

  useEffect(() => {
    dispatch(assistantActions.resetForProject());
    void dispatch(loadAssistantSessions({ projectId }));
  }, [dispatch, projectId]);

  function handleNewChat(): void {
    void dispatch(startAssistantSession({ projectId, contentId }));
  }

  function handleSend(): void {
    if (!activeSessionId || !trimmedDraft) return;
    void dispatch(sendAssistantMessage({ projectId, sessionId: activeSessionId, bodyMarkdown: trimmedDraft }));
  }

  return (
    <aside className="flex h-full flex-col gap-5" aria-label="Assistant">
      <header className="flex flex-col gap-1">
        <h2 className="text-base font-semibold text-foreground">Assistant</h2>
        <p className="text-sm text-muted-foreground">
          {activeSession ? activeSession.title : sessionsStatus === "pending" ? "Loading chats..." : "No active chat"}
        </p>
      </header>

      <section className="flex flex-col gap-3" aria-labelledby="assistant-transcript-title">
        <div className="flex items-center justify-between gap-3">
          <h3 id="assistant-transcript-title" className="truncate text-sm font-medium text-foreground">
            {activeSession?.title ?? "Transcript"}
          </h3>
          <Button type="button" variant="outline" size="xs" onClick={handleNewChat} disabled={isStarting}>
            <Plus data-icon="inline-start" aria-hidden="true" />
            New chat
          </Button>
        </div>

        {sessions.length === 0 && sessionsStatus !== "pending" ? (
          <div className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
            Start a new chat to work with the assistant in this project.
          </div>
        ) : null}

        {activeSession && activeMessages.length === 0 ? (
          <div className="rounded-md border border-border bg-background p-3 text-sm text-muted-foreground">
            No messages in this chat yet.
          </div>
        ) : null}

        {activeMessages.length > 0 ? (
          <div className="flex max-h-80 flex-col gap-2 overflow-auto pr-1">
            {activeMessages.map((message) => (
              <article
                key={message.id}
                className="rounded-md border border-border bg-background p-3 text-sm"
              >
                <p className="mb-1 text-xs font-medium uppercase text-muted-foreground">{message.role}</p>
                <p className="whitespace-pre-wrap text-foreground">{message.bodyMarkdown}</p>
              </article>
            ))}
          </div>
        ) : null}

        {error ? (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        ) : null}

        <div className="flex flex-col gap-2">
          <Textarea
            value={draftMessage}
            onChange={(event) => dispatch(assistantActions.setDraftMessage(event.target.value))}
            placeholder="Message the assistant..."
            aria-label="Assistant message"
            disabled={!activeSessionId || isSending}
          />
          <Button
            type="button"
            onClick={handleSend}
            disabled={!activeSessionId || !trimmedDraft || isSending}
          >
            <Send data-icon="inline-start" aria-hidden="true" />
            Send
          </Button>
        </div>
      </section>

      <section className="flex flex-col gap-3" aria-labelledby="quick-actions-title">
        <div className="flex items-center justify-between gap-3">
          <h3 id="quick-actions-title" className="text-sm font-medium text-foreground">
            Quick actions
          </h3>
          <Button type="button" variant="outline" size="xs">
            <WandSparkles />
            Ready
          </Button>
        </div>
        <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
          Selection actions will attach to the editor in the next task.
        </p>
      </section>

      <SkillChipsPanel />
      <ModelStatus />
    </aside>
  );
}
