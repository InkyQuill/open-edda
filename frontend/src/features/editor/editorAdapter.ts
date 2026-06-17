import type { ContentKind } from "../../types";
import { textareaSelectionToByteRange, utf8ByteLength } from "../../shared/lib/byteOffsets";
import type { EditorSelection } from "./editorSlice";

export type EditorContentContext = {
  projectId: string;
  contentId: string;
  contentKind: ContentKind;
  revision: number;
};

export type EditorSnapshot = {
  cursorByte: number;
  selection: EditorSelection;
};

export type GalleySelectionSnapshot = {
  from: number;
  to: number;
  anchor: number;
  head: number;
};

export type HydrateEditorContextInput = EditorContentContext & {
  bodyMarkdown: string;
};

export function createTextareaEditorSnapshot(
  value: string,
  selectionStart: number,
  selectionEnd: number,
): EditorSnapshot {
  return {
    cursorByte: utf8ByteLength(value.slice(0, selectionEnd)),
    selection: textareaSelectionToByteRange(value, selectionStart, selectionEnd),
  };
}

export function createGalleyEditorSnapshot(
  value: string,
  selection: GalleySelectionSnapshot,
): EditorSnapshot {
  return {
    cursorByte: utf8ByteLength(value.slice(0, selection.head)),
    selection: textareaSelectionToByteRange(value, selection.from, selection.to),
  };
}
