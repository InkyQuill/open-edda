export function diffLines(current: string, selected: string): Array<{ kind: "same" | "added" | "removed"; text: string }> {
  const currentLines = current.split("\n");
  const selectedLines = selected.split("\n");
  // Bound the LCS matrix; a whole changed region is preferable to freezing the editor.
  if ((currentLines.length + 1) * (selectedLines.length + 1) > 250_000) {
    let prefix = 0, suffix = 0;
    while (prefix < currentLines.length && prefix < selectedLines.length && currentLines[prefix] === selectedLines[prefix]) prefix++;
    while (suffix < currentLines.length - prefix && suffix < selectedLines.length - prefix && currentLines[currentLines.length - 1 - suffix] === selectedLines[selectedLines.length - 1 - suffix]) suffix++;
    return [
      ...currentLines.slice(0, prefix).map(text => ({kind: "same" as const, text})),
      ...currentLines.slice(prefix, currentLines.length - suffix).map(text => ({kind: "removed" as const, text})),
      ...selectedLines.slice(prefix, selectedLines.length - suffix).map(text => ({kind: "added" as const, text})),
      ...currentLines.slice(currentLines.length - suffix).map(text => ({kind: "same" as const, text})),
    ];
  }
  const lengths = Array.from({ length: currentLines.length + 1 }, () => Array<number>(selectedLines.length + 1).fill(0));

  for (let currentIndex = currentLines.length - 1; currentIndex >= 0; currentIndex -= 1) {
    for (let selectedIndex = selectedLines.length - 1; selectedIndex >= 0; selectedIndex -= 1) {
      lengths[currentIndex][selectedIndex] =
        currentLines[currentIndex] === selectedLines[selectedIndex]
          ? lengths[currentIndex + 1][selectedIndex + 1] + 1
          : Math.max(lengths[currentIndex + 1][selectedIndex], lengths[currentIndex][selectedIndex + 1]);
    }
  }

  const lines: Array<{ kind: "same" | "added" | "removed"; text: string }> = [];
  let currentIndex = 0;
  let selectedIndex = 0;

  while (currentIndex < currentLines.length && selectedIndex < selectedLines.length) {
    if (currentLines[currentIndex] === selectedLines[selectedIndex]) {
      lines.push({ kind: "same", text: currentLines[currentIndex] });
      currentIndex += 1;
      selectedIndex += 1;
      continue;
    }

    if (lengths[currentIndex + 1][selectedIndex] >= lengths[currentIndex][selectedIndex + 1]) {
      lines.push({ kind: "removed", text: currentLines[currentIndex] });
      currentIndex += 1;
    } else {
      lines.push({ kind: "added", text: selectedLines[selectedIndex] });
      selectedIndex += 1;
    }
  }
  while (currentIndex < currentLines.length) {
    lines.push({ kind: "removed", text: currentLines[currentIndex] });
    currentIndex += 1;
  }
  while (selectedIndex < selectedLines.length) {
    lines.push({ kind: "added", text: selectedLines[selectedIndex] });
    selectedIndex += 1;
  }

  return lines;
}
