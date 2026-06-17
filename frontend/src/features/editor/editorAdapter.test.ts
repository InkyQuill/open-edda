import { describe, expect, it } from "vitest";

import { createGalleyEditorSnapshot, createTextareaEditorSnapshot } from "./editorAdapter";

describe("editorAdapter", () => {
  it("converts a cursor position to a UTF-8 byte offset", () => {
    const snapshot = createTextareaEditorSnapshot("A Ж🙂 end", 5, 5);

    expect(snapshot.cursorByte).toBe(8);
    expect(snapshot.selection).toBeNull();
  });

  it("converts a selected DOM range to UTF-8 byte offsets and preview text", () => {
    const value = "A Ж🙂 end";
    const snapshot = createTextareaEditorSnapshot(value, 2, 5);

    expect(snapshot.cursorByte).toBe(8);
    expect(snapshot.selection).toEqual({
      startByte: 2,
      endByte: 8,
      previewText: "Ж🙂",
    });
  });

  it("uses Galley head position for cursor byte while keeping normalized selection bytes", () => {
    const value = "A Ж🙂 end";
    const snapshot = createGalleyEditorSnapshot(value, {
      from: 2,
      to: 5,
      anchor: 5,
      head: 2,
    });

    expect(snapshot.cursorByte).toBe(2);
    expect(snapshot.selection).toEqual({
      startByte: 2,
      endByte: 8,
      previewText: "Ж🙂",
    });
  });
});
