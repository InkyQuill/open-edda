import { ClipboardCheck, FileText, MessageSquarePlus, PenLine } from "lucide-react";
import { useEffect, useState, type SyntheticEvent } from "react";
import { useDispatch, useSelector } from "react-redux";

import { textareaSelectionToByteRange, utf8ByteLength } from "../../shared/lib/byteOffsets";
import { Button } from "../../shared/ui/button";
import { Textarea } from "../../shared/ui/textarea";
import type { ContentItem } from "../../types";
import type { WorkspaceMode } from "../workspace/workspaceSlice";
import { GenerateComposer } from "./GenerateComposer";
import { SelectionActionDialog } from "./SelectionActionDialog";
import { editorActions, type EditorActionKind, type EditorState } from "./editorSlice";

export type EditorFrameProps = {
  content: ContentItem | null;
  mode: WorkspaceMode;
  contentLoading?: boolean;
  contentError?: string | null;
};

type EditorFrameRootState = {
  editor: EditorState;
};

const actionButtons: Array<{
  kind: EditorActionKind;
  label: string;
  icon: typeof PenLine;
}> = [
  { kind: "rewrite", label: "Rewrite", icon: PenLine },
  { kind: "check", label: "Check", icon: ClipboardCheck },
  { kind: "note", label: "Note", icon: MessageSquarePlus },
];

function formatKind(kind: ContentItem["kind"]): string {
  return kind.replaceAll("_", " ");
}

function SelectionActions({ className }: { className: string }) {
  const dispatch = useDispatch();

  return (
    <div className={className}>
      {actionButtons.map(({ kind, label, icon: Icon }) => (
        <Button
          key={kind}
          type="button"
          variant="secondary"
          size="sm"
          onClick={() => dispatch(editorActions.openActionModal({ kind }))}
        >
          <Icon data-icon="inline-start" aria-hidden="true" />
          {label}
        </Button>
      ))}
    </div>
  );
}

export function EditorFrame({ content, mode, contentLoading = false, contentError = null }: EditorFrameProps) {
  const dispatch = useDispatch();
  const selection = useSelector((state: EditorFrameRootState) => state.editor.selection);
  const [generateStatus, setGenerateStatus] = useState<string | null>(null);
  const body = content?.bodyMarkdown ?? "";

  useEffect(() => {
    dispatch(editorActions.resetEditorContext());
  }, [content?.id, content?.currentRevision, body, dispatch]);

  function handleSelect(event: SyntheticEvent<HTMLTextAreaElement>) {
    const textarea = event.currentTarget;
    const { selectionStart, selectionEnd, value } = textarea;
    if (value.length === 0) {
      dispatch(editorActions.setCursorByte(null));
      dispatch(editorActions.setSelection(null));
      return;
    }
    const cursorByte = utf8ByteLength(value.slice(0, selectionStart));
    const byteRange = textareaSelectionToByteRange(value, selectionStart, selectionEnd);

    dispatch(editorActions.setCursorByte(cursorByte));
    dispatch(editorActions.setSelection(byteRange));
  }

  if (contentLoading) {
    return (
      <article className="workspace-prose-column flex min-h-80 w-full max-w-3xl items-center justify-center rounded-md border border-border bg-background px-6 py-7 text-sm text-muted-foreground shadow-sm">
        Loading content...
      </article>
    );
  }

  if (contentError) {
    return (
      <article
        className="workspace-prose-column flex min-h-80 w-full max-w-3xl items-center justify-center rounded-md border border-border bg-background px-6 py-7 text-sm text-destructive shadow-sm"
        role="alert"
      >
        Could not load content: {contentError}
      </article>
    );
  }

  if (!content) {
    return (
      <article className="workspace-prose-column flex min-h-80 w-full max-w-3xl items-center justify-center rounded-md border border-border bg-background px-6 py-7 text-sm text-muted-foreground shadow-sm">
        Select content to start drafting.
      </article>
    );
  }

  return (
    <article className="workspace-prose-column relative flex w-full max-w-3xl flex-col gap-5 rounded-md border border-border bg-background px-4 py-5 shadow-sm sm:px-6 sm:py-7">
      <header className="flex flex-col gap-2 border-b border-border pb-4">
        <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
          <FileText data-icon="inline-start" aria-hidden="true" />
          <span className="capitalize">{formatKind(content.kind)}</span>
          <span aria-hidden="true">/</span>
          <span>Revision {content.currentRevision}</span>
          <span aria-hidden="true">/</span>
          <span className="capitalize">{mode} mode</span>
        </div>
        <h2 className="text-2xl font-semibold text-foreground">{content.title}</h2>
      </header>

      <div className="relative">
        {selection ? (
          <SelectionActions className="selection-bubble absolute right-2 top-2 z-10 hidden items-center gap-2 rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-md md:flex" />
        ) : null}
        <Textarea
          readOnly
          value={body}
          onSelect={handleSelect}
          placeholder="No draft text yet."
          className="min-h-[48dvh] resize-none bg-background text-sm leading-7 text-foreground"
          aria-label={`${content.title} draft text`}
        />
      </div>

      {selection ? (
        <SelectionActions className="mobile-selection-toolbar sticky bottom-2 z-10 flex items-center justify-center gap-2 rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-md md:hidden" />
      ) : null}

      <GenerateComposer disabled={!content} onGenerate={() => setGenerateStatus("Generate wiring is next")} />
      {generateStatus ? <p className="text-sm text-muted-foreground">{generateStatus}</p> : null}
      <SelectionActionDialog />
    </article>
  );
}
