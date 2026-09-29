import { createContext, useContext, useEffect, useLayoutEffect, useRef, useState, type ReactNode } from "react";
import { useLocation } from "react-router-dom";
import { getToken } from "../../authApi";
import { DEFAULT_THEME_ID, resolveTheme, themeVariables } from "./theme";

type ThemeContextValue = {
  themeId: string;
  loading: boolean;
  ready: boolean;
  saving: boolean;
  error: string | null;
  reload: () => void;
  selectTheme: (id: string) => Promise<void>;
};
const ThemeContext = createContext<ThemeContextValue | null>(null);

async function preferencesRequest(token: string, themeId?: string) {
  const response = await fetch("/api/auth/preferences", {
    method: themeId ? "PUT" : "GET",
    headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
    ...(themeId ? { body: JSON.stringify({ themeId }) } : {}),
  });
  if (!response.ok) throw new Error(themeId ? "Could not save appearance. Please try again." : "Could not load your appearance preferences.");
  return response.json() as Promise<{ themeId: string }>;
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  useLocation(); // Observe login/logout navigation: token changes are stored by authApi.
  const [, refreshAuth] = useState(0);
  useEffect(() => {
    const onStorage = (event: StorageEvent) => {
      if (event.key === "open_edda_token" || event.key === null) refreshAuth((version) => version + 1);
    };
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, []);
  const token = getToken();
  const activeToken = useRef(token);
  activeToken.current = token;
  const [preference, setPreference] = useState<{ token: string | null; id: string }>({ token: null, id: DEFAULT_THEME_ID });
  const [loadedToken, setLoadedToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(Boolean(token));
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [reloadKey, setReloadKey] = useState(0);
  const ready = Boolean(token && loadedToken === token);
  const themeId = preference.token === token ? preference.id : DEFAULT_THEME_ID;

  useLayoutEffect(() => {
    const theme = resolveTheme(themeId);
    const root = document.documentElement;
    for (const [key, value] of Object.entries(themeVariables(themeId))) {
      if (key === "colorScheme") root.style.colorScheme = value;
      else root.style.setProperty(key, value);
    }
    root.classList.toggle("dark", theme.scheme === "dark");
    root.dataset.theme = theme.id;
  }, [themeId]);

  useEffect(() => {
    let cancelled = false;
    setLoadedToken(null);
    setSaving(false);
    setError(null);
    if (!token) {
      setPreference({ token: null, id: DEFAULT_THEME_ID });
      setLoading(false);
      return;
    }
    setLoading(true);
    void preferencesRequest(token).then((data) => {
      if (!cancelled) {
        setPreference({ token, id: resolveTheme(data.themeId).id });
        setLoadedToken(token);
      }
    }).catch((cause: unknown) => {
      if (!cancelled) setError(cause instanceof Error ? cause.message : "Could not load appearance.");
    }).finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [token, reloadKey]);

  async function selectTheme(id: string) {
    if (!token || !ready || loading || saving) return;
    const previous = themeId;
    setPreference({ token, id: resolveTheme(id).id });
    setSaving(true);
    setError(null);
    try {
      const data = await preferencesRequest(token, id);
      if (activeToken.current === token) setPreference({ token, id: resolveTheme(data.themeId).id });
    } catch (cause) {
      if (activeToken.current === token) {
        setPreference({ token, id: previous });
        setError(cause instanceof Error ? cause.message : "Could not save appearance.");
      }
    } finally { if (activeToken.current === token) setSaving(false); }
  }

  return <ThemeContext.Provider value={{ themeId, loading, ready, saving, error, reload: () => setReloadKey((key) => key + 1), selectTheme }}>{children}</ThemeContext.Provider>;
}

export function useTheme() {
  const value = useContext(ThemeContext);
  if (!value) throw new Error("useTheme requires ThemeProvider");
  return value;
}
