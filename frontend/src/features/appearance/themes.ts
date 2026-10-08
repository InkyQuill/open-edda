import { BUILT_IN_THEMES, getTheme, themeToCssVariables, type ThemeScheme } from '@inkyquill/galley-themes';
export const themeFamilies = [...new Set(BUILT_IN_THEMES.map(theme => theme.family))];
export function resolveTheme(family: string, scheme: ThemeScheme) {
  return BUILT_IN_THEMES.find(theme => theme.family === family && theme.scheme === scheme) ?? getTheme(`edda-${scheme}`)!;
}
export function applyTheme(family: string, scheme: ThemeScheme, root: HTMLElement) {
  const theme = resolveTheme(family, scheme), { app, editor } = theme.tokens;
  const variables = themeToCssVariables(theme);
  for (const [key, value] of Object.entries(variables)) if (key.startsWith('--')) root.style.setProperty(key, value);
  const bridge: Record<string, string> = {
    '--shell': app.bg, '--paper': editor.bg, '--raised': editor.surfaceElevated,
    '--ink': app.text, '--edda-muted': app.textMuted, '--line': app.border,
    '--hover': app.hover, '--selected': editor.selection, '--edda-accent': app.focus,
    '--accent-ink': editor.linkHover, '--danger': app.errorText, '--overlay': app.backdrop,
    '--selection': editor.selection, '--shadow': app.dialogShadow,
  };
  for (const [key, value] of Object.entries(bridge)) root.style.setProperty(key, value);
  root.dataset.theme = scheme; root.dataset.themeId = theme.id;
  root.classList.toggle('dark', scheme === 'dark'); root.style.colorScheme = scheme;
}
