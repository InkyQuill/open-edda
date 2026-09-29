import { BUILT_IN_THEMES } from "@inkyquill/galley-themes";
import { Check, Moon, Sun } from "lucide-react";
import { Button } from "../../shared/ui/button";
import { useTheme } from "./ThemeProvider";

export function AppearancePanel() {
  const { themeId, loading, ready, saving, error, reload, selectTheme } = useTheme();
  const themes = [...BUILT_IN_THEMES].sort((a, b) => Number(b.family === "Thoth") - Number(a.family === "Thoth"));
  return (
    <section className="flex flex-col gap-5" aria-labelledby="appearance-title" aria-busy={loading || saving}>
      <header>
        <h2 id="appearance-title" className="text-base font-medium">Appearance</h2>
        <p className="mt-1 text-sm text-muted-foreground">A palette for your entire writing space. Saved to your author profile across devices.</p>
      </header>
      <p role="status" className="text-sm text-muted-foreground">{loading ? "Loading your preference…" : saving ? "Saving appearance…" : "Thoth is Edda’s default palette."}</p>
      {error ? <div role="alert" className="flex flex-wrap items-center gap-3 text-sm text-destructive">{error}<Button variant="outline" size="sm" disabled={saving || loading} onClick={reload}>Retry</Button></div> : null}
      <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3" role="group" aria-label="Color theme">
        {themes.map((theme) => {
          const selected = theme.id === themeId;
          const Icon = theme.scheme === "light" ? Sun : Moon;
          return (
            <button key={theme.id} type="button" aria-pressed={selected} disabled={!ready || loading || saving} onClick={() => void selectTheme(theme.id)} className="theme-choice flex items-center gap-3 rounded-md border px-3 py-3 text-left transition-colors disabled:opacity-60" data-selected={selected}>
              <span className="flex h-10 w-14 shrink-0 items-center justify-center gap-1 rounded border" style={{ background: theme.tokens.app.bg, borderColor: theme.tokens.app.border }} aria-hidden="true">
                {[theme.tokens.app.text, theme.tokens.app.focus, theme.tokens.app.panelMuted].map((color, index) => <span key={index} className="size-2.5 rounded-full" style={{ background: color }} />)}
              </span>
              <span className="min-w-0 flex-1"><span className="block text-sm font-medium">{theme.label}</span><span className="flex items-center gap-1 text-xs text-muted-foreground"><Icon className="size-3" aria-hidden="true" />{theme.scheme === "light" ? "Light" : "Dark"}</span></span>
              {selected ? <Check className="size-4 text-primary" aria-hidden="true" /> : null}
            </button>
          );
        })}
      </div>
    </section>
  );
}
