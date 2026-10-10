export const editorText = (source: string) => source.replace(/\r\n?/g, '\n');

// CodeMirror and HTML textareas normalize newlines. Keep untouched source bytes,
// including mixed line endings, when mapping an actual author edit back to disk.
export function sourceText(source: string, next: string): string {
  const previous = editorText(source);
  if (previous === next) return source;
  let start = 0, end = previous.length, nextEnd = next.length;
  while (start < end && start < nextEnd && previous[start] === next[start]) start++;
  while (end > start && nextEnd > start && previous[end - 1] === next[nextEnd - 1]) { end--; nextEnd--; }
  let from = 0, to = source.length, position = 0;
  for (let i = 0; i <= source.length; i++) {
    if (position === start) from = i;
    if (position === end) { to = i; break; }
    if (source[i] === '\r' && source[i + 1] === '\n') i++;
    position++;
  }
  const newline = source.match(/\r\n|\r|\n/)?.[0] ?? '\n';
  return source.slice(0, from) + next.slice(start, nextEnd).replace(/\n/g, newline) + source.slice(to);
}
