export type AuthResponse = {
  token: string;
  refreshExpiresAt?: number;
  author: { id: string; email: string };
};

const sessionListeners = new Set<() => void>();
let sessionExpired = false;

export function subscribeSession(listener: () => void): () => void {
  sessionListeners.add(listener);
  const storageChanged = (event: StorageEvent) => {
    if (event.key === "open_edda_token" || event.key === null) listener();
  };
  if (typeof window !== "undefined") window.addEventListener("storage", storageChanged);
  return () => {
    sessionListeners.delete(listener);
    if (typeof window !== "undefined") window.removeEventListener("storage", storageChanged);
  };
}

function notifySession(): void {
  for (const listener of sessionListeners) listener();
}

export function wasSessionExpired(): boolean { return sessionExpired; }

// A delayed response from an older session must not log out a newer login.
export function expireSession(requestToken: string | null): void {
  if (getToken() !== requestToken) return;
  localStorage.removeItem("open_edda_session");
  localStorage.removeItem("open_edda_token");
  localStorage.removeItem("open_edda_refresh_expires_at");
  sessionExpired = true;
  notifySession();
}

export function getSessionID(): string | null { return localStorage.getItem("open_edda_session"); }

export function getToken(): string | null {
  return localStorage.getItem("open_edda_token");
}

export function clearToken(): void {
  localStorage.removeItem("open_edda_session");
  localStorage.removeItem("open_edda_token");
  localStorage.removeItem("open_edda_refresh_expires_at");
  sessionExpired = false;
  notifySession();
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const operation = () => performLogin(email,password);
  if (typeof navigator !== "undefined" && navigator.locks) return navigator.locks.request("edda-session", operation);
  if (refreshFlight) await refreshFlight;
  return operation();
}

async function performLogin(email: string, password: string): Promise<AuthResponse> {
  const resp = await fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-Edda-Session": "cookie" },
    body: JSON.stringify({ email, password }),
  });
  if (!resp.ok) {
    const err = await resp.json().catch(() => ({ error: `login failed: ${resp.status}` }));
    throw new Error(err.error ?? `login failed: ${resp.status}`);
  }
  const data: AuthResponse = await resp.json();
  if (!data.token) {
    throw new Error("login failed: missing token");
  }
  localStorage.setItem("open_edda_session", crypto.randomUUID());
  saveSession(data);
  return data;
}

function saveSession(data: AuthResponse): void {
  if (typeof data.token !== "string" || !data.token) throw new Error("Сервер не вернул токен сессии.");
  if (typeof data.refreshExpiresAt === "number") localStorage.setItem("open_edda_refresh_expires_at", String(data.refreshExpiresAt));
  else localStorage.removeItem("open_edda_refresh_expires_at");
  localStorage.setItem("open_edda_token", data.token);
  sessionExpired = false;
  notifySession();
}

export function sessionNeedsRefresh(token: string): boolean {
  const now = Date.now() / 1000;
  const refreshExpiry = Number(localStorage.getItem("open_edda_refresh_expires_at"));
  if (refreshExpiry && refreshExpiry - now < 7 * 86400) return true;
  try {
    const claims = JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/'))) as {exp?: unknown};
    return typeof claims.exp === "number" && claims.exp - now < 60;
  } catch { return false; }
}

let refreshFlight: Promise<boolean> | null = null;
export async function refreshSession(requestToken: string | null, sessionID = getSessionID()): Promise<boolean> {
  const alreadyRefreshed = () => !!sessionID && getSessionID() === sessionID && !!getToken();
  if (getToken() !== requestToken) return alreadyRefreshed();
  if (!requestToken) return false;
  if (refreshFlight) return refreshFlight;
  const refresh = async () => {
    if (getToken() !== requestToken) return alreadyRefreshed();
    const response = await fetch("/api/auth/refresh", {method:"POST", headers:{"X-Edda-Session":"cookie"}});
    if (response.status === 401) { expireSession(requestToken); return false; }
    if (!response.ok) throw new Error("Не удалось продлить сессию. Повторите попытку.");
    const data = await response.json() as AuthResponse;
    if (getToken() !== requestToken) return alreadyRefreshed();
    saveSession(data);
    return true;
  };
  // Serialize cookie rotation across tabs as well as concurrent requests.
  refreshFlight = (typeof navigator !== "undefined" && navigator.locks
    ? navigator.locks.request("edda-session", refresh) : refresh());
  try { return await refreshFlight; } finally { refreshFlight = null; }
}

export async function logout(): Promise<void> {
  const operation = async () => {
    const response = await fetch("/api/auth/logout", {method:"POST", headers:{"X-Edda-Session":"cookie"}});
    if (!response.ok) throw new Error("Не удалось завершить сессию. Повторите попытку.");
    clearToken();
  };
  if (typeof navigator !== "undefined" && navigator.locks) await navigator.locks.request("edda-session", operation);
  else { if (refreshFlight) await refreshFlight; await operation(); }
}
