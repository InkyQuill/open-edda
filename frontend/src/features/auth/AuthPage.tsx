import { FormEvent, useState } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";

import { getToken, login, register } from "../../authApi";
import { Button } from "../../shared/ui/button";
import { Input } from "../../shared/ui/input";

type AuthMode = "login" | "register";

function safeRedirectPath(value: unknown): string | null {
  if (typeof value !== "string") {
    return null;
  }
  if (!value.startsWith("/") || value.startsWith("//")) {
    return null;
  }
  return value;
}

function redirectTargetFromState(state: unknown): string {
  if (!state || typeof state !== "object" || !("from" in state)) {
    return "/projects";
  }
  return safeRedirectPath((state as { from: unknown }).from) ?? "/projects";
}

export function AuthPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const [mode, setMode] = useState<AuthMode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (getToken()) {
    return <Navigate to="/projects" replace />;
  }

  const isRegistering = mode === "register";

  async function handleSubmit(event: FormEvent<HTMLFormElement>): Promise<void> {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      const authenticate = isRegistering ? register : login;
      await authenticate(email, password);
      navigate(redirectTargetFromState(location.state), { replace: true });
    } catch (cause: unknown) {
      setError(cause instanceof Error ? cause.message : "Authentication failed");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="app-shell">
      <section className="auth-form-section" aria-labelledby="auth-page-title">
        <header>
          <h1 id="auth-page-title">Writer</h1>
          <p>{isRegistering ? "Create your author account." : "Sign in to your writing workspace."}</p>
        </header>

        <form className="auth-form" onSubmit={(event) => void handleSubmit(event)}>
          <label>
            Email
            <Input
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              required
              autoComplete="email"
              disabled={isSubmitting}
            />
          </label>

          <label>
            Password
            <Input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              required
              minLength={8}
              autoComplete={isRegistering ? "new-password" : "current-password"}
              disabled={isSubmitting}
            />
          </label>

          {error ? (
            <p className="auth-error" role="alert">
              {error}
            </p>
          ) : null}

          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Please wait..." : isRegistering ? "Register" : "Login"}
          </Button>

          <Button
            type="button"
            variant="outline"
            disabled={isSubmitting}
            onClick={() => {
              setError(null);
              setMode((current) => (current === "login" ? "register" : "login"));
            }}
          >
            {isRegistering ? "Use existing account" : "Create account"}
          </Button>
        </form>
      </section>
    </main>
  );
}
