import { editorToolbarIcons } from "./editorToolbarIcons";
import { ClipboardCheck, FileText, MessageSquarePlus, PenLine, Save, SlidersHorizontal, ChevronDown } from "lucide-react";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { GalleyEditor, type GalleyMode } from "@inkyquill/galley-editor";

import { updateContent } from "../../api";
import { Button } from "../../shared/ui/button";
import type { ContentItem } from "../../types";
import type { WorkspaceMode } from "../workspace/workspaceSlice";
import { createGalleyEditorSnapshot, type GalleySelectionSnapshot } from "./editorAdapter";
import { GenerateComposer } from "./GenerateComposer";
import { SelectionActionDialog } from "./SelectionActionDialog";
import { editorActions, type EditorActionKind, type EditorState } from "./editorSlice";

export type EditorFrameProps = {
  projectId: string;
  content: ContentItem | null;
  mode: WorkspaceMode;
  contentLoading?: boolean;
  contentError?: string | null;
  onContentSaved: (content: ContentItem) => void;
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

const workspaceModeToGalleyMode: Record<WorkspaceMode, GalleyMode> = {
  draft: "live",
  assistant: "live",
  review: "preview",
};

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

export function EditorFrame({
  projectId,
  content,
  mode,
  contentLoading = false,
  contentError = null,
  onContentSaved,
}: EditorFrameProps) {
  const dispatch = useDispatch();
  const editor = useSelector((state: EditorFrameRootState) => state.editor);
  const [generateStatus, setGenerateStatus] = useState<string | null>(null);
  const contextMatchesContent =
    editor.contentContext?.projectId === projectId && editor.contentContext.contentId === content?.id;
  const selection = contextMatchesContent ? editor.selection : null;
  const draftMarkdown = contextMatchesContent ? editor.draftMarkdown : (content?.bodyMarkdown ?? "");
  const dirty = contextMatchesContent && editor.dirty;
  const saveStatus = contextMatchesContent ? editor.saveStatus : "idle";
  const generateDisabled = !content || !contextMatchesContent || editor.cursorByte === null;
  const generateHelperText = !content
    ? "Select content before generating."
    : editor.cursorByte === null
      ? "Place the cursor in the draft before generating."
      : `Ready at byte ${editor.cursorByte} on revision ${editor.contentContext?.revision ?? content.currentRevision}.`;

  useEffect(() => {
    if (!content) {
      dispatch(editorActions.resetEditorContext());
      return;
    }
    dispatch(
      editorActions.hydrateEditorContext({
        projectId,
        contentId: content.id,
        contentKind: content.kind,
        revision: content.currentRevision,
        bodyMarkdown: content.bodyMarkdown,
      }),
    );
  }, [content?.bodyMarkdown, content?.currentRevision, content?.id, content?.kind, dispatch, projectId]);

  function handleSelectionChange(galleySelection: GalleySelectionSnapshot) {
    const snapshot = createGalleyEditorSnapshot(editor.draftMarkdown, galleySelection);

    dispatch(editorActions.setCursorByte(snapshot.cursorByte));
    dispatch(editorActions.setSelection(snapshot.selection));
  }

  async function handleSave(): Promise<void> {
    if (!content || !contextMatchesContent || !editor.contentContext || editor.saveStatus === "pending") return;

    dispatch(editorActions.saveStarted());
    try {
      const saved = await updateContent(projectId, content.id, {
        expectedRevision: editor.contentContext.revision,
        bodyMarkdown: editor.draftMarkdown,
        metadataJson: content.metadataJson,
        reason: "editor save",
      });
      dispatch(
        editorActions.saveSucceeded({
          revision: saved.currentRevision,
          bodyMarkdown: saved.bodyMarkdown,
        }),
      );
      onContentSaved(saved);
    } catch (cause: unknown) {
      dispatch(editorActions.saveFailed(cause instanceof Error ? cause.message : "Could not save content"));
    }
  }

  function handleGenerate(): void {
    if (!content || !contextMatchesContent || !editor.contentContext) {
      setGenerateStatus("Select content before generating.");
      return;
    }
    if (editor.cursorByte === null) {
      setGenerateStatus("Place the cursor in the draft before generating.");
      return;
    }
    setGenerateStatus(`Generate is ready for revision ${editor.contentContext.revision} at byte ${editor.cursorByte}.`);
  }

  if (contentLoading) {
    return (
      <article className="workspace-prose-column flex min-h-80 w-full max-w-3xl items-center justify-center bg-background px-6 py-7 text-sm text-muted-foreground">
        Loading content...
      </article>
    );
  }

  if (contentError) {
    return (
      <article
        className="workspace-prose-column flex min-h-80 w-full max-w-3xl items-center justify-center bg-background px-6 py-7 text-sm text-destructive"
        role="alert"
      >
        Could not load content: {contentError}
      </article>
    );
  }

  if (!content) {
    return (
      <article className="workspace-prose-column flex min-h-80 w-full max-w-3xl items-center justify-center bg-background px-6 py-7 text-sm text-muted-foreground">
        Select content to start drafting.
      </article>
    );
  }

  return (
    <article className="workspace-prose-column relative flex w-full max-w-3xl flex-col gap-5 bg-background px-4 py-5 sm:px-6 sm:py-7">
      <header className="chapter-header flex flex-col gap-2 border-b border-border pb-4">
        <div className="hidden flex-wrap items-center gap-2 text-xs text-muted-foreground md:flex">
          <FileText data-icon="inline-start" aria-hidden="true" />
          <span className="capitalize">{formatKind(content.kind)}</span>
          <span aria-hidden="true">/</span>
          <span>Revision {content.currentRevision}</span>
          <span aria-hidden="true">/</span>
          <span className="capitalize">{mode} mode</span>
        </div>
        <h2 className="hidden text-xl font-semibold text-foreground md:block">{content.title}</h2>
        <details className="chapter-details min-w-0 md:hidden">
          <summary className="flex min-h-10 cursor-pointer items-center gap-1 text-sm font-medium">
            <span className="min-w-0 flex-1 truncate">{content.title}</span><ChevronDown className="size-3.5 shrink-0" aria-hidden="true" />
          </summary>
          <p className="pb-2 text-xs text-muted-foreground">{formatKind(content.kind)} · Revision {content.currentRevision} · {mode} mode</p>
        </details>
        <div className="chapter-save flex flex-wrap items-center gap-2">
          <Button
            type="button"
            variant="secondary"
            size="sm"
            disabled={!dirty || saveStatus === "pending"}
            onClick={() => {
              void handleSave();
            }}
          >
            <Save data-icon="inline-start" aria-hidden="true" />
            {saveStatus === "pending" ? "Saving" : "Save"}
          </Button>
          {dirty ? <span className="text-xs text-muted-foreground">Unsaved changes</span> : null}
          {saveStatus === "succeeded" ? <span className="text-xs text-muted-foreground">Saved</span> : null}
          {saveStatus === "failed" ? (
            <span role="alert" className="text-xs text-destructive">
              {editor.saveError}
            </span>
          ) : null}
        </div>
      </header>

      <div className="relative">
        <details className="mobile-formatting md:hidden">
          <summary className="flex min-h-10 cursor-pointer items-center gap-1 text-xs text-muted-foreground"><SlidersHorizontal className="size-3.5" aria-hidden="true" />Formatting</summary>
        </details>
        {selection ? (
          <SelectionActions className="selection-bubble absolute right-2 top-2 z-10 hidden items-center gap-2 rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-md md:flex" />
        ) : null}
        <GalleyEditor
          value={draftMarkdown}
          onChange={(nextValue) => dispatch(editorActions.setDraftMarkdown(nextValue))}
          onSelectionChange={handleSelectionChange}
          placeholder="No draft text yet."
          ariaLabel={`${content.title} draft text`}
          mode={workspaceModeToGalleyMode[mode]}
          theme="inherit"
          toolbar={{ icons: editorToolbarIcons }}
          footer={{ wordCount: true, characterCount: false, logo: false }}
          minRows={18}
          className="open-edda-galley"
          surface={{
            className: "open-edda-galley-surface",
            contentPadding: "1.5rem 0",
            toolbarPadding: "0.5rem 0",
            footerPadding: "0.5rem 0",
          }}
        />
      </div>

      {selection ? (
        <SelectionActions className="mobile-selection-toolbar sticky bottom-2 z-10 flex items-center justify-center gap-2 rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-md md:hidden" />
      ) : null}

      <GenerateComposer disabled={generateDisabled} helperText={generateHelperText} onGenerate={handleGenerate} />
      {generateStatus ? <p className="text-sm text-muted-foreground">{generateStatus}</p> : null}
      <SelectionActionDialog />
    </article>
  );
}
