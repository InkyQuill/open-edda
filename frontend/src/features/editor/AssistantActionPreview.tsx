import { useSelector } from "react-redux";

import { Button } from "../../shared/ui/button";
import type { AssistantActionState } from "../assistant-actions/assistantActionsSlice";

type AssistantActionPreviewRootState = {
  assistantActions: AssistantActionState;
};

export interface AssistantActionPreviewProps {
  onAccept: () => void;
  onReject: () => void;
  onDismiss: () => void;
}

function actionLabel(kind: AssistantActionState["actionKind"]): string {
  switch (kind) {
    case "generate":
      return "Generate";
    case "rewrite":
      return "Rewrite";
    case "check":
      return "Check";
    default:
      return "Assistant action";
  }
}

export function AssistantActionPreview({
  onAccept,
  onReject,
  onDismiss,
}: AssistantActionPreviewProps) {
  const actionState = useSelector(
    (state: AssistantActionPreviewRootState) => state.assistantActions,
  );
  const label = actionLabel(actionState.actionKind);

  if (actionState.status === "idle") {
    return null;
  }

  if (actionState.status === "running") {
    return (
      <section
        aria-busy="true"
        className="rounded-md border border-border bg-muted/30 p-3 text-sm text-muted-foreground"
      >
        Running {label.toLowerCase()}...
      </section>
    );
  }

  if (actionState.error) {
    return (
      <section className="rounded-md border border-destructive/40 bg-background p-3" role="alert">
        <p className="text-sm text-destructive">{actionState.error}</p>
        <div className="mt-3">
          <Button type="button" variant="outline" size="sm" onClick={onDismiss}>
            Dismiss
          </Button>
        </div>
      </section>
    );
  }

  if (actionState.checkResult) {
    return (
      <section className="rounded-md border border-border bg-background p-4">
        <h3 className="text-sm font-medium text-foreground">Check result</h3>
        <p className="mt-2 whitespace-pre-wrap text-sm leading-6 text-foreground">
          {actionState.checkResult.assistantMessage.bodyMarkdown}
        </p>
        <div className="mt-3">
          <Button type="button" variant="outline" size="sm" onClick={onDismiss}>
            Dismiss check result
          </Button>
        </div>
      </section>
    );
  }

  if (actionState.candidate) {
    return (
      <section className="rounded-md border border-border bg-background p-4">
        <h3 className="text-sm font-medium text-foreground">{label} preview</h3>
        <p className="mt-2 whitespace-pre-wrap text-sm leading-6 text-foreground">
          {actionState.candidate.generatedMarkdown}
        </p>
        {actionState.acceptError ? (
          <p className="mt-3 text-sm text-destructive" role="alert">
            {actionState.acceptError}
          </p>
        ) : null}
        {actionState.rejectError ? (
          <p className="mt-3 text-sm text-destructive" role="alert">
            {actionState.rejectError}
          </p>
        ) : null}
        <div className="mt-4 flex flex-wrap gap-2">
          <Button
            type="button"
            size="sm"
            disabled={actionState.acceptStatus === "running"}
            onClick={onAccept}
          >
            Accept preview
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={actionState.rejectStatus === "running"}
            onClick={onReject}
          >
            Reject preview
          </Button>
        </div>
      </section>
    );
  }

  return null;
}
