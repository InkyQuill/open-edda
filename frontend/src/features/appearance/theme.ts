import { getTheme, isThemeId, themeToCssVariables, type ThemeCssVariables } from "@inkyquill/galley-themes";

export const DEFAULT_THEME_ID = "thoth-light";

export function resolveTheme(value: unknown) {
  return getTheme(isThemeId(value) ? value : DEFAULT_THEME_ID)!;
}

export function themeVariables(id: unknown): ThemeCssVariables {
  const theme = resolveTheme(id);
  const { app, editor } = theme.tokens;
  return {
    ...themeToCssVariables(theme),
    "--background": app.bg,
    "--foreground": app.text,
    "--card": app.panel,
    "--card-foreground": app.text,
    "--popover": app.panel,
    "--popover-foreground": app.text,
    "--primary": app.focus,
    "--primary-foreground": app.panel,
    "--secondary": app.hover,
    "--secondary-foreground": app.text,
    "--muted": app.panelMuted,
    "--muted-foreground": app.textMuted,
    "--accent": app.hover,
    "--accent-foreground": app.text,
    "--destructive": app.errorText,
    "--border": app.border,
    "--input": app.border,
    "--ring": app.focus,
    "--sidebar": app.panelMuted,
    "--sidebar-foreground": app.text,
    "--sidebar-primary": app.focus,
    "--sidebar-primary-foreground": app.panel,
    "--sidebar-accent": app.hover,
    "--sidebar-accent-foreground": app.text,
    "--sidebar-border": app.border,
    "--sidebar-ring": app.focus,
    "--app-background": app.bg,
    "--app-surface": app.panel,
    "--app-surface-muted": app.panelMuted,
    "--app-accent": app.focus,
    "--selection-background": editor.selection,
  };
}
