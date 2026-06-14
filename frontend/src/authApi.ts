export type AuthResponse = {
  token: string;
  author: { id: string; email: string };
};

export function getToken(): string | null {
  return localStorage.getItem("open_edda_token");
}

export function clearToken(): void {
  localStorage.removeItem("open_edda_token");
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const resp = await fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!resp.ok) {
    const err = await resp.json().catch(() => ({ error: `login failed: ${resp.status}` }));
    throw new Error(err.error ?? `login failed: ${resp.status}`);
  }
  const data: AuthResponse = await resp.json();
  localStorage.setItem("open_edda_token", data.token);
  return data;
}

export async function register(email: string, password: string): Promise<AuthResponse> {
  const resp = await fetch("/api/auth/register", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!resp.ok) {
    const err = await resp.json().catch(() => ({ error: `register failed: ${resp.status}` }));
    throw new Error(err.error ?? `register failed: ${resp.status}`);
  }
  const data: AuthResponse = await resp.json();
  localStorage.setItem("open_edda_token", data.token);
  return data;
}
