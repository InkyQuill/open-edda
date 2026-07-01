import type {
  RunGenerateArgs,
  RunRewriteArgs,
} from "../assistant-actions/assistantActionsSlice";
import type { EditorContentContext } from "./editorSlice";

export interface GenerateValidationInput {
  contentContext: EditorContentContext | null;
  cursorByte: number | null;
  activeModelVariantId: string | null;
}

export function validateGenerateAction({
  contentContext,
  cursorByte,
  activeModelVariantId,
}: GenerateValidationInput): string | null {
  if (!contentContext) {
    return "Open a chapter or story-bible entry before generating.";
  }
  if (!activeModelVariantId) {
    return "Choose an active model variant in Settings before generating.";
  }
  if (cursorByte === null) {
    return "Place the cursor where generated text should be inserted.";
  }
  return null;
}

export function buildAssistantActionRequestKey(
  routePathname: string,
  contentContext: EditorContentContext,
): string {
  return `${routePathname}:${contentContext.projectId}:${contentContext.contentId}:${contentContext.revision}`;
}

export interface BuildGenerateRequestInput {
  contentContext: EditorContentContext;
  activeModelVariantId: string;
  cursorByte: number;
  instructions: string;
  skillIds: string[];
  requestKey: string;
  requestToken: string;
}

export function buildGenerateRequest({
  contentContext,
  activeModelVariantId,
  cursorByte,
  instructions,
  skillIds,
  requestKey,
  requestToken,
}: BuildGenerateRequestInput): RunGenerateArgs {
  return {
    projectId: contentContext.projectId,
    contentId: contentContext.contentId,
    modelVariantId: activeModelVariantId,
    expectedRevision: contentContext.revision,
    cursorByte,
    instructions: instructions.trim(),
    skillIds,
    requestKey,
    requestToken,
  };
}

export interface SelectionValidationInput {
  contentContext: EditorContentContext | null;
  selection: { startByte: number; endByte: number } | null;
  activeModelVariantId: string | null;
}

export function validateSelectionAction({
  contentContext,
  selection,
  activeModelVariantId,
}: SelectionValidationInput): string | null {
  if (!contentContext) {
    return "Open a chapter or story-bible entry before running this action.";
  }
  if (!activeModelVariantId) {
    return "Choose an active model variant in Settings before running this action.";
  }
  if (!selection || selection.startByte === selection.endByte) {
    return "Select text before running this action.";
  }
  return null;
}

export interface BuildSelectionActionRequestInput {
  contentContext: EditorContentContext;
  activeModelVariantId: string;
  selection: { startByte: number; endByte: number };
  instructions: string;
  skillIds: string[];
  requestKey: string;
  requestToken: string;
}

export function buildRewriteRequest({
  contentContext,
  activeModelVariantId,
  selection,
  instructions,
  skillIds,
  requestKey,
  requestToken,
}: BuildSelectionActionRequestInput): RunRewriteArgs {
  return {
    projectId: contentContext.projectId,
    contentId: contentContext.contentId,
    modelVariantId: activeModelVariantId,
    expectedRevision: contentContext.revision,
    selectionStartByte: selection.startByte,
    selectionEndByte: selection.endByte,
    instructions: instructions.trim(),
    skillIds,
    requestKey,
    requestToken,
  };
}

export const buildCheckRequest = buildRewriteRequest;
