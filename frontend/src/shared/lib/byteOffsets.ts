const encoder = new TextEncoder();

export function utf8ByteLength(value: string): number {
  return encoder.encode(value).length;
}

export function textareaSelectionToByteRange(
  value: string,
  selectionStart: number,
  selectionEnd: number,
): { startByte: number; endByte: number; previewText: string } | null {
  if (selectionStart === selectionEnd) {
    return null;
  }
  return {
    startByte: utf8ByteLength(value.slice(0, selectionStart)),
    endByte: utf8ByteLength(value.slice(0, selectionEnd)),
    previewText: value.slice(selectionStart, selectionEnd),
  };
}
