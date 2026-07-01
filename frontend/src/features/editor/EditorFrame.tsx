import { ClipboardCheck, FileText, MessageSquarePlus, PenLine, Save } from "lucide-react";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useLocation } from "react-router-dom";
import { GalleyEditor, type GalleyMode } from "@inky/galley-editor";

import type { AppDispatch } from "../../app/store/store";
import { updateContent } from "../../api";
import { Button } from "../../shared/ui/button";
import type { ContentItem } from "../../types";
import type { WorkspaceMode } from "../workspace/workspaceSlice";
import {
  acceptAssistantCandidate,
  assistantActions,
  type AssistantActionState,
  rejectAssistantCandidate,
  runGenerate,
} from "../assistant-actions/assistantActionsSlice";
import { AssistantActionPreview } from "./AssistantActionPreview";
import {
  buildAssistantActionRequestKey,
  buildGenerateRequest,
  validateGenerateAction,
} from "./assistantActionRequests";
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
  assistantActions: AssistantActionState;
  editor: EditorState;
  modelSettings: {
    activeModelVariantId: string | null;
  };
  skills: {
    selectedSkillIds: string[];
  };
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
  const dispatch = useDispatch<AppDispatch>();
  const location = useLocation();
  const assistantActionState = useSelector((state: EditorFrameRootState) => state.assistantActions);
  const editor = useSelector((state: EditorFrameRootState) => state.editor);
  const activeModelVariantId = useSelector(
    (state: EditorFrameRootState) => state.modelSettings.activeModelVariantId,
  );
  const selectedSkillIds = useSelector((state: EditorFrameRootState) => state.skills.selectedSkillIds);
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
      dispatch(assistantActions.clearAssistantActionResult());
      return;
    }
    dispatch(assistantActions.clearAssistantActionResult());
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
    const validationError = validateGenerateAction({
      contentContext: contextMatchesContent ? editor.contentContext : null,
      cursorByte: editor.cursorByte,
      activeModelVariantId,
    });
    if (validationError) {
      setGenerateStatus(validationError);
      return;
    }
    if (!editor.contentContext || editor.cursorByte === null || !activeModelVariantId) return;

    const requestKey = buildAssistantActionRequestKey(location.pathname, editor.contentContext);
    const requestToken = crypto.randomUUID();
    setGenerateStatus(null);
    void dispatch(
      runGenerate(
        buildGenerateRequest({
          contentContext: editor.contentContext,
          activeModelVariantId,
          cursorByte: editor.cursorByte,
          instructions: editor.generateInstructions,
          skillIds: selectedSkillIds,
          requestKey,
          requestToken,
        }),
      ),
    );
  }

  async function handleAcceptPreview(): Promise<void> {
    if (
      !assistantActionState.candidate ||
      !assistantActionState.requestKey ||
      !assistantActionState.requestToken
    ) {
      return;
    }

    const result = await dispatch(
      acceptAssistantCandidate({
        projectId: assistantActionState.candidate.projectId,
        candidateId: assistantActionState.candidate.id,
        requestKey: assistantActionState.requestKey,
        requestToken: assistantActionState.requestToken,
      }),
    );

    if (!acceptAssistantCandidate.fulfilled.match(result)) return;
    dispatch(
      editorActions.saveSucceeded({
        revision: result.payload.content.currentRevision,
        bodyMarkdown: result.payload.content.bodyMarkdown,
      }),
    );
    onContentSaved(result.payload.content);
  }

  function handleRejectPreview(): void {
    if (
      !assistantActionState.candidate ||
      !assistantActionState.requestKey ||
      !assistantActionState.requestToken
    ) {
      return;
    }

    void dispatch(
      rejectAssistantCandidate({
        projectId: assistantActionState.candidate.projectId,
        candidateId: assistantActionState.candidate.id,
        requestKey: assistantActionState.requestKey,
        requestToken: assistantActionState.requestToken,
      }),
    );
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
        <div className="flex flex-wrap items-center gap-2">
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
          minRows={18}
          className="open-edda-galley"
          surface={{
            className: "open-edda-galley-surface",
            contentPadding: "1rem",
            toolbarPadding: "0.5rem 0.75rem",
            footerPadding: "0.5rem 0.75rem",
          }}
        />
      </div>

      {selection ? (
        <SelectionActions className="mobile-selection-toolbar sticky bottom-2 z-10 flex items-center justify-center gap-2 rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-md md:hidden" />
      ) : null}

      <GenerateComposer disabled={generateDisabled} helperText={generateHelperText} onGenerate={handleGenerate} />
      {generateStatus ? <p className="text-sm text-muted-foreground">{generateStatus}</p> : null}
      <AssistantActionPreview
        onAccept={() => {
          void handleAcceptPreview();
        }}
        onReject={handleRejectPreview}
        onDismiss={() => dispatch(assistantActions.clearAssistantActionResult())}
      />
      <SelectionActionDialog />
    </article>
  );
}
